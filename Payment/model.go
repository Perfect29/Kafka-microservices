package payment

import ()

type Payment struct {
	ID      int    `json:"id"`
	OrderID int    `json:"order_id"`
	Status  string `json:"status"`
}

type Order struct {
	ID     int     `json:"id"`
	Item   string  `json:"item"`
	Amount int     `json:"amount"`
	Price  float64 `json:"price"`
	Status string  `json:"status"`
}

type Handler struct {
	Repo     Repo
}
