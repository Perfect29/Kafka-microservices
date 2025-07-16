package main

import (
	"os/signal"
	"context"
	"os"
	log "github.com/sirupsen/logrus"
	"order"

	"github.com/labstack/echo/v4"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	e := echo.New()

	db, err := order.NewDB()

	if err != nil {
		log.WithError(err).Fatal("Could not connect to database")
	}
	defer db.Close()

	producer := order.NewKafkaProducer([]string{"kafka:29092"}, "orders")
	defer producer.Close()

	consumer := order.NewKafkaConsumer(
		[]string{"kafka:29092"},
		"payment.status",
		"order-status-group",
	)

	go func () {
		if err := consumer.StartConsuming(ctx, db); err != nil {
			log.Errorf("Kafka consumer error: %v", err)
		}
	} ()

	Handler := order.Handler{
		Repo: db,
		Producer: producer,
	}



	e.POST("/orders", Handler.PostHandler)

	go func() {
		<-ctx.Done()
		log.Warn("Shutting down order service")
		e.Shutdown(ctx)
	}()

	e.Logger.Fatal(e.Start(":1323"))
}