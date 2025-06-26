package models

import "time"

type Trade struct {
	BuyOrderID  string
	SellOrderID string
	BuyerID     string `json:"buyer_id"`
	SellerID    string `json:"seller_id"`
	Quantity    float64
	Price       float64
	Timestamp   time.Time
}
