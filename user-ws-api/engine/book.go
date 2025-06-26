package engine

import (
	"sync"
	"user-ws-api/matcher"
	"user-ws-api/models"
	"user-ws-api/utils"
)

type Matcher interface {
	Match(order models.Order, book matcher.BookView) []models.Trade
}

type OrderQueue interface {
	Push(order models.Order)
	Pop() models.Order
	Peek() (models.Order, bool)
	Len() int
}

type Book struct {
	assetID    string
	buyOrders  OrderQueue
	sellOrders OrderQueue
	mu         sync.Mutex
	matcher    Matcher
	tradeCh    chan<- models.Trade
}

func NewBook(assetID string, matcher Matcher, tradeCh chan<- models.Trade) *Book {
	return &Book{
		assetID: assetID,
		matcher: matcher,
		buyOrders: utils.NewOrderHeapQueue(func(a, b models.Order) bool {
			// MinHeap for buy: lower price first; then earlier time
			if a.Price == b.Price {
				return a.CreatedAt.Before(b.CreatedAt)
			}
			return a.Price < b.Price
		}),
		sellOrders: utils.NewOrderHeapQueue(func(a, b models.Order) bool {
			// MaxHeap for sell: higher price first; then earlier time
			if a.Price == b.Price {
				return a.CreatedAt.Before(b.CreatedAt)
			}
			return a.Price > b.Price
		}),
		tradeCh: tradeCh,
	}
}

func (b *Book) PeekBuy() (models.Order, bool)  { return b.buyOrders.Peek() }
func (b *Book) PeekSell() (models.Order, bool) { return b.sellOrders.Peek() }
func (b *Book) PopBuy() models.Order           { return b.buyOrders.Pop() }
func (b *Book) PopSell() models.Order          { return b.sellOrders.Pop() }

func (b *Book) Submit(order models.Order) {

	b.mu.Lock()
	if order.Side == models.Buy {
		b.buyOrders.Push(order)
	} else {
		b.sellOrders.Push(order)
	}

	trades := b.matcher.Match(order, b)
	b.mu.Unlock()

	for _, trade := range trades {
		b.tradeCh <- trade
	}

	totalFilled := 0.0
	for _, t := range trades {
		totalFilled += t.Quantity
	}
	remaining := order.Quantity - totalFilled

	if remaining > 0 {
		order.Quantity = remaining
		b.mu.Lock()
		if order.Side == models.Buy {
			b.buyOrders.Push(order)
		} else {
			b.sellOrders.Push(order)
		}
		b.mu.Unlock()
	}

}
