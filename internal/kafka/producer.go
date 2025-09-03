package kafka

import (
	"L0WB/internal/domain"
	"context"
	"github.com/ogen-go/ogen/json"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

type OrderProducer struct {
	writer *kafka.Writer
	topic  string
}

func NewOrderProducer(brokers []string, topic string) *OrderProducer {
	return &OrderProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
		topic: topic,
	}
}

func (p *OrderProducer) SendOrder(ctx context.Context, order *domain.CompleteFakeOrder) error {
	jsonData, err := json.Marshal(order)
	if err != nil {
		return err
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(time.Now().Format(time.RFC3339)),
		Value: jsonData,
	})
	if err != nil {
		return err
	}

	log.Printf("Producer sent order to kafka: %s", string(jsonData))
	return nil
}

func (p *OrderProducer) Close() error {
	return p.writer.Close()
}
