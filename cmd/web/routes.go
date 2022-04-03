package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shoheiKU/golang_poker/pkg/config"
	"github.com/shoheiKU/golang_poker/pkg/handlers"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/poker", handlers.Repo.Poker)
	mux.Get("/getpotdata", handlers.Repo.BetsizeAjax)

	mux.Post("/initmobilepoker", handlers.Repo.InitMobilePoker)
	mux.Get("/mobilepoker", handlers.Repo.MobilePoker)
	mux.Post("/mobilepoker", handlers.Repo.MobilePokerBetPost)

	mux.Get("/contact", handlers.Repo.Contact)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	return mux
}
