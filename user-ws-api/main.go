package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"user-ws-api/internal/config"

	"github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"
	"log"
	"net/http"

	"user-ws-api/ws"
)

func main() {
	config.LoadConfig("config/config.yaml")
	sqlDB, err := sql.Open(config.AppConfig.Database.Driver, config.AppConfig.Database.URL)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	userService := userservice.NewService(sqlDB)

	hub := ws.NewHub(userService)
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	go ws.StartNATSListener(hub, config.AppConfig.NATS.URL)

	log.Println("WebSocket server started on :" + config.AppConfig.Server.Port)
	log.Fatal(http.ListenAndServe(":"+config.AppConfig.Server.Port, nil))
}
