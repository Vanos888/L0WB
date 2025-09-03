package kafka

import (
	"L0WB/internal/domain"
	"L0WB/internal/service"
	"context"
	"github.com/google/uuid"
	"github.com/ogen-go/ogen/json"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

type OrderConsumer struct {
	reader  *kafka.Reader
	service *service.Service
	topic   string
}

func NewOrderConsumer(brokers []string, topic string, groupID string, service *service.Service) *OrderConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		MaxWait:        1 * time.Second,
		CommitInterval: time.Second,
	})

	return &OrderConsumer{
		reader:  reader,
		service: service,
		topic:   topic,
	}
}

func (c *OrderConsumer) Consume(ctx context.Context) {
	log.Printf("Starting Kafka consumer for topic: %s", c.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer")
			return

		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}

			log.Printf("Received message: offset=%d", msg.Offset)
			c.processMessage(ctx, msg)
		}
	}
}
func (c *OrderConsumer) processMessage(ctx context.Context, msg kafka.Message) {
	log.Printf("Received message: %s", string(msg.Value))

	// ВРЕМЕННО: распарсим как CompleteFakeOrder чтобы посмотреть структуру
	var fakeOrder domain.CompleteFakeOrder
	if err := json.Unmarshal(msg.Value, &fakeOrder); err != nil {
		log.Printf("Error unmarshaling as CompleteFakeOrder: %v", err)
		return
	}
	log.Printf("CompleteFakeOrder: %+v", fakeOrder)

	// Преобразуем в domain.Order
	order := convertFakeToDomainOrder(fakeOrder)

	if err := c.service.SaveOrderFromKafka(ctx, order); err != nil {
		log.Printf("Error processing order: %v", err)
		return
	}

	log.Printf("Order processed successfully: %s", order.ID)
}

// Добавьте эту функцию преобразования
func convertFakeToDomainOrder(fake domain.CompleteFakeOrder) *domain.Order {
	// Преобразуем items
	var items []domain.Item
	for _, fakeItem := range fake.Items {
		items = append(items, domain.Item{
			ChartID:     fakeItem.ChrtID,
			TrackNumber: fakeItem.TrackNumber,
			Price:       fakeItem.Price,
			RID:         fakeItem.Rid,
			Name:        fakeItem.Name,
			Sale:        fakeItem.Sale,
			Size:        fakeItem.Size,
			TotalPrice:  fakeItem.TotalPrice,
			NmID:        fakeItem.NmID,
			Brand:       fakeItem.Brand,
			Status:      fakeItem.Status,
		})
	}

	// Создаем UUID из строки
	orderUID, _ := uuid.Parse(fake.OrderUID)
	dateCreated, _ := time.Parse(time.RFC3339, fake.DateCreated)

	return &domain.Order{
		ID:                orderUID,
		TrackNumber:       fake.TrackNumber,
		Entry:             fake.Entry,
		Locale:            fake.Locale,
		InternalSignature: fake.InternalSignature,
		CustumerID:        fake.CustomerID,
		DeliveryService:   fake.DeliveryService,
		ShardKey:          fake.ShardKey,
		SmID:              fake.SmID,
		DateCreated:       dateCreated,
		OofShard:          fake.OofShard,
		Delivery: domain.Delivery{
			Name:    fake.Delivery.Name,
			Phone:   fake.Delivery.Phone,
			Zip:     fake.Delivery.Zip,
			City:    fake.Delivery.City,
			Address: fake.Delivery.Address,
			Region:  fake.Delivery.Region,
			Email:   fake.Delivery.Email,
		},
		Payment: domain.Payment{
			Transaction:  fake.Payment.Transaction,
			RequestID:    fake.Payment.RequestID,
			Currency:     fake.Payment.Currency,
			Provider:     fake.Payment.Provider,
			Amount:       fake.Payment.Amount,
			PaymentDt:    fake.Payment.PaymentDt,
			Bank:         fake.Payment.Bank,
			DeliveryCost: fake.Payment.DeliveryCost,
			GoodsTotal:   fake.Payment.GoodsTotal,
			CustomFee:    fake.Payment.CustomFee,
		},
		Items: items,
	}
}

func (c *OrderConsumer) Close() error {
	return c.reader.Close()
}
