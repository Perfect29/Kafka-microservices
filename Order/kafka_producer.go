package order

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

type Producer interface {
	PublishOrderCreated(ctx context.Context, order_event *OrderEvent) error
}


func NewKafkaProducer(brokers []string, topic string) *KafkaProducer{
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr: kafka.TCP(brokers...),
			Topic: topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaProducer) PublishOrderCreated(ctx context.Context, order_event *OrderEvent) error {
	value, err := json.Marshal(order_event)
	if err != nil {
		log.Errorf("Could not Marshal the struct")
		return err
	}

	msg := kafka.Message{
		Key: []byte(fmt.Sprintf("order-%d", order_event.Order.ID)),
		Value: value,
	}

	return p.writer.WriteMessages(ctx, msg)
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}