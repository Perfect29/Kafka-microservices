package order

import ()

type Order struct {
	ID     int     `json:"id"`
	Item   string  `json:"item"`
	Amount int     `json:"amount"`
	Price  float64 `json:"price"`
	Status string  `json:"status"`
}

type Handler struct {
	Repo Repo
	Producer Producer
}
