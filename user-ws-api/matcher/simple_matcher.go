package matcher

import (
	"log/slog"
	"math"
	"user-ws-api/models"
)

type SimpleMatcher struct{}

func (m *SimpleMatcher) Match(order models.Order, book BookView) MatchResult {
	var trades []models.Trade
	remainingQty := order.Quantity
	switch order.Side {
	case models.Buy:
		for remainingQty > 0 {
			slog.Debug("Buy Side", "order", order)
			sell, ok := book.PeekSell()
			if !ok {
				slog.Debug("Buy Side: top sell order not ok")
				// no matching order
				return MatchResult{
					trades,
					remainingQty,
				}
			}
			if order.UserID == sell.UserID {
				slog.Debug("Buy Side: self trade skipping")
				break // prevent self-trade
			}
			if sell.Quantity == 0 || sell.Price > order.Price {
				slog.Debug("Buy Side: No match", "sell.Quantity", sell.Quantity, "sell.Price", sell.Price, "order.Price", order.Price)
				break // no match
			}
			matchQty := math.Min(remainingQty, sell.Quantity)
			slog.Debug("Buy Side", "matchQty", matchQty)
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
			slog.Debug("Buy Side", "remainingQty", remainingQty, "trade", trade)
			if matchQty == sell.Quantity {
				slog.Debug("Buy Side", "matchQty", matchQty)
				book.PopSell()
			} else {
				book.PopSell()
				sell.Quantity -= matchQty
				book.AddSell(sell)
				slog.Info("[MATCH] Buy Side ",
					slog.String("buy_order_id", order.ID),
					slog.String("buy_user_id", order.UserID),
					slog.String("sell_order_id", sell.ID),
					slog.String("sell_user_id", sell.UserID),
					slog.Float64("qty", matchQty),
					slog.Float64("price", sell.Price),
					slog.Float64("buy_remaining", remainingQty),
					slog.Float64("sell_remaining", sell.Quantity),
				)
			}
		}
	case models.Sell:
		for remainingQty > 0 {
			slog.Debug("Sell Side", "order", order)
			buy, ok := book.PeekBuy()
			if !ok {
				slog.Debug("Sell Side: top sell order not ok")
				// no matching order
				return MatchResult{
					trades,
					remainingQty,
				}
			}
			if order.UserID == buy.UserID {
				slog.Debug("Sell Side: self trade skipping")
				break // prevent self-trade
			}
			if buy.Quantity == 0 || buy.Price < order.Price {
				slog.Debug("Sell Side: no match", "sell.Quantity", buy.Quantity, "sell.Price", buy.Price, "order.Price", order.Price)
				break // no match
			}
			matchQty := math.Min(remainingQty, buy.Quantity)
			slog.Debug("Sell Side", "matchQty", matchQty)
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
			slog.Debug("Sell Side", "remainingQty", remainingQty, "trade", trade)
			if matchQty == buy.Quantity {
				slog.Debug("Sell Side", "matchQty", matchQty)
				book.PopBuy()
			} else {
				book.PopBuy()
				buy.Quantity -= matchQty
				book.AddBuy(buy)
				slog.Debug("[MATCH] Sell Side ",
					slog.String("buy_order_id", buy.ID),
					slog.String("buy_user_id", buy.UserID),
					slog.String("sell_order_id", order.ID),
					slog.String("sell_user_id", order.UserID),
					slog.Float64("qty", matchQty),
					slog.Float64("price", buy.Price),
					slog.Float64("buy_remaining", buy.Quantity),
					slog.Float64("sell_remaining", remainingQty),
				)
			}
		}
	}
	if remainingQty > 0 {
		slog.Debug("Order partially filled", "orderID", order.ID, "remainingQty", remainingQty)
	} else if len(trades) > 0 {
		slog.Debug("Order fully filled", "orderID", order.ID)
	} else {
		slog.Debug("Order did not match anything", "orderID", order.ID)
	}

	return MatchResult{
		trades,
		remainingQty,
	}
}
