package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log/slog"
	"os"
	"user-ws-api/engine"
	"user-ws-api/internal/config"
	"user-ws-api/matcher"
	"user-ws-api/models"

	"github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"
	"net/http"

	"user-ws-api/ws"
)

func main() {
	// Setup structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Loading config", "path", "config/config.yaml")
	config.LoadConfig("config/config.yaml")
	
	slog.Info("Connecting to database", "driver", config.AppConfig.Database.Driver)
	sqlDB, err := sql.Open(config.AppConfig.Database.Driver, config.AppConfig.Database.URL)
	if err != nil {
		slog.Error("cannot connect to db:", err)
		os.Exit(1)
	}

	userService := userservice.NewService(sqlDB)

	systemMatcher := &matcher.SimpleMatcher{}
	tradeCh := make(chan models.Trade, 100)
	orderRouter := engine.NewOrderRouter(systemMatcher, tradeCh)

	hub := ws.NewHub(userService, orderRouter)
	hub.SetTradeChannel(tradeCh)
	go hub.Run() // todo refactor this to startHub and move these to there

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	slog.Info("Starting NATS listener", "url", config.AppConfig.NATS.URL)
	go ws.StartNATSListener(hub, config.AppConfig.NATS.URL)

	slog.Info("WebSocket server started", "port", config.AppConfig.Server.Port)
	if err := http.ListenAndServe(":"+config.AppConfig.Server.Port, nil); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
	waitForShutdown()

}

func waitForShutdown() {
	fmt.Println("Type 'shutdown' and press Enter to stop the system.")
	var input string
	for {
		fmt.Scanln(&input)
		if input == "shutdown" {
			fmt.Println("Shutdown triggered. Exiting simulation...")
			return
		}
	}
}
