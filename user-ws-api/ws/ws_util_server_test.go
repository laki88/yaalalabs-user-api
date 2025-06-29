package ws_test

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
	"user-ws-api/config"
	"user-ws-api/engine"
	"user-ws-api/matcher"
	"user-ws-api/models"
	"user-ws-api/ws"
)

func SetupTestServer(t *testing.T) (chan models.Trade, map[string]*websocket.Conn, func(), *engine.OrderRouter) {
	port := fmt.Sprintf("%d", 9000+rand.Intn(1000)) // e.g., 9091, 9134...
	config.AppConfig.Server.Port = port
	addr := "localhost:" + port

	tradeCh := make(chan models.Trade, 20)
	router := engine.NewOrderRouter(&matcher.SimpleMatcher{}, tradeCh)
	hub := ws.NewHub(nil, router)
	hub.SetTradeChannel(tradeCh)

	go hub.Run()

	srv := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ws.ServeWs(hub, w, r)
		}),
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			t.Fatalf("HTTP server failed: %v", err)
		}
	}()

	time.Sleep(200 * time.Millisecond)

	clients := make(map[string]*websocket.Conn)
	for _, uid := range []string{"u1", "u2", "u3", "u4"} {
		u := url.URL{Scheme: "ws", Host: addr, Path: "/ws", RawQuery: "user_id=" + uid}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			t.Fatalf("user %s failed to connect: %v", uid, err)
		}
		clients[uid] = conn
	}

	// Return a cleanup function
	cleanup := func() {
		srv.Close()
		for _, c := range clients {
			c.Close()
		}
	}

	return tradeCh, clients, cleanup, router
}

func SetupTestServerWithUsers(t *testing.T, numUsers int) (chan models.Trade, map[string]*websocket.Conn, func(), *engine.OrderRouter) {
	port := fmt.Sprintf("%d", 9000+rand.Intn(1000))
	config.AppConfig.Server.Port = port
	addr := "localhost:" + port

	tradeCh := make(chan models.Trade, numUsers*5)
	router := engine.NewOrderRouter(&matcher.SimpleMatcher{}, tradeCh)
	hub := ws.NewHub(nil, router)
	hub.SetTradeChannel(tradeCh)

	go hub.Run()

	srv := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ws.ServeWs(hub, w, r)
		}),
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			t.Fatalf("HTTP server failed: %v", err)
		}
	}()

	time.Sleep(200 * time.Millisecond)

	clients := make(map[string]*websocket.Conn)
	for i := 1; i <= numUsers; i++ {
		uid := fmt.Sprintf("u%d", i)
		u := url.URL{Scheme: "ws", Host: addr, Path: "/ws", RawQuery: "user_id=" + uid}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			t.Fatalf("user %s failed to connect: %v", uid, err)
		}
		clients[uid] = conn
	}

	cleanup := func() {
		srv.Close()
		for _, c := range clients {
			c.Close()
		}
	}

	return tradeCh, clients, cleanup, router
}

func SendOrders(t *testing.T, conn *websocket.Conn, orders []models.Order) {
	for _, o := range orders {
		sendOrder(t, conn, o)
	}
}

func ReadTradeMessages(t *testing.T, conn *websocket.Conn, expected int, timeout time.Duration) []models.Trade {
	if expected == 0 {
		return nil // No need to read
	}
	var trades []models.Trade
	deadline := time.Now().Add(timeout)

	for len(trades) < expected && time.Now().Before(deadline) {
		// Always set a read deadline
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

		_, msg, err := conn.ReadMessage()
		if err != nil {
			// Stop reading completely on any read error to avoid panic
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) ||
				strings.Contains(err.Error(), "use of closed network connection") ||
				strings.Contains(err.Error(), "EOF") {
				t.Logf("❌ WebSocket closed unexpectedly: %v", err)
				return trades
			}

			// For temporary timeouts (due to SetReadDeadline)
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}

			// On any other error, also stop
			t.Logf("❌ WebSocket read error: %v", err)
			return trades
		}

		var wsMsg ws.WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			t.Logf("Invalid WS message: %v", err)
			continue
		}

		if wsMsg.Type == "trade" && wsMsg.Entity == "orders" {
			var trade models.Trade
			if err := json.Unmarshal(wsMsg.Payload, &trade); err != nil {
				t.Logf("Invalid trade payload: %v", err)
				continue
			}
			trades = append(trades, trade)
		}
	}

	return trades
}

func sendOrder(t *testing.T, conn *websocket.Conn, order models.Order) {
	payload, _ := json.Marshal(order)
	msg := map[string]any{
		"type":    "order",
		"entity":  "orders",
		"payload": json.RawMessage(payload),
	}
	raw, _ := json.Marshal(msg)
	err := conn.WriteMessage(websocket.TextMessage, raw)
	if err != nil {
		t.Fatalf("failed to send order: %v", err)
	}
}
