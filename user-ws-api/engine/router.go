package engine

import (
	"sync"
	"user-ws-api/interfaces"
	"user-ws-api/models"
)

type OrderRouter struct {
	books   map[string]interfaces.OrderSubmitter
	bookMu  sync.RWMutex
	matcher Matcher
	tradeCh chan models.Trade
}

func NewOrderRouter(m Matcher, tradeCh chan models.Trade) *OrderRouter {
	return &OrderRouter{
		books:   make(map[string]interfaces.OrderSubmitter),
		matcher: m,
		tradeCh: tradeCh,
	}
}

func (r *OrderRouter) Submit(order models.Order) {
	r.bookMu.RLock()
	book, exists := r.books[order.AssetID]
	r.bookMu.RUnlock()

	if !exists {
		r.bookMu.Lock()
		book, exists = r.books[order.AssetID]
		if !exists {
			book = NewBook(order.AssetID, r.matcher, r.tradeCh)
			r.books[order.AssetID] = book
		}
		r.bookMu.Unlock()
	}

	go book.Submit(order)
}
