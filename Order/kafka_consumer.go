package order

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

type KafkaConsumer struct {
	reader *kafka.Reader
}

func NewKafkaConsumer(brokers []string, topic string, groupID string) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic: topic,
			GroupID: groupID,
		}),
	}
}

func (c *KafkaConsumer) StartConsuming(ctx context.Context, repo Repo) error {
	defer c.reader.Close()

	const MaxRetries = 3
	const delay = time.Second * 2

	dlqWriter := &kafka.Writer{
		Addr: kafka.TCP("kafka:29092"),
		Topic: "payment-status.dlq",
		Balancer: &kafka.LeastBytes{},
	}
	defer dlqWriter.Close()
	for {
		select{
		case <-ctx.Done():
			log.Info("Kafka consumer shutting down")
			return nil
		default:
			msg, err := c.reader.ReadMessage(ctx)

			if err != nil {
				log.Errorf("Could not read message from context: %v", err)
				continue
			}

			var event PaymentStatusEvent

			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Errorf("Failed to Unmarshal from context message: %v", err)
				continue
			}

			log.Infof("Successfully received data from payment.status topic: %+v", event)
			var LastErr error 
			
			for i := 0; i < MaxRetries; i++ {
				err := repo.ChangeStatus(ctx, event)
				if err == nil {
					log.Infof("The status order with id %v successfully changed", event.OrderID)
					LastErr = nil
					break
				}
				LastErr = err
				time.Sleep(delay)
			}

			if LastErr != nil {
				log.Infof("Failed to change the status of order with id %d", event.OrderID)
				msg := kafka.Message{
					Key:[]byte(fmt.Sprintf("order-%d", event.OrderID)),
					Value: msg.Value,
				}
				log.Infof("Sending payment-status message to DLQ")

				err :=  dlqWriter.WriteMessages(ctx, msg)

				if err != nil {
					log.Errorf("Failed to send message to DLQ")
				} else{
					log.Error("Message was send to DLQ")
				}
			}
		}
	} 
}