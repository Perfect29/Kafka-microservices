package order

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

type Producer interface {
	PublishOrderCreated(ctx context.Context, order *Order) error
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

func (p *KafkaProducer) PublishOrderCreated(ctx context.Context, order *Order) error {
	value, err := json.Marshal(order)
	if err != nil {
		log.Errorf("Could not Marshal the struct")
		return err
	}

	msg := kafka.Message{
		Key: []byte("order"),
		Value: value,
	}

	return p.writer.WriteMessages(ctx, msg)
}