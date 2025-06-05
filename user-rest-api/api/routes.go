package api

import "github.com/go-chi/chi/v5"

func Routes(handler *Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Post("/", handler.CreateUser)
	r.Get("/", handler.ListUsers)
	r.Get("/{id}", handler.GetUser)
	r.Patch("/{id}", handler.UpdateUser)
	r.Delete("/{id}", handler.DeleteUser)

	return r
}
