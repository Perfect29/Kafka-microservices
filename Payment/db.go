package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
) 

type DB struct {
	conn *pgx.Conn
}

type Repo interface {
	SavePayment(ctx context.Context, payment *Payment) error
}

func NewDB() (*DB, error) {
	dsn := "postgres://user:pass@postgres_payment:5432/paymentdb?sslmode=disable"
	var conn *pgx.Conn
	var err error

	for i := 1; i <= 10; i++ {
		conn, err = pgx.Connect(context.Background(), dsn)
		if err == nil {
			log.Infof("Payment DB connected")
			return &DB{conn: conn}, nil
		}
		log.Infof("Payment DB not ready (%d/10): %v\n", i, err)
		time.Sleep(3 * time.Second)
	}

	return nil, fmt.Errorf("Payment DB connection failed: %w", err)
}

func (db *DB) Close() {
	db.conn.Close(context.Background())
}

func (db *DB) SavePayment(ctx context.Context, payment *Payment) error {
	// Test case
	// if payment.OrderID == 998 {
	// 	return fmt.Errorf("💥 forced failure for OrderID 999")
	// }
	payment.Status = "paid"
	query := `
		INSERT INTO payments (order_id, status)
		VALUES ($1, $2)
		RETURNING id;
	`
	err := db.conn.QueryRow(ctx, query, payment.OrderID, payment.Status).Scan(&payment.ID)

	if err != nil {
		log.Error("Could not save order in payment database")
		return fmt.Errorf("could not save order: %w", err)
	}

	log.Infof("Request saved to payment database")
	return nil
}

func (db *DB) CheckDuplicate(ctx context.Context, eventID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM processed_events WHERE event_id = $1
		)
	`
	var exists bool

	err := db.conn.QueryRow(ctx, query, eventID).Scan(&exists)
	if err != nil {
		return false, err
	}
	log.Info("Order was checked for duplicate")
	return exists, nil
}

func (db *DB) MarkProcessed(ctx context.Context, eventID string) error {
	query := `
		INSERT INTO processed_events (event_id)
		VALUES (
			$1
		)
	`

	_, err := db.conn.Exec(ctx, query, eventID)

	if err != nil {
		return err
	}
	log.Infof("Event with id %v is marked and processed", eventID)
	return nil
}