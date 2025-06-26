package models

import "time"

type OrderSide string

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"
)

type Order struct {
	ID        string
	UserID    string
	AssetID   string
	Quantity  float64
	Price     float64
	Side      OrderSide
	CreatedAt time.Time
}
