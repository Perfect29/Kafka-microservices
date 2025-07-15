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