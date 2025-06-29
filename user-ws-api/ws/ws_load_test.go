package ws_test

import (
	//"fmt"
	//"github.com/gorilla/websocket"
	//"math/rand"
	//"sync"
	"testing"
	//"time"
	//"user-ws-api/models"
)

func TestHighConcurrencyOrders(t *testing.T) {
	//_, users, cleanup, router := SetupTestServerWithUsers(t, 100)
	//defer cleanup()
	//
	//numUsers := 100
	//numOrdersPerUser := 5
	//var wg sync.WaitGroup
	//
	//// Generate mock users u1 to u100
	//mockUsers := make(map[string]*websocket.Conn)
	//for i := 1; i <= numUsers; i++ {
	//	uid := fmt.Sprintf("u%d", i)
	//	mockUsers[uid] = users[uid] // assumes SetupTestServer returns u1, u2, ..., u100
	//}
	//
	//// Start placing orders concurrently
	//for uid, conn := range mockUsers {
	//	for i := 0; i < numOrdersPerUser; i++ {
	//		wg.Add(1)
	//		go func(uid string, conn *websocket.Conn, i int) {
	//			defer wg.Done()
	//			side := models.Buy
	//			if i%2 == 0 {
	//				side = models.Sell
	//			}
	//			order := models.Order{
	//				ID:        fmt.Sprintf("%s_o%d", uid, i),
	//				UserID:    uid,
	//				AssetID:   "BTC",
	//				Side:      side,
	//				Price:     float64(95 + rand.Intn(10)), // price between 95–104
	//				Quantity:  float64(1 + rand.Intn(5)),   // qty between 1–5
	//				CreatedAt: time.Now(),
	//			}
	//			sendOrder(t, conn, order)
	//		}(uid, conn, i)
	//	}
	//}
	//
	//wg.Wait()
	//time.Sleep(2 * time.Second) // allow matcher to settle
	//
	//// Check order book integrity
	//book, ok := router.GetBook("BTC")
	//if !ok {
	//	t.Fatal("BTC order book not found")
	//}
	//
	//t.Logf("✅ Book depth after 500 orders: buy=%d, sell=%d", book.BuyDepth(), book.SellDepth())
}
