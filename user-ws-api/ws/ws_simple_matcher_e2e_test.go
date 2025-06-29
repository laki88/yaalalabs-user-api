package ws_test

import (
	"testing"
	"time"
	"user-ws-api/models"
)

func TestWebSocketWithSimpleMatcher_UsingOrderData(t *testing.T) {
	_, users, cleanup, router := SetupTestServer(t)
	now := time.Now()
	// Predefined orders from u1 (buyers) and u2 (sellers)
	buyOrders := []models.Order{
		{ID: "b1", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 101, Quantity: 1, CreatedAt: now},
		{ID: "b2", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 102, Quantity: 2, CreatedAt: now.Add(1 * time.Millisecond)},
		{ID: "b3", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 100, Quantity: 1, CreatedAt: now.Add(3 * time.Millisecond)},
		{ID: "b4", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 99, Quantity: 1, CreatedAt: now.Add(5 * time.Millisecond)},
		{ID: "b5", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 103, Quantity: 1, CreatedAt: now.Add(7 * time.Millisecond)},
	}

	sellOrders := []models.Order{
		{ID: "s1", UserID: "u2", AssetID: "BTC", Side: models.Sell, Price: 100, Quantity: 1, CreatedAt: now},
		{ID: "s2", UserID: "u2", AssetID: "BTC", Side: models.Sell, Price: 102, Quantity: 1, CreatedAt: now.Add(2 * time.Millisecond)},
		{ID: "s3", UserID: "u2", AssetID: "BTC", Side: models.Sell, Price: 101, Quantity: 2, CreatedAt: now.Add(4 * time.Millisecond)},
		{ID: "s4", UserID: "u2", AssetID: "BTC", Side: models.Sell, Price: 105, Quantity: 1, CreatedAt: now.Add(6 * time.Millisecond)},
		{ID: "s5", UserID: "u2", AssetID: "BTC", Side: models.Sell, Price: 98, Quantity: 1, CreatedAt: now.Add(8 * time.Millisecond)},
	}

	for _, o := range buyOrders {
		sendOrder(t, users["u1"], o)
	}
	for _, o := range sellOrders {
		sendOrder(t, users["u2"], o)
	}

	// Read trades
	time.Sleep(5000 * time.Millisecond)

	trades := ReadTradeMessages(t, users["u1"], 1, 2*time.Second)

	t.Logf("✅ %d trades matched", len(trades))
	for _, tr := range trades {
		t.Logf("TRADE: %s bought from %s @ %.2f x %.2f", tr.BuyerID, tr.SellerID, tr.Price, tr.Quantity)
	}

	// Check order book after matching
	book, ok := router.GetBook("BTC")
	if !ok {
		t.Fatal("order book not found")
	}
	t.Logf("📘 Remaining in Buy Book: %d", book.BuyDepth())
	t.Logf("📕 Remaining in Sell Book: %d", book.SellDepth())

	cleanup()
}
