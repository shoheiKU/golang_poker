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

	// home page
	mux.Get("/", handlers.Repo.Home)

	// about page
	mux.Get("/about", handlers.Repo.About)

	// remote poker api
	mux.Route("/remotepoker", func(r chi.Router) {
		r.Get("/", handlers.Repo.RemotePoker)
		r.Post("/init", handlers.Repo.RemotePokerInitPost)
		r.Get("/result", handlers.Repo.RemotePokerResult)
		r.Route("/action", func(r chi.Router) {
			r.Post("/check", handlers.Repo.RemotePokerCheckPost)

			r.Post("/call", handlers.Repo.RemotePokerCallPost)
			r.Post("/bet", handlers.Repo.RemotePokerBetPost)
			r.Post("/all-in", handlers.Repo.RemotePokerAllInPost)

			r.Post("/fold", handlers.Repo.RemotePokerFoldPost)
		})
		r.Get("/start", handlers.Repo.RemotePokerStartGame)
		r.Get("/reset", handlers.Repo.RemotePokerResetGame)
		r.Get("/next", handlers.Repo.RemotePokerNextGame)
	})

	// ajax related api
	mux.Route("/ajax", func(r chi.Router) {
		r.Get("/waitingphase", handlers.Repo.WaitingPhaseAjax)
		r.Get("/mobilewaitingphase", handlers.Repo.MobileWaitingPhaseAjax)
		r.Get("/waitingpokerdata", handlers.Repo.WaitingDataAjax)
	})

	// control api
	mux.Route("/control", func(r chi.Router) {
		r.Get("/", handlers.Repo.Control)
		r.Post("/reset", handlers.Repo.PokerRepoResetPost)
	})

	// control page
	mux.Get("/contact", handlers.Repo.Contact)

	// file server setting
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	return mux
}
