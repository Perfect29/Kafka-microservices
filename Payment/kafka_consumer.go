package payment

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

type KafkaConsumer struct {
	reader *kafka.Reader
}


func NewKafkaComsumer(brokers []string, topic string, groupID string) *KafkaConsumer{
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic: topic,
			GroupID: groupID,
		}),
	}
}

func (c *KafkaConsumer) StartConsuming(ctx context.Context, savePayment func(orderID int, status string)) error {
	defer c.reader.Close()

	for {
		select {
		case <-ctx.Done():
			log.Info("Kafka consumer context done")
			return nil
		default:
			m, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Errorf("Error reading message: %v", err)
				continue
			}

			var order Order
			if err := json.Unmarshal(m.Value, &order); err != nil {
				log.Errorf("Could not unmarshal message: %v", err)
				continue
			}

			log.Infof("Consumed order ID: %d", order.ID)
			savePayment(order.ID, "paid")
		}
	}
}