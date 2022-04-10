package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/shoheiKU/golang_poker/pkg/config"
	"github.com/shoheiKU/golang_poker/pkg/models"
	"github.com/shoheiKU/golang_poker/pkg/render"
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

// PokerRepository is the repository type for poker
type PokerRepository struct {
	PlayersBet        map[models.PlayerSeat]chan int
	PlayersList       [models.MaxPlayer]*models.PlayerData
	NumOfActivePlayer *int
	Bet               *int
	Pot               *int
	OriginalRaiser    *models.PlayerSeat
	ButtonPlayer      *models.PlayerSeat
}

// NewPokerRepo make a PokerRepository
func NewPokerRepo(
	ch map[models.PlayerSeat]chan int,
	ls [models.MaxPlayer]*models.PlayerData,
	numofplayer *int,
	bet *int,
	pot *int,
	originalRaiser *models.PlayerSeat,
	buttonPlayer *models.PlayerSeat,

) *PokerRepository {
	return &PokerRepository{
		PlayersBet:        ch,
		PlayersList:       ls,
		NumOfActivePlayer: numofplayer,
		Bet:               bet,
		Pot:               pot,
		OriginalRaiser:    originalRaiser,
		ButtonPlayer:      buttonPlayer,
	}
}

// NewHandlers sets the repository for the handlers.
func NewHandlers(repo *Repository) {
	Repo = repo
}

// setPlayerDataInSession sets playerdata to Session.
func (m *Repository) setPlayerDataInSession(r *http.Request, playerdata models.PlayerData) {
	v := reflect.ValueOf(playerdata)
	for i := 0; i < v.NumField(); i++ {
		key := v.Type().Field(i).Name
		val := fmt.Sprint(v.Field(i).Interface())
		m.App.Session.Put(r.Context(), key, val)
	}
}

// getPlayerDataFromSession gets playerdata from Session.
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
			case reflect.TypeOf(models.PlayerSeat(0)):
				i, _ := strconv.Atoi(str.String())
				a := models.ItoPlayerSeat(i)
				val = reflect.ValueOf(&a).Elem()
			case reflect.TypeOf(true):
				a, _ := strconv.ParseBool(str.String())
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

// betFunc handles bet and returns msg and bool.
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
		*Repo.PokerRepo.Pot += playerdata.Bet
		if *Repo.PokerRepo.Bet < playerdata.Bet || Repo.PokerRepo.OriginalRaiser == nil {
			*Repo.PokerRepo.OriginalRaiser = playerdata.PlayerSeat
		}
		next := nextPlayer(playerdata.PlayerSeat)
		log.Println(next.ToString())
		if isDeal(next) {
			// Next Phase.
		} else {
			Repo.PokerRepo.PlayersBet[next] <- playerdata.Bet
		}
	}
	return
}

func isDeal(s models.PlayerSeat) bool {
	if *Repo.PokerRepo.OriginalRaiser == s {
		return true
	} else {
		return false
	}
}

func nextPlayer(s models.PlayerSeat) (next models.PlayerSeat) {
	next = s.NextSeat()
	if Repo.PokerRepo.PlayersList[next] == nil {
		return nextPlayer(next)
	}
	return
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
		m.PokerRepo.PlayersBet[playerdata.PlayerSeat] <- a
	}()
	betdata["betsize"] = <-m.PokerRepo.PlayersBet[playerdata.PlayerSeat]
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
		PlayerSeat: models.AtoPlayerSeat(r.FormValue("PlayerSeat")),
		Stack:      500,
		Bet:        0,
		IsPlaying:  true,
	}
	m.PokerRepo.PlayersBet[playerdata.PlayerSeat] = make(chan int)
	m.PokerRepo.PlayersList[playerdata.PlayerSeat] = &playerdata
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
func (m *Repository) WaitingTurnAjax(w http.ResponseWriter, r *http.Request) {
	betdata := map[string]interface{}{}
	playerdata := m.getPlayerDataFromSession(r)
	select {
	// request.Context is cancelled.
	case <-r.Context().Done():
		return
	// Get a data from the former player.
	case betsize := <-m.PokerRepo.PlayersBet[playerdata.PlayerSeat]:
		betdata["BetSize"] = betsize
		originalraiser := *m.PokerRepo.OriginalRaiser
		log.Println(betsize)
		betdata["OriginalRaiser"] = originalraiser.ToString()

		// Write a json as a return to ajax.
		betdataJson, err := json.Marshal(betdata)
		if err != nil {
			log.Println(err)
		}
		w.Write(betdataJson)
	}
}

// Contact is the handler for the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.tmpl", &models.TemplateData{})
}
