package order

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
	SaveOrder(ctx context.Context, order *Order) error
	ChangeStatus(ctx context.Context, payEvent PaymentStatusEvent) error
	CheckStatus(ctx context.Context, id int) (string, error)
}

func NewDB() (*DB, error) {
	dsn := "postgres://user:pass@postgres_order:5432/orderdb?sslmode=disable"
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

	return nil, fmt.Errorf("payment DB connection failed: %w", err)
}

func (db *DB) Close() {
	db.conn.Close(context.Background())
}

func (db *DB) SaveOrder(ctx context.Context, order *Order) error {
	order.Status = "pending"
	query := `
		INSERT INTO orders (item, amount, price, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`
	err := db.conn.QueryRow(ctx, query, order.Item, order.Amount, order.Price, order.Status).Scan(&order.ID)

	if err != nil {
		return fmt.Errorf("could not save order: %w", err)
	}

	log.Infof("Request saved to order database")
	return nil
}

func (db *DB) ChangeStatus(ctx context.Context, payEvent PaymentStatusEvent) error {
	query := `
		UPDATE ORDERS
		SET status = $1
		WHERE id = $2
	`
	_, err := db.conn.Exec(ctx, query, payEvent.Status, payEvent.OrderID)

	if err != nil {
		return err
	}
	return nil
}

func (db *DB) CheckStatus(ctx context.Context, id int) (string, error) {
	var status string
	query := `
		SELECT status
		FROM orders
		WHERE id = $1
	`

	err := db.conn.QueryRow(ctx, query, id).Scan(&status)
	if err != nil {
		return "", err
	}

	return status, nil
}