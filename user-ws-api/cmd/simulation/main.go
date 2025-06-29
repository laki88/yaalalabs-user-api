package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"user-ws-api/models"
)

// WSMessage matches the message envelope expected by the server
type WSMessage struct {
	Type    string      `json:"type"`
	Entity  string      `json:"entity"`
	Payload interface{} `json:"payload"`
}

func main() {
	orders, err := loadOrdersFromCSV("./cmd/testdata/orders1.csv")
	if err != nil {
		slog.Error("failed to load orders: ", "Error", err)
		os.Exit(1)
	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.Before(orders[j].CreatedAt)
	})

	// Open a connection per user and keep them in a map
	conns := map[string]*websocket.Conn{}
	for _, o := range orders {
		if _, exists := conns[o.UserID]; !exists {
			u := url.URL{
				Scheme:   "ws",
				Host:     "localhost:8081",
				Path:     "/ws",
				RawQuery: "user_id=" + o.UserID,
			}
			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				slog.Error("connection error for ", "UserId", o.UserID, "Error", err)
			}
			slog.Info("Connected", "User", o.UserID)
			conns[o.UserID] = conn
		}
	}

	// Start read goroutines to print received messages
	for userID, conn := range conns {
		go func(uid string, c *websocket.Conn) {
			for {
				_, msg, err := c.ReadMessage()
				if err != nil {
					return
				}
				slog.Info("Received: ", "UserId", uid, "Message", msg)
			}
		}(userID, conn)
	}

	base := orders[0].CreatedAt
	start := time.Now()
	for _, order := range orders {
		delay := time.Until(start.Add(order.CreatedAt.Sub(base)))
		if delay > 0 {
			time.Sleep(delay)
		}
		msg := WSMessage{
			Type:    "order",
			Entity:  "orders",
			Payload: order,
		}
		data, _ := json.Marshal(msg)
		conn := conns[order.UserID]
		slog.Info("Sending order: ", "User Id", order.UserID, "Id", order.ID, "At time", time.Now().Format(time.RFC3339Nano))
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			slog.Error("Failed to send order: ", "UserId", order.UserID, "Error", err)
		}
	}

	slog.Info("Simulation finished. Press ENTER to exit...")
	fmt.Scanln()
	for _, conn := range conns {
		err := conn.Close()
		if err != nil {
			return
		}
	}
}

func loadOrdersFromCSV(path string) ([]models.Order, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			slog.Error("Error occurred while closing CSV file", "Error", err)
		}
	}(file)

	r := csv.NewReader(file)
	r.Read()

	var orders []models.Order
	for {
		record, err := r.Read()
		if err != nil {
			break
		}
		order, err := parseOrderRecord(record)
		if err != nil {
			slog.Error("Skipping order due to parse error:", err)
			continue
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func parseOrderRecord(record []string) (models.Order, error) {
	quantity, _ := parseFloat(record[3])
	price, _ := parseFloat(record[4])
	createdAt, _ := time.Parse(time.RFC3339Nano, record[6])
	return models.Order{
		ID:        record[0],
		UserID:    record[1],
		AssetID:   record[2],
		Quantity:  quantity,
		Price:     price,
		Side:      models.OrderSide(record[5]),
		CreatedAt: createdAt,
	}, nil
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
