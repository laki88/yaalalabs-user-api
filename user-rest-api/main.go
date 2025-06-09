package main

import (
	"database/sql"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal/config"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal/db"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal/nats"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal/repository"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/laki88/yaalalabs-user-api/user-rest-api/api"
)
import _ "github.com/lib/pq"

func main() {
	config.LoadConfig("config/config.yaml")

	conn, err := sql.Open(config.AppConfig.Database.Driver, config.AppConfig.Database.URL)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	err = nats.InitNATS(config.AppConfig.NATS.URL)
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}

	internal.InitValidator()

	queries := db.New(conn)
	repo := repository.NewPostgresUserRepository(queries)
	userService := userservice.NewService(repo)
	handler := api.NewHandler(userService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/users", api.Routes(handler))
	r.Get("/docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/openapi.yaml")
	})

	fileServer := http.FileServer(http.Dir("./docs/swagger-ui"))
	r.Handle("/doc/*", http.StripPrefix("/doc/", fileServer))

	log.Println("Server running at http://localhost:" + config.AppConfig.Server.Port)
	log.Fatal(http.ListenAndServe(":"+config.AppConfig.Server.Port, r))
}
