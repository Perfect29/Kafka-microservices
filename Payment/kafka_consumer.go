package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

type Kafka struct {
	reader *kafka.Reader
	writer *kafka.Writer
}


func NewKafkaComsumer(brokers []string, topic string, groupID string) *Kafka{
	return &Kafka{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic: topic,
			GroupID: groupID,
		}),
		writer: &kafka.Writer{
			Addr: kafka.TCP("kafka:29092"),
			Topic: "payment.status",
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (c *Kafka) StartConsuming(ctx context.Context, db *DB) error {
	defer c.reader.Close()
	defer c.writer.Close()

	const MaxRetries = 3
	const delay = time.Second * 2

	dlqWriter := &kafka.Writer{
		Addr: kafka.TCP("kafka:29092"),
		Topic: "orders.dlq",
		Balancer: &kafka.LeastBytes{},
	}

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

			var order_event OrderEvent
			if err := json.Unmarshal(m.Value, &order_event); err != nil {
				log.Errorf("Could not unmarshal message: %v", err)
				continue
			}
			ok, err := db.CheckDuplicate(ctx, order_event.EventID)

			if err != nil {
				log.Infof("Error checking for duplicate: %v", err)
				return err
			}

			if ok {
				log.Info("It is a duplicate")
				continue
			} 

			var LastErr error
			var PayEvent PaymentStatusEvent

			PayEvent.OrderID = order_event.Order.ID
			PayEvent.EventID = uuid.New().String()
			for i := 0; i < MaxRetries; i++ {
				log.Debugf("Kafka making attemp number: %v", i + 1)
				payment := &Payment{
					OrderID: order_event.Order.ID,
					Status: "paid",
				}
				
				err := db.SavePayment(ctx, payment)
				if err == nil {
					err = db.MarkProcessed(ctx, order_event.EventID)
					if err == nil {
						log.Infof("Successfully processed event: %v", order_event.EventID)
						LastErr = err
						break
					}
				}

				LastErr = err
				log.Warnf("Retry %d out of %d failed: %v", i + 1, MaxRetries, err)
				time.Sleep(delay)
			}


			if LastErr != nil {
				PayEvent.Status = "cancelled"
				log.Errorf("All retries failed. Sending an event to DLQ: %v", order_event.EventID)
				body, err := json.Marshal(order_event) 

				if err != nil {
					log.Errorf("Failed to Marshal order event: %v", err)
					continue
				}
				msg := kafka.Message{
					Key: []byte(fmt.Sprintf("order-%d", order_event.Order.ID)),
					Value: body,
				}

				err = dlqWriter.WriteMessages(ctx, msg) 
				if err != nil {
					log.Errorf("Failed to write to DLQ: %v", err)
				} else {
					log.Infof("message sent to DLQ %v", order_event.EventID)
				} 
			} else {
				PayEvent.Status = "paid"
			}
			value, err := json.Marshal(PayEvent)
			if err != nil {
				log.Errorf("Could not Marshal PaymentStatusEvent: %v", err)
			}
			msg := kafka.Message{
				Key: []byte(fmt.Sprintf("payment-%d", PayEvent.OrderID)),
				Value: value,
			}

			err = c.writer.WriteMessages(ctx, msg)
			if err != nil {
				log.Errorf("Could not write message to the topic: %v", err)
			}
			log.Infof("Order with id %d processed as paid", order_event.Order.ID)
		}
	}
}