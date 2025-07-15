package main

import (
	"context"
	"os"
	"os/signal"
	"payment"

	log "github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	e := echo.New()

	db, err := payment.NewDB() 

	if err != nil {
		log.WithError(err).Fatal("Could not connect to database")
	}
	defer db.Close()

	consumer := payment.NewKafkaComsumer(
		[]string{"kafka:29092"},
		"orders",
		"payment-group",
	)

	go func() {
		err := consumer.StartConsuming(ctx, func(orderID int, status string){
			p := &payment.Payment{
				OrderID: orderID,
				Status: status,
			}
			if err := db.SavePayment(ctx, p); err != nil {
				log.Errorf("Error saving payment: %v", err)
			}
		})

		if err != nil {
			log.Warnf("Kafka error: %v", err)
		}
	} ()

	Handler := payment.Handler{
		Repo: db,
	}
	

	<-ctx.Done()
	log.Warn("Payment service is Done")
	e.GET("/health", Handler.GetHandler)
	e.Logger.Fatal(e.Start(":1323"))
}