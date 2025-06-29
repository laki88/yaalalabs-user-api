package matcher

import "user-ws-api/models"

type MatchResult struct {
	Trades       []models.Trade
	RemainingQty float64
}

type Matcher interface {
	Match(order models.Order, book BookView) MatchResult
}

type BookView interface {
	PeekBuy() (models.Order, bool)
	PeekSell() (models.Order, bool)
	PopBuy() models.Order
	PopSell() models.Order
	AddBuy(order models.Order)
	AddSell(order models.Order)
}
