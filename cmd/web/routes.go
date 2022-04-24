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
	mux.Route("/poker", func(r chi.Router) {
		r.Get("/", handlers.Repo.Poker)
		r.Get("/start", handlers.Repo.PokerStartGame)
		r.Get("/reset", handlers.Repo.PokerResetGame)
		r.Get("/next", handlers.Repo.PokerNextGame)
		r.Get("/waitingphase", handlers.Repo.WaitingPhaseAjax)
		r.Get("/result", handlers.Repo.PokerResult)
	})
	mux.Get("/getpotdata", handlers.Repo.BetsizeAjax)

	mux.Route("/mobilepoker", func(r chi.Router) {
		r.Get("/", handlers.Repo.MobilePoker)
		r.Post("/init", handlers.Repo.MobilePokerInitPost)
		r.Get("/waitingturn", handlers.Repo.WaitingTurnAjax)
		r.Get("/pokerdata", handlers.Repo.WaitingDataAjax)
		r.Get("/result", handlers.Repo.MobilePokerResult)
		r.Route("/action", func(r chi.Router) {
			r.Post("/check", handlers.Repo.MobilePokerCheckPost)

			r.Post("/call", handlers.Repo.MobilePokerCallPost)
			r.Post("/bet", handlers.Repo.MobilePokerBetPost)
			r.Post("/all-in", handlers.Repo.MobilePokerAllInPost)

			r.Post("/fold", handlers.Repo.MobilePokerFoldPost)
		})
	})

	mux.Get("/contact", handlers.Repo.Contact)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	return mux
}
