package engine

import (
	"log/slog"
	"user-ws-api/matcher"
	"user-ws-api/models"
	"user-ws-api/utils"
)

type OrderMessage struct {
	Order models.Order
}

type Asset struct {
	book       *Book
	submitCh   chan models.Order
	depthReqCh chan chan BookDepthResponse
}

type BookDepthResponse struct {
	BuyDepth  int
	SellDepth int
}

func NewAsset(assetID string, matcher matcher.Matcher, tradeCh chan<- models.Trade) *Asset {
	asset := &Asset{
		book:       NewBook(assetID, matcher, tradeCh),
		submitCh:   make(chan models.Order, 100),
		depthReqCh: make(chan chan BookDepthResponse),
	}
	go asset.run()
	return asset
}

func (a *Asset) run() {
	for {
		select {
		case order := <-a.submitCh:
			a.book.Submit(order)

		case respCh := <-a.depthReqCh:
			respCh <- BookDepthResponse{
				BuyDepth:  a.book.BuyDepth(),
				SellDepth: a.book.SellDepth(),
			}
		}
	}
}

func (a *Asset) GetBookDepth() BookDepthResponse {
	respCh := make(chan BookDepthResponse)
	a.depthReqCh <- respCh
	return <-respCh
}

func (a *Asset) Submit(order models.Order) {
	a.submitCh <- order
}

func (r *OrderRouter) GetBook(assetID string) (*Book, bool) {
	asset, ok := r.assets[assetID]
	if !ok {
		return nil, false
	}
	return asset.book, true
}

func (r *OrderRouter) GetAsset(assetID string) *Asset {
	respCh := make(chan *Asset)
	r.getAssetCh <- getAssetRequest{
		assetID: assetID,
		respCh:  respCh,
	}
	return <-respCh
}

type Book struct {
	assetID    string
	buyOrders  *utils.OrderHeapQueue
	sellOrders *utils.OrderHeapQueue
	matcher    matcher.Matcher
	tradeCh    chan<- models.Trade
}

func NewBook(assetID string, matcher matcher.Matcher, tradeCh chan<- models.Trade) *Book {
	buyQueue := utils.NewOrderHeapQueue(func(a, b models.Order) bool {
		if a.Price == b.Price {
			return a.CreatedAt.Before(b.CreatedAt)
		}
		return a.Price < b.Price
	})
	sellQueue := utils.NewOrderHeapQueue(func(a, b models.Order) bool {
		if a.Price == b.Price {
			return a.CreatedAt.Before(b.CreatedAt)
		}
		return a.Price > b.Price
	})

	return &Book{
		assetID:    assetID,
		matcher:    matcher,
		buyOrders:  buyQueue,
		sellOrders: sellQueue,
		tradeCh:    tradeCh,
	}
}

func (b *Book) Submit(order models.Order) {
	slog.Debug("Book.Submit", "order", order)

	matchResult := b.matcher.Match(order, b)
	slog.Debug("Book.Submit after Match", "matchResult", matchResult)

	order.Quantity = matchResult.RemainingQty
	slog.Debug("Book.Submit", "order.Quantity", order.Quantity)
	if order.Quantity > 0 {
		if order.Side == models.Buy {
			b.buyOrders.Push(order)
		} else {
			b.sellOrders.Push(order)
		}
	}

	for _, trade := range matchResult.Trades {
		b.tradeCh <- trade
	}
}

func (b *Book) PeekBuy() (models.Order, bool)  { return b.buyOrders.Peek() }
func (b *Book) PeekSell() (models.Order, bool) { return b.sellOrders.Peek() }
func (b *Book) PopBuy() models.Order           { return b.buyOrders.Pop() }
func (b *Book) PopSell() models.Order          { return b.sellOrders.Pop() }
func (b *Book) AddSell(order models.Order)     { b.sellOrders.Push(order) }
func (b *Book) AddBuy(order models.Order)      { b.buyOrders.Push(order) }

func (b *Book) BuyDepth() int {
	return b.buyOrders.Len()
}

func (b *Book) SellDepth() int {
	return b.sellOrders.Len()
}
