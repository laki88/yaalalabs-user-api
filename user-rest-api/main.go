package main

import (
	"database/sql"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal/db"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal/nats"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/laki88/yaalalabs-user-api/user-rest-api/api"
)
import _ "github.com/lib/pq"

func main() {

	conn, err := sql.Open("postgres", "postgres://user:pass@localhost:5432/userdb?sslmode=disable")
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	err = nats.InitNATS("nats://localhost:4222")
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}

	internal.InitValidator()

	queries := db.New(conn)
	handler := api.NewHandler(queries)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/users", api.Routes(handler))
	r.Get("/docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/openapi.yaml")
	})

	fileServer := http.FileServer(http.Dir("./docs/swagger-ui"))
	r.Handle("/doc/*", http.StripPrefix("/doc/", fileServer))

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
