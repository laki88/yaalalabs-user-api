package main

import (
	"database/sql"
	_ "github.com/lib/pq"

	"github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"
	"log"
	"net/http"

	"user-ws-api/ws"
)

func main() {
	connStr := "postgres://user:pass@localhost:5432/userdb?sslmode=disable" // adjust credentials/db
	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	userService := userservice.NewService(sqlDB)

	hub := ws.NewHub(userService)
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	go ws.StartNATSListener(hub)

	log.Println("WebSocket server started on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
