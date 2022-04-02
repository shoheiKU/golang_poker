package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/shoheiKU/web_application/GO_UDEMY/pkg/config"
	"github.com/shoheiKU/web_application/GO_UDEMY/pkg/handlers"
	"github.com/shoheiKU/web_application/GO_UDEMY/pkg/models"
	"github.com/shoheiKU/web_application/GO_UDEMY/pkg/render"
)

const ipAddr = "127.0.0.1"
const portNum = ":8081"
const mainTmplDir = "./templates/"

var app config.AppConfig
var session *scs.SessionManager

func main() {
	pokerRepo := handlers.NewPokerRepo(
		map[models.PlayerId]chan int{},
		[models.MaxPlayer]*models.PlayerData{},
		new(int),
		new(int),
		new(models.PlayerId),
		new(models.PlayerId),
	)
	// set up AppConfig & load templates
	// set true when in production
	app.InProduction = false
	app.UseCache = false
	app.TemplateCache = render.CreateTemplateCache(mainTmplDir)
	app.Session = scs.New()
	render.NewTemplates(&app)

	// set up the session
	session = app.Session
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	repo := handlers.NewRepo(&app, pokerRepo)
	handlers.NewHandlers(repo)

	srv := &http.Server{
		Addr:    ipAddr + portNum,
		Handler: routes(&app),
	}

	fmt.Println("Starting application. Port_number is", portNum)
	err := srv.ListenAndServe()
	if err != nil {
		fmt.Println("error starting server", err)
	}
}
