package ws_test

import (
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
	"time"
	"user-ws-api/models"
)

func TestMain(m *testing.M) {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(handler))
	os.Exit(m.Run())
}

func TestPartialMatch(t *testing.T) {
	_, users, cleanup, router := SetupTestServer(t)

	buyOrders := []models.Order{
		{ID: "b1", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 100, Quantity: 3, CreatedAt: time.Now()},
	}
	sellOrders := []models.Order{
		{ID: "s1", UserID: "u2", AssetID: "BTC", Side: models.Sell, Price: 100, Quantity: 1, CreatedAt: time.Now()},
	}

	SendOrders(t, users["u1"], buyOrders)
	SendOrders(t, users["u2"], sellOrders)

	time.Sleep(5000 * time.Millisecond)

	trades := ReadTradeMessages(t, users["u1"], 1, 2*time.Second)

	assert.Equal(t, 1, len(trades), "Expected 1 trade")

	if len(trades) == 1 {
		tr := trades[0]
		assert.Equal(t, "u1", tr.BuyerID, "Unexpected buyer ID")
		assert.Equal(t, "u2", tr.SellerID, "Unexpected seller ID")
		assert.Equal(t, 1, int(tr.Quantity), "Unexpected trade quantity")
	}

	asset := router.GetAsset("BTC")
	assert.Equal(t, 1, asset.GetBookDepth().BuyDepth, "Expected 1 order remaining in buy book")
	assert.Equal(t, 0, asset.GetBookDepth().SellDepth, "Expected no orders in sell book")

	cleanup()
}

func TestFullMatch(t *testing.T) {
	_, users, cleanup, router := SetupTestServer(t)
	buy := models.Order{ID: "b1", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 100, Quantity: 1, CreatedAt: time.Now()}
	sell := models.Order{ID: "s1", UserID: "u2", AssetID: "BTC", Side: models.Sell, Price: 100, Quantity: 1, CreatedAt: time.Now()}
	SendOrders(t, users["u1"], []models.Order{buy})
	SendOrders(t, users["u2"], []models.Order{sell})
	time.Sleep(time.Second)
	trades := ReadTradeMessages(t, users["u1"], 1, time.Second*2)
	assert.Len(t, trades, 1, "Expected 1 trade")

	if len(trades) == 1 {
		assert.Equal(t, float64(1), trades[0].Quantity, "Unexpected trade quantity")
	}

	asset := router.GetAsset("BTC")
	assert.Equal(t, 0, asset.GetBookDepth().BuyDepth, "Buy book should be empty after full match")
	assert.Equal(t, 0, asset.GetBookDepth().SellDepth, "Sell book should be empty after full match")

	cleanup()
}

func TestPartialMatch_BuyLarger(t *testing.T) {
	_, users, cleanup, router := SetupTestServer(t)
	buy := models.Order{ID: "b1", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 100, Quantity: 3, CreatedAt: time.Now()}
	sell := models.Order{ID: "s1", UserID: "u2", AssetID: "BTC", Side: models.Sell, Price: 100, Quantity: 1, CreatedAt: time.Now()}
	SendOrders(t, users["u1"], []models.Order{buy})
	time.Sleep(300 * time.Millisecond)
	SendOrders(t, users["u2"], []models.Order{sell})
	time.Sleep(2 * time.Second)
	trades := ReadTradeMessages(t, users["u1"], 1, time.Second*2)
	assert.Len(t, trades, 1, "Expected 1 trade for partial match")

	if len(trades) == 1 {
		assert.Equal(t, float64(1), trades[0].Quantity, "Unexpected trade quantity")
	}

	asset := router.GetAsset("BTC")
	assert.Equal(t, 1, asset.GetBookDepth().BuyDepth, "Expected 1 buy order remaining")
	assert.Equal(t, 0, asset.GetBookDepth().SellDepth, "Expected no sell orders remaining")

	cleanup()
}

func TestPartialMatch_SellLarger(t *testing.T) {
	_, users, cleanup, router := SetupTestServer(t)
	buy := models.Order{ID: "b1", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 100, Quantity: 1, CreatedAt: time.Now()}
	sell := models.Order{ID: "s1", UserID: "u2", AssetID: "BTC", Side: models.Sell, Price: 100, Quantity: 3, CreatedAt: time.Now()}
	SendOrders(t, users["u1"], []models.Order{buy})
	SendOrders(t, users["u2"], []models.Order{sell})
	time.Sleep(time.Second)
	trades := ReadTradeMessages(t, users["u1"], 1, time.Second*2)
	assert.Len(t, trades, 1, "Expected 1 trade for partial match")

	if len(trades) == 1 {
		assert.Equal(t, float64(1), trades[0].Quantity, "Unexpected trade quantity")
		assert.Equal(t, "u1", trades[0].BuyerID, "Unexpected buyer ID")
		assert.Equal(t, "u2", trades[0].SellerID, "Unexpected seller ID")
		assert.Equal(t, 100.0, trades[0].Price, "Unexpected trade price")
	}

	asset := router.GetAsset("BTC")
	assert.Equal(t, 0, asset.GetBookDepth().BuyDepth, "Expected no buy orders remaining")
	assert.Equal(t, 1, asset.GetBookDepth().SellDepth, "Expected 1 sell order remaining")

	cleanup()
}

func TestNoMatch_PriceMismatch(t *testing.T) {
	_, users, cleanup, router := SetupTestServer(t)
	buy := models.Order{ID: "b1", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 90, Quantity: 1, CreatedAt: time.Now()}
	sell := models.Order{ID: "s1", UserID: "u2", AssetID: "BTC", Side: models.Sell, Price: 100, Quantity: 1, CreatedAt: time.Now()}
	SendOrders(t, users["u1"], []models.Order{buy})
	SendOrders(t, users["u2"], []models.Order{sell})
	time.Sleep(time.Second)
	trades := ReadTradeMessages(t, users["u1"], 0, time.Second*2)
	assert.Len(t, trades, 0, "Expected 0 trades due to price mismatch")

	asset := router.GetAsset("BTC")
	assert.Equal(t, 1, asset.GetBookDepth().BuyDepth, "Expected 1 buy order to remain in book")
	assert.Equal(t, 1, asset.GetBookDepth().SellDepth, "Expected 1 sell order to remain in book")

	cleanup()
}

func TestCrossedOrders(t *testing.T) {
	_, users, cleanup, router := SetupTestServer(t)
	sell := models.Order{ID: "s1", UserID: "u2", AssetID: "BTC", Side: models.Sell, Price: 100, Quantity: 1, CreatedAt: time.Now()}
	buy := models.Order{ID: "b1", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 100, Quantity: 1, CreatedAt: time.Now().Add(100 * time.Millisecond)}
	SendOrders(t, users["u2"], []models.Order{sell})
	time.Sleep(100 * time.Millisecond)
	SendOrders(t, users["u1"], []models.Order{buy})
	time.Sleep(time.Second)
	trades := ReadTradeMessages(t, users["u1"], 1, time.Second*2)
	assert.Len(t, trades, 1, "Expected 1 trade on crossed orders")

	if len(trades) == 1 {
		assert.Equal(t, float64(1), trades[0].Quantity, "Unexpected trade quantity")
		assert.Equal(t, "u1", trades[0].BuyerID, "Unexpected buyer ID")
		assert.Equal(t, "u2", trades[0].SellerID, "Unexpected seller ID")
		assert.Equal(t, 100.0, trades[0].Price, "Unexpected trade price")
	}

	asset := router.GetAsset("BTC")
	assert.Equal(t, 0, asset.GetBookDepth().BuyDepth, "Expected buy book to be empty after trade")
	assert.Equal(t, 0, asset.GetBookDepth().SellDepth, "Expected sell book to be empty after trade")

	cleanup()
}

func TestMultipleBuyersSingleSeller(t *testing.T) {
	_, users, cleanup, router := SetupTestServer(t)
	buys := []models.Order{
		{ID: "b1", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 100, Quantity: 1, CreatedAt: time.Now()},
		{ID: "b2", UserID: "u3", AssetID: "BTC", Side: models.Buy, Price: 101, Quantity: 1, CreatedAt: time.Now()},
	}
	sell := models.Order{ID: "s1", UserID: "u2", AssetID: "BTC", Side: models.Sell, Price: 100, Quantity: 2, CreatedAt: time.Now()}
	SendOrders(t, users["u1"], []models.Order{buys[0]})
	SendOrders(t, users["u3"], []models.Order{buys[1]})
	SendOrders(t, users["u2"], []models.Order{sell})
	time.Sleep(time.Second)
	tradesU1 := ReadTradeMessages(t, users["u1"], 1, time.Second*2)
	tradesU3 := ReadTradeMessages(t, users["u3"], 1, time.Second*2)
	assert.Len(t, tradesU1, 1, "Expected 1 trade for u1")
	assert.Len(t, tradesU3, 1, "Expected 1 trade for u3")

	if len(tradesU1) == 1 {
		assert.Equal(t, float64(1), tradesU1[0].Quantity, "Unexpected quantity for u1")
		assert.Equal(t, "u1", tradesU1[0].BuyerID, "Unexpected buyer for u1")
		assert.Equal(t, "u2", tradesU1[0].SellerID, "Unexpected seller for u1")
	}

	if len(tradesU3) == 1 {
		assert.Equal(t, float64(1), tradesU3[0].Quantity, "Unexpected quantity for u3")
		assert.Equal(t, "u3", tradesU3[0].BuyerID, "Unexpected buyer for u3")
		assert.Equal(t, "u2", tradesU3[0].SellerID, "Unexpected seller for u3")
	}

	asset := router.GetAsset("BTC")
	assert.Equal(t, 0, asset.GetBookDepth().BuyDepth, "Expected buy book to be empty")
	assert.Equal(t, 0, asset.GetBookDepth().SellDepth, "Expected sell book to be empty")

	cleanup()
}

func TestMultipleSellersSingleBuyer(t *testing.T) {
	_, users, cleanup, router := SetupTestServer(t)
	sells := []models.Order{
		{ID: "s1", UserID: "u2", AssetID: "BTC", Side: models.Sell, Price: 100, Quantity: 1, CreatedAt: time.Now()},
		{ID: "s2", UserID: "u3", AssetID: "BTC", Side: models.Sell, Price: 99, Quantity: 1, CreatedAt: time.Now()},
	}
	buy := models.Order{ID: "b1", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 101, Quantity: 2, CreatedAt: time.Now()}
	SendOrders(t, users["u2"], []models.Order{sells[0]})
	SendOrders(t, users["u3"], []models.Order{sells[1]})
	SendOrders(t, users["u1"], []models.Order{buy})
	time.Sleep(time.Second)
	tradesU2 := ReadTradeMessages(t, users["u2"], 1, time.Second*2)
	tradesU3 := ReadTradeMessages(t, users["u3"], 1, time.Second*2)
	assert.Len(t, tradesU2, 1, "Expected 1 trade for seller u2")
	assert.Len(t, tradesU3, 1, "Expected 1 trade for seller u3")

	if len(tradesU2) == 1 {
		assert.Equal(t, float64(1), tradesU2[0].Quantity, "Unexpected quantity for u2")
		assert.Equal(t, "u1", tradesU2[0].BuyerID, "Unexpected buyer for u2")
		assert.Equal(t, "u2", tradesU2[0].SellerID, "Unexpected seller for u2")
	}

	if len(tradesU3) == 1 {
		assert.Equal(t, float64(1), tradesU3[0].Quantity, "Unexpected quantity for u3")
		assert.Equal(t, "u1", tradesU3[0].BuyerID, "Unexpected buyer for u3")
		assert.Equal(t, "u3", tradesU3[0].SellerID, "Unexpected seller for u3")
	}

	asset := router.GetAsset("BTC")
	assert.Equal(t, 0, asset.GetBookDepth().BuyDepth, "Expected buy book to be empty")
	assert.Equal(t, 0, asset.GetBookDepth().SellDepth, "Expected sell book to be empty")

	cleanup()
}

func TestPriorityPriceTieBreaking(t *testing.T) {
	_, users, cleanup, _ := SetupTestServer(t)
	buy1 := models.Order{ID: "b1", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 100, Quantity: 1, CreatedAt: time.Now()}
	buy2 := models.Order{ID: "b2", UserID: "u2", AssetID: "BTC", Side: models.Buy, Price: 100, Quantity: 1, CreatedAt: time.Now().Add(10 * time.Millisecond)}
	sell := models.Order{ID: "s1", UserID: "u3", AssetID: "BTC", Side: models.Sell, Price: 100, Quantity: 1, CreatedAt: time.Now()}

	SendOrders(t, users["u1"], []models.Order{buy1})
	SendOrders(t, users["u2"], []models.Order{buy2})
	time.Sleep(500 * time.Millisecond)
	SendOrders(t, users["u3"], []models.Order{sell})
	time.Sleep(time.Second)
	trades := ReadTradeMessages(t, users["u1"], 1, 2*time.Second)
	assert.Len(t, trades, 1, "Expected 1 trade for priority match on same price")
	if len(trades) == 1 {
		assert.Equal(t, "u1", trades[0].BuyerID, "Older order (u1) should be prioritized")
		assert.Equal(t, "u3", trades[0].SellerID, "Seller should be u3")
		assert.Equal(t, float64(1), trades[0].Quantity, "Expected full match of 1 quantity")
		assert.Equal(t, 100.0, trades[0].Price, "Expected matched price of 100")
	}
	cleanup()
}

func TestSelfMatchPrevention(t *testing.T) {
	_, users, cleanup, router := SetupTestServer(t)
	buy := models.Order{ID: "b1", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 100, Quantity: 1, CreatedAt: time.Now()}
	sell := models.Order{ID: "s1", UserID: "u1", AssetID: "BTC", Side: models.Sell, Price: 100, Quantity: 1, CreatedAt: time.Now()}

	SendOrders(t, users["u1"], []models.Order{buy, sell})
	time.Sleep(time.Second)
	trades := ReadTradeMessages(t, users["u1"], 0, 2*time.Second)
	assert.Len(t, trades, 0, "Expected 0 trades due to self-match prevention")

	asset := router.GetAsset("BTC")
	assert.Equal(t, 1, asset.GetBookDepth().BuyDepth, "Expected 1 buy order to remain in book")
	assert.Equal(t, 1, asset.GetBookDepth().SellDepth, "Expected 1 sell order to remain in book")

	cleanup()
}

func TestInvalidOrdersAreIgnored(t *testing.T) {
	_, users, cleanup, router := SetupTestServer(t)
	invalid := models.Order{ID: "x1", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: -100, Quantity: 0, CreatedAt: time.Now()}
	SendOrders(t, users["u1"], []models.Order{invalid})
	time.Sleep(500 * time.Millisecond)
	asset := router.GetAsset("BTC")
	assert.Equal(t, 0, asset.GetBookDepth().BuyDepth, "Expected no buy orders for invalid input")
	assert.Equal(t, 0, asset.GetBookDepth().SellDepth, "Expected no sell orders for invalid input")

	cleanup()
}

func TestZeroQuantityOrderIsIgnored(t *testing.T) {
	_, users, cleanup, router := SetupTestServer(t)
	zeroQty := models.Order{ID: "z1", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 100, Quantity: 0, CreatedAt: time.Now()}
	SendOrders(t, users["u1"], []models.Order{zeroQty})
	time.Sleep(500 * time.Millisecond)
	asset := router.GetAsset("BTC")
	assert.Equal(t, 0, asset.GetBookDepth().BuyDepth, "Expected zero quantity order to be ignored")
	assert.Equal(t, 0, asset.GetBookDepth().SellDepth, "Expected no sell orders as none were sent")

	cleanup()
}

func TestSameUserMultipleOrders(t *testing.T) {
	_, users, cleanup, _ := SetupTestServer(t)
	orders := []models.Order{
		{ID: "o1", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 100, Quantity: 1, CreatedAt: time.Now()},
		{ID: "o2", UserID: "u1", AssetID: "BTC", Side: models.Sell, Price: 99, Quantity: 1, CreatedAt: time.Now()},
		{ID: "o3", UserID: "u1", AssetID: "BTC", Side: models.Sell, Price: 101, Quantity: 1, CreatedAt: time.Now()},
	}
	SendOrders(t, users["u1"], orders)
	time.Sleep(time.Second)
	trades := ReadTradeMessages(t, users["u1"], 0, 2*time.Second)
	assert.Len(t, trades, 0, "Expected 0 trades due to self-trade prevention")
	cleanup()
}

func TestMultiAssetMatching(t *testing.T) {
	_, users, cleanup, _ := SetupTestServer(t)
	orders := []models.Order{
		{ID: "btc-b1", UserID: "u1", AssetID: "BTC", Side: models.Buy, Price: 100, Quantity: 1, CreatedAt: time.Now()},
		{ID: "btc-s1", UserID: "u2", AssetID: "BTC", Side: models.Sell, Price: 100, Quantity: 1, CreatedAt: time.Now()},
		{ID: "eth-b1", UserID: "u3", AssetID: "ETH", Side: models.Buy, Price: 200, Quantity: 1, CreatedAt: time.Now()},
		{ID: "eth-s1", UserID: "u4", AssetID: "ETH", Side: models.Sell, Price: 200, Quantity: 1, CreatedAt: time.Now()},
	}
	SendOrders(t, users["u1"], []models.Order{orders[0]})
	SendOrders(t, users["u2"], []models.Order{orders[1]})
	SendOrders(t, users["u3"], []models.Order{orders[2]})
	SendOrders(t, users["u4"], []models.Order{orders[3]})
	time.Sleep(time.Second)
	tradesBTC := ReadTradeMessages(t, users["u1"], 1, 2*time.Second)
	tradesETH := ReadTradeMessages(t, users["u3"], 1, 2*time.Second)
	assert.Len(t, tradesBTC, 1, "Expected 1 BTC trade for user u1")
	assert.Len(t, tradesETH, 1, "Expected 1 ETH trade for user u3")

	if len(tradesBTC) == 1 {
		assert.Equal(t, "u1", tradesBTC[0].BuyerID, "Unexpected BTC buyer ID")
		assert.Equal(t, "u2", tradesBTC[0].SellerID, "Unexpected BTC seller ID")
		assert.Equal(t, float64(1), tradesBTC[0].Quantity, "Unexpected BTC quantity")
		assert.Equal(t, 100.0, tradesBTC[0].Price, "Unexpected BTC price")
	}

	if len(tradesETH) == 1 {
		assert.Equal(t, "u3", tradesETH[0].BuyerID, "Unexpected ETH buyer ID")
		assert.Equal(t, "u4", tradesETH[0].SellerID, "Unexpected ETH seller ID")
		assert.Equal(t, float64(1), tradesETH[0].Quantity, "Unexpected ETH quantity")
		assert.Equal(t, 200.0, tradesETH[0].Price, "Unexpected ETH price")
	}
	cleanup()
}
