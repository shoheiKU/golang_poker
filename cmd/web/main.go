package main

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/shoheiKU/golang_poker/pkg/config"
	"github.com/shoheiKU/golang_poker/pkg/handlers"
	"github.com/shoheiKU/golang_poker/pkg/poker"
	"github.com/shoheiKU/golang_poker/pkg/render"
)

//const ipAddr = "127.0.0.1"
const portNum = ":8081"
const mainTmplDir = "./templates/"

var app config.AppConfig
var session *scs.SessionManager

func main() {
	pokerRepo := handlers.NewPokerRepo()
	// set up AppConfig & load templates
	// set true when in production
	app.InProduction = false
	app.UseCache = false
	app.TemplateCache = render.CreateTemplateCache(mainTmplDir)
	app.Session = scs.New()
	render.NewTemplates(&app)

	// set up the session
	gob.Register(poker.Card{})
	session = app.Session
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	repo := handlers.NewRepo(&app, pokerRepo)
	handlers.NewHandlers(repo)

	srv := &http.Server{
		Addr:/*ipAddr + */ portNum,
		Handler: routes(&app),
	}

	fmt.Println("Starting application. Port_number is", portNum)
	err := srv.ListenAndServe()
	if err != nil {
		fmt.Println("error starting server", err)
	}
}
