package engine

import (
	"log/slog"
	"user-ws-api/matcher"
	"user-ws-api/models"
)

type OrderRouter struct {
	matcher    matcher.Matcher
	tradeCh    chan models.Trade
	submitCh   chan models.Order
	assets     map[string]*Asset
	getAssetCh chan getAssetRequest
}

type getAssetRequest struct {
	assetID string
	respCh  chan *Asset
}

func NewOrderRouter(m matcher.Matcher, tradeCh chan models.Trade) *OrderRouter {
	r := &OrderRouter{
		matcher:    m,
		tradeCh:    tradeCh,
		submitCh:   make(chan models.Order, 100),
		assets:     make(map[string]*Asset),
		getAssetCh: make(chan getAssetRequest),
	}
	go r.run()
	return r
}

func (r *OrderRouter) run() {
	for {
		select {
		case order := <-r.submitCh:
			asset, ok := r.assets[order.AssetID]
			if !ok {
				asset = NewAsset(order.AssetID, r.matcher, r.tradeCh)
				r.assets[order.AssetID] = asset
			}
			slog.Debug("OrderRouter.run", "order", order)
			asset.Submit(order)

		case req := <-r.getAssetCh:
			req.respCh <- r.assets[req.assetID]
		}
	}
}

func (r *OrderRouter) Submit(order models.Order) {
	r.submitCh <- order
}
