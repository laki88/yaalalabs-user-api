package matcher

import (
	"log/slog"
	"math"
	"user-ws-api/models"
)

type BookView interface {
	PeekBuy() (models.Order, bool)
	PeekSell() (models.Order, bool)
	PopBuy() models.Order
	PopSell() models.Order
}

type SimpleMatcher struct{}

func (m *SimpleMatcher) Match(order models.Order, book BookView) []models.Trade {
	var trades []models.Trade
	remainingQty := order.Quantity
	switch order.Side {
	case models.Buy:
		for remainingQty > 0 {
			sell, ok := book.PeekSell()
			if !ok {
				// no matching order
				return nil
			}
			if order.UserID == sell.UserID {
				break // prevent self-trade
			}
			if sell.Quantity == 0 || sell.Price > order.Price {
				break // no match
			}

			matchQty := math.Min(remainingQty, sell.Quantity)
			trade := models.Trade{
				BuyOrderID:  order.ID,
				SellOrderID: sell.ID,
				BuyerID:     order.UserID,
				SellerID:    sell.UserID,
				Quantity:    matchQty,
				Price:       sell.Price,
				Timestamp:   order.CreatedAt,
			}
			trades = append(trades, trade)

			remainingQty -= matchQty
			if matchQty == sell.Quantity {
				book.PopSell()
			} else {
				sell.Quantity -= matchQty
				slog.Info(
					"[MATCH] BUY %s (user %s) matched SELL %s (user %s): qty=%.2f, price=%.2f, orderRemaining=%.2f, sellRemaining=%.2f",
					order.ID, order.UserID, sell.ID, sell.UserID,
					matchQty, sell.Price, remainingQty, sell.Quantity,
				)
				break
			}
		}
	case models.Sell:
		for remainingQty > 0 {
			buy, ok := book.PeekBuy()
			if !ok {
				// no matching order
				return nil
			}
			if buy.Quantity == 0 || buy.Price < order.Price {
				break // no match
			}
			if order.UserID == buy.UserID {
				break // prevent self-trade
			}

			matchQty := math.Min(remainingQty, buy.Quantity)
			trade := models.Trade{
				BuyOrderID:  buy.ID,
				SellOrderID: order.ID,
				BuyerID:     buy.UserID,
				SellerID:    order.UserID,
				Quantity:    matchQty,
				Price:       buy.Price,
				Timestamp:   order.CreatedAt,
			}
			trades = append(trades, trade)

			remainingQty -= matchQty
			if matchQty == buy.Quantity {
				book.PopBuy()
			} else {
				buy.Quantity -= matchQty
				slog.Info(
					"[MATCH] SELL %s (user %s) matched BUY %s (user %s): qty=%.2f, price=%.2f, orderRemaining=%.2f, buyRemaining=%.2f",
					order.ID, order.UserID, buy.ID, buy.UserID,
					matchQty, buy.Price, remainingQty, buy.Quantity,
				)
				break
			}
		}
	}
	if remainingQty > 0 {
		slog.Info("Order partially filled", "orderID", order.ID, "remainingQty", remainingQty)
	} else if len(trades) > 0 {
		slog.Info("Order fully filled", "orderID", order.ID)
	} else {
		slog.Info("Order did not match anything", "orderID", order.ID)
	}
	return trades
}
