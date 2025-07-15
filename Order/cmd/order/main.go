package main

import (
	log "github.com/sirupsen/logrus"
	"order"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	db, err := order.NewDB()
	producer := order.NewKafkaProducer([]string{"kafka:29092"}, "orders")
	if err != nil {
		log.WithError(err).Fatal("Could not connect to database")
	}
	defer db.Close()

	Handler := order.Handler{
		Repo: db,
		Producer: producer,
	}

	e.POST("/orders", Handler.PostHandler)
	e.Logger.Fatal(e.Start(":1323"))
}