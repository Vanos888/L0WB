package service

import (
	"L0WB/internal/domain"
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"
)

type IOrderCache interface {
	Set(orderUID uuid.UUID, order *domain.Order)
	Get(orderUID uuid.UUID) (*domain.Order, bool)
}

type IRepository interface {
	GetOrder(ctx context.Context, orderUID uuid.UUID) (domain.Order, error)
	GetAllOrdersByUID(ctx context.Context) ([]uuid.UUID, error)
	SaveOrder(ctx context.Context, order *domain.Order) error
}

type OrderGenerator interface {
	GenerateFakeOrders(count int) []*domain.CompleteFakeOrder
}

type OrderSender interface {
	SendOrder(ctx context.Context, order *domain.CompleteFakeOrder) error
}

type Service struct {
	repo      IRepository
	cache     IOrderCache
	generator OrderGenerator
	sender    OrderSender
}

func NewService(repo IRepository, cache IOrderCache, generator OrderGenerator, sender OrderSender) *Service {
	return &Service{
		repo:      repo,
		cache:     cache,
		generator: generator,
		sender:    sender,
	}
}

func (s *Service) GetOrder(ctx context.Context, orderUID uuid.UUID) (*domain.Order, error) {
	//Пробуем получить данные заказа из кэша
	if cacheOrder, exist := s.cache.Get(orderUID); exist {
		fmt.Println("Ордер получен из кэша")
		log.Printf("Cache hit for order: %s:", orderUID)
		return cacheOrder, nil
	}

	log.Printf("Cache miss for order: %s:", orderUID)

	//Если нет данных в кеше - Получаем из БД
	order, err := s.repo.GetOrder(ctx, orderUID)
	fmt.Println("Ордер получен из БД")
	if err != nil {
		log.Println("Order not found for DB", err)
		return nil, fmt.Errorf("GetOrder: %w", err)
	}

	//Сохраняем в кеш
	s.cache.Set(orderUID, &order)

	return &order, nil
}

func (s *Service) WarmUpCache(ctx context.Context) error {
	log.Println("Warming up cache...")

	//Получаю все ID из БД
	orderUIDs, err := s.repo.GetAllOrdersByUID(ctx)
	if err != nil {
		return err
	}

	log.Printf("Found orders by uid: %d", len(orderUIDs))

	//Добавления ордеров в кеш

	for i, orderUID := range orderUIDs {
		order, err := s.repo.GetOrder(ctx, orderUID)
		if err != nil {
			log.Printf("Error loading order %s: %v", orderUID, err)
			continue
		}

		s.cache.Set(orderUID, &order)

		if (i+1)%100 == 0 {
			log.Printf("Warmed up %d orders...", i+1)
		}
	}
	log.Printf("Cache warmed up %d orders...", len(orderUIDs))
	return nil
}

func (s *Service) SaveOrderFromKafka(ctx context.Context, order *domain.Order) error {
	if err := s.repo.SaveOrder(ctx, order); err != nil {
		return err
	}

	log.Printf("Saved order to kafka: %s", order)
	return nil
}

func (s *Service) GenerateFakeOrdersFromKafka(ctx context.Context, count int) error {
	orders := s.generator.GenerateFakeOrders(count)

	for _, order := range orders {
		if err := s.sender.SendOrder(ctx, order); err != nil {
			return err
		}
		log.Printf("Generated fake order: %s", order.OrderUID)
		time.Sleep(2 * time.Second)
	}
	return nil
}
