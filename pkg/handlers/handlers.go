package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/shoheiKU/web_application/GO_UDEMY/pkg/config"
	"github.com/shoheiKU/web_application/GO_UDEMY/pkg/models"
	"github.com/shoheiKU/web_application/GO_UDEMY/pkg/render"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App       *config.AppConfig
	PokerRepo *PokerRepository
}

// NewRepo make a Repository
func NewRepo(
	a *config.AppConfig,
	p *PokerRepository,
) *Repository {
	return &Repository{
		App:       a,
		PokerRepo: p,
	}
}

type PokerRepository struct {
	PlayersBet     map[models.PlayerId]chan int
	PlayerList     [models.MaxPlayer]*models.PlayerData
	Bet            *int
	Pot            *int
	OriginalRaiser *models.PlayerId
	ButtonPlayer   *models.PlayerId
}

// NewPokerRepo make a PokerRepository
func NewPokerRepo(
	ch map[models.PlayerId]chan int,
	ls [models.MaxPlayer]*models.PlayerData,
	bet *int,
	pot *int,
	originalRaiser *models.PlayerId,
	buttonPlayer *models.PlayerId,

) *PokerRepository {
	return &PokerRepository{
		PlayersBet:     ch,
		PlayerList:     ls,
		Bet:            bet,
		Pot:            pot,
		OriginalRaiser: originalRaiser,
		ButtonPlayer:   buttonPlayer,
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(repo *Repository) {
	Repo = repo
}

// setPlayerDataInSession sets playerdata to http.ResponseWriter Cookie.
func (m *Repository) setPlayerDataInSession(r *http.Request, playerdata models.PlayerData) {
	v := reflect.ValueOf(playerdata)
	for i := 0; i < v.NumField(); i++ {
		key := v.Type().Field(i).Name
		val := fmt.Sprint(v.Field(i).Interface())
		m.App.Session.Put(r.Context(), key, val)
	}
}

// setPlayerCookies sets playerdata to http.ResponseWriter Cookie.
func (m *Repository) getPlayerDataFromSession(r *http.Request) (playerdata models.PlayerData) {
	v := reflect.ValueOf(&playerdata).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)
		indirect := reflect.Indirect(v.Field(i))
		fkey := f.Name
		ftype := f.Type
		isexist := m.App.Session.Exists(r.Context(), fkey)
		if isexist {
			temp := m.App.Session.Get(r.Context(), fkey).(string)
			str := reflect.ValueOf(&temp).Elem()
			val := reflect.Value{}
			switch ftype {
			case reflect.TypeOf(0):
				a, _ := strconv.Atoi(str.String())
				val = reflect.ValueOf(&a).Elem()
			case reflect.TypeOf(models.PlayerId(0)):
				i, _ := strconv.Atoi(str.String())
				a := models.ItoPlayerId(i)
				val = reflect.ValueOf(&a).Elem()
			default:
				log.Println(ftype)
			}
			indirect.Set(val)
		} else {
			return
		}

	}
	return
}
func betFunc(playerdata *models.PlayerData) (msg string, ok bool) {
	ok = true
	if *Repo.PokerRepo.Bet > playerdata.Bet {
		msg += fmt.Sprintf("You have to bet at least %d dollars.\n", *Repo.PokerRepo.Bet)
		ok = false
	}
	if playerdata.Bet > playerdata.Stack {
		msg += "You can't bet more than your stack.\n"
		ok = false
	}
	if ok {
		msg = fmt.Sprintf("You bet %d dollars", playerdata.Bet)
		playerdata.Stack -= playerdata.Bet
		if *Repo.PokerRepo.Bet < playerdata.Bet || Repo.PokerRepo.OriginalRaiser == nil {
			*Repo.PokerRepo.OriginalRaiser = playerdata.PlayerId
		}
	}
	return
}

func isDeal(id models.PlayerId) bool {
	if *Repo.PokerRepo.OriginalRaiser == id {
		return true
	} else {
		return false
	}
}

// Home is the handler for the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "about.page.tmpl", &models.TemplateData{})
}

// Porker is the handler for the porker page
func (m *Repository) Poker(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "poker.page.tmpl", &models.TemplateData{})
}

// BetsizeAjax is the handler for the betsize
func (m *Repository) BetsizeAjax(w http.ResponseWriter, r *http.Request) {
	betdata := map[string]int{}
	a := 0
	playerdata := m.getPlayerDataFromSession(r)
	go func() {
		fmt.Println("Waiting Scan")
		fmt.Scan(&a)
		m.PokerRepo.PlayersBet[playerdata.PlayerId] <- a
	}()
	betdata["betsize"] = <-m.PokerRepo.PlayersBet[playerdata.PlayerId]
	betdata["potsize"] = 100

	betdataJson, err := json.Marshal(betdata)
	if err != nil {
		log.Println(err)
	}
	w.Write(betdataJson)
}

// InitMobilePoker is the handler to initialize the mobile porker page
func (m *Repository) InitMobilePoker(w http.ResponseWriter, r *http.Request) {
	expiration := time.Now()
	expiration = expiration.AddDate(0, 0, 1)
	playerdata := models.PlayerData{
		PlayerId: models.AtoPlayerId(r.FormValue("PlayerId")),
		Stack:    500,
		Bet:      0,
	}
	m.PokerRepo.PlayersBet[playerdata.PlayerId] = make(chan int)
	m.setPlayerDataInSession(r, playerdata)
	http.Redirect(w, r, "/mobilepoker", http.StatusFound)
}

// MobilePoker is the handler for the mobile porker page
func (m *Repository) MobilePoker(w http.ResponseWriter, r *http.Request) {
	if len(r.Cookies()) <= 1 {
		render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &models.TemplateData{})
		return
	}
	playerdata := m.getPlayerDataFromSession(r)
	m.setPlayerDataInSession(r, playerdata)
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl",
		&models.TemplateData{IntMap: map[string]int{"stack": playerdata.Stack}})
}

// MobilePokerBetPost is the handler for the bet.
func (m *Repository) MobilePokerBetPost(w http.ResponseWriter, r *http.Request) {
	bet, _ := strconv.Atoi(r.FormValue("Bet"))
	playerdata := m.getPlayerDataFromSession(r)
	playerdata.Bet = bet
	td := models.TemplateData{}
	if msg, ok := betFunc(&playerdata); ok {
		// sucess message
		td.Flash = msg
	} else {
		// error message
		td.Error = msg
	}
	td.IntMap = map[string]int{"stack": playerdata.Stack}
	m.setPlayerDataInSession(r, playerdata)
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &td)
}

// Contact is the handler for the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.tmpl", &models.TemplateData{})
}
