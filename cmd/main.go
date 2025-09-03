package main

import (
	ogen_server "L0WB/internal/generated/servers/http/ordergen"
	handler "L0WB/internal/handler/http"
	"L0WB/internal/kafka"
	"L0WB/internal/repository/order"
	"L0WB/internal/service"
	"L0WB/internal/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	// Проверяем наличие web директории
	if _, err := os.Stat("./web"); os.IsNotExist(err) {
		log.Fatal("Web directory not found! Create web/ folder with index.html, style.css, script.js")
	}

	// Проверяем наличие необходимых файлов
	requiredFiles := []string{"index.html", "style.css", "script.js"}
	for _, file := range requiredFiles {
		if _, err := os.Stat(filepath.Join("./web", file)); os.IsNotExist(err) {
			log.Printf("Warning: web/%s not found", file)
		}
	}

	ctx := context.Background()

	// Инициализирую БД
	conn, err := pgxpool.New(context.Background(), "postgres://Ivan:1q2w3e4r@127.0.0.1:5432/orders?sslmode=disable&pool_max_conns=10&pool_max_conn_lifetime=1h30m")
	if err != nil {
		log.Fatal("Database connection error: ", err)
	}
	defer conn.Close()

	// Инициализирую Producer
	kafkaBrokers := []string{"localhost:9092"}
	kafkaTopic := "orders"
	kafkaProducer := kafka.NewOrderProducer(kafkaBrokers, kafkaTopic)
	defer kafkaProducer.Close()

	// Создаем генератор
	orderGenerator := &kafka.OrderGeneratorImpl{}

	// Инициализация кеша с TTL 1 час
	orderCache := storage.NewOrderCache(1 * time.Hour)

	// Инициализирую Репозиторий и Сервис
	repository := order.NewRepository(conn)
	orderService := service.NewService(repository, orderCache, orderGenerator, kafkaProducer)

	// Прогрев кеша
	warmupCtx, warmupCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer warmupCancel()

	if err := orderService.WarmUpCache(warmupCtx); err != nil {
		log.Printf("WarmUpCache failed: %v", err)
	} else {
		log.Printf("WarmUpCache size: %d", orderCache.Size())
	}

	api := handler.NewHandler(orderService)

	srv, err := ogen_server.NewServer(api)
	if err != nil {
		log.Fatal("Server creation error: ", err)
	}

	mux := http.NewServeMux()

	webDir := "./web"
	fileServer := http.FileServer(http.Dir(webDir))

	// Обработка статических файлов
	mux.Handle("/web/", http.StripPrefix("/web/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		}
		fileServer.ServeHTTP(w, r)
	})))

	// API endpoint для генерации заказов
	mux.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		countStr := r.URL.Query().Get("count")
		count, err := strconv.Atoi(countStr)
		if err != nil || count <= 0 {
			count = 1
		}

		log.Printf("Generating %d test orders", count)

		orders := kafka.GenerateFakeOrders(count)
		generatedCount := 0

		for _, order := range orders {
			if order != nil {
				err := kafkaProducer.SendOrder(r.Context(), order)
				if err != nil {
					log.Printf("Error sending order: %v", err)
				} else {
					generatedCount++
					log.Printf("Generated order: %s", order.OrderUID)
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"success": true, "generated": %d}`, generatedCount)))
	})

	// API endpoint для получения заказа по ID
	mux.HandleFunc("/order/get-order/", func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Извлекаем UUID из URL: /order/get-order/{uuid}
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 4 {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		orderUID := pathParts[3]
		log.Printf("GET order request: %s", orderUID)

		id, err := uuid.Parse(orderUID)
		if err != nil {
			http.Error(w, "Invalid order UUID", http.StatusBadRequest)
			return
		}

		order, err := orderService.GetOrder(r.Context(), id)
		if err != nil {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    order,
			"cached":  false,
		})
	})

	mux.Handle("/order/", srv)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Если запрос к статическим файлам
		if strings.HasPrefix(r.URL.Path, "/web/") {
			fileServer.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/order/get-order/") ||
			strings.HasPrefix(r.URL.Path, "/generate") ||
			strings.HasPrefix(r.URL.Path, "/order/order/") {
			mux.ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, filepath.Join(webDir, "index.html"))
	})

	// Запускаю обработчик Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	server := http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	var wg sync.WaitGroup

	// Инициализирую Consumer
	kafkaConsumer := kafka.NewOrderConsumer(
		kafkaBrokers,
		kafkaTopic,
		"order-service-group",
		orderService,
	)
	defer kafkaConsumer.Close()

	// Запуск Consumer в горутине
	wg.Add(1)
	go func() {
		defer wg.Done()
		kafkaConsumer.Consume(ctx)
	}()

	// Запускаю сервер в горутине
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Server starting on :8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Авто-генерация тестового заказа через 3 секунды
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(3 * time.Second)
		log.Println("Auto-generating test order...")
		orders := kafka.GenerateFakeOrders(1)
		for _, order := range orders {
			if order != nil {
				err := kafkaProducer.SendOrder(context.Background(), order)
				if err != nil {
					log.Printf("Error auto-generating order: %v", err)
				} else {
					log.Printf("Auto-generated order: %s", order.OrderUID)
				}
			}
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Останавливаю HTTP сервер
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Ждем завершения всех горутин
	wg.Wait()

	log.Println("Server stopped gracefully")
}
