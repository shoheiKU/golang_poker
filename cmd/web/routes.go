package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shoheiKU/web_application/GO_UDEMY/pkg/config"
	"github.com/shoheiKU/web_application/GO_UDEMY/pkg/handlers"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/poker", handlers.Repo.Poker)
	mux.Get("/contact", handlers.Repo.Contact)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	return mux
}
