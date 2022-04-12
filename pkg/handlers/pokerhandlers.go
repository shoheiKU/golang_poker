package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/shoheiKU/golang_poker/pkg/models"
	"github.com/shoheiKU/golang_poker/pkg/render"
)

// PokerRepository is the repository type for poker.
type PokerRepository struct {
	PlayersBet        map[models.PlayerSeat]chan int
	PlayersList       [models.MaxPlayer]*models.Player
	numOfActivePlayer *int
	Bet               *int
	Pot               *int
	OriginalRaiser    *models.PlayerSeat
	ButtonPlayer      *models.PlayerSeat
}

// NewPokerRepo make a PokerRepository
func NewPokerRepo(
	ch map[models.PlayerSeat]chan int,
	ls [models.MaxPlayer]*models.Player,
	numofplayer *int,
	bet *int,
	pot *int,
	originalRaiser *models.PlayerSeat,
	buttonPlayer *models.PlayerSeat,

) *PokerRepository {
	return &PokerRepository{
		PlayersBet:        ch,
		PlayersList:       ls,
		numOfActivePlayer: numofplayer,
		Bet:               bet,
		Pot:               pot,
		OriginalRaiser:    originalRaiser,
		ButtonPlayer:      buttonPlayer,
	}
}

// setPlayerInSession sets Player to Session.
func (m *Repository) setPlayerInSession(r *http.Request, player *models.Player) {
	m.App.Session.Put(r.Context(), "playerSeat", int(player.PlayerSeat()))
	m.App.Session.Put(r.Context(), "stack", player.Stack())
	m.App.Session.Put(r.Context(), "set", player.Bet())
	m.App.Session.Put(r.Context(), "isPlaying", player.IsPlaying())
}

// getPlayerFromSession gets Player from Session.
func (m *Repository) getPlayerFromSession(r *http.Request) *models.Player {
	seat := m.App.Session.GetInt(r.Context(), "playerSeat")
	stack := m.App.Session.GetInt(r.Context(), "stack")
	set := m.App.Session.GetInt(r.Context(), "set")
	isPlaying := m.App.Session.GetBool(r.Context(), "isPlaying")
	return models.NewPlayer(models.ItoPlayerSeat(seat), stack, set, isPlaying)
}

// betFunc handles bet and returns msg and bool.
func betFunc(player *models.Player) (msg string, ok bool) {
	ok = true
	if *Repo.PokerRepo.Bet > player.Bet() {
		// Bet less than the originalraiser's bet
		msg += fmt.Sprintf("You have to bet at least %d dollars.\n", *Repo.PokerRepo.Bet)
		ok = false
	} else if player.Bet() == player.Stack() {
		// All in
		msg += fmt.Sprintf("You all in %d dollars.\n", player.Bet())
		//player.Stack() -= player.Bet()
		*Repo.PokerRepo.Pot += player.Bet()
		if *Repo.PokerRepo.Bet < player.Bet() || Repo.PokerRepo.OriginalRaiser == nil {
			*Repo.PokerRepo.OriginalRaiser = player.PlayerSeat()
		}

	} else {
		msg = fmt.Sprintf("You bet %d dollars\n", player.Bet())
		//player.Stack() -= player.Bet()
		*Repo.PokerRepo.Pot += player.Bet()
		if *Repo.PokerRepo.Bet < player.Bet() || Repo.PokerRepo.OriginalRaiser == nil {
			*Repo.PokerRepo.OriginalRaiser = player.PlayerSeat()
		}
	}
	return
}

func isDeal(nextplayer models.PlayerSeat) bool {
	if *Repo.PokerRepo.OriginalRaiser == nextplayer {
		return true
	} else {
		return false
	}
}

func nextPlayer(s models.PlayerSeat) (next models.PlayerSeat) {
	next = s.NextSeat()
	if Repo.PokerRepo.PlayersList[next] == nil || Repo.PokerRepo.PlayersList[next].Stack() == 0 {
		return nextPlayer(next)
	}
	return
}

// Porker is the handler for the porker page.
func (m *Repository) Poker(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "poker.page.tmpl", &models.TemplateData{})
}

// InitMobilePoker is the handler to initialize the mobile porker page
func (m *Repository) InitMobilePoker(w http.ResponseWriter, r *http.Request) {
	expiration := time.Now()
	expiration = expiration.AddDate(0, 0, 1)
	log.Println(r.FormValue("PlayerSeat"))
	player := models.NewPlayer(
		models.AtoPlayerSeat(r.FormValue("PlayerSeat")),
		500,
		0,
		true,
	)
	m.PokerRepo.PlayersBet[player.PlayerSeat()] = make(chan int)
	m.PokerRepo.PlayersList[player.PlayerSeat()] = player
	*m.PokerRepo.numOfActivePlayer++
	m.setPlayerInSession(r, player)
	http.Redirect(w, r, "/mobilepoker", http.StatusFound)
}

// MobilePoker is the handler for the mobile porker page.
func (m *Repository) MobilePoker(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl",
		&models.TemplateData{IntMap: map[string]int{"stack": player.Stack()}})
}

// MobilePokerCallPost is the handler for Call.
func (m *Repository) MobilePokerCallPost(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	err := player.SetBet(*m.PokerRepo.Bet)
	if err != nil {
		// Can't bet correctly.
		render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &models.TemplateData{Error: "You should All in.\n"})
		return
	}
	td := models.TemplateData{}
	if msg, ok := betFunc(player); ok {
		// sucess message
		td.Flash = msg
		next := nextPlayer(player.PlayerSeat())
		log.Println(next.ToString())
		if isDeal(next) {
			// Next Phase.
		} else {
			Repo.PokerRepo.PlayersBet[next] <- player.Bet()
		}
	} else {
		// error message
		td.Error = msg
	}
	td.IntMap = map[string]int{"stack": player.Stack()}
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &td)
}

// MobilePokerBetPost is the handler for All in.
func (m *Repository) MobilePokerAllInPost(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	player.AllIn()
	td := models.TemplateData{}
	if msg, ok := betFunc(player); ok {
		// sucess message
		td.Flash = msg
		next := nextPlayer(player.PlayerSeat())
		log.Println(next.ToString())
		if isDeal(next) {
			// Next Phase.
		} else {
			Repo.PokerRepo.PlayersBet[next] <- player.Bet()
		}
	} else {
		// error message
		td.Error = msg
	}
	td.IntMap = map[string]int{"stack": player.Stack()}
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &td)
}

// MobilePokerBetPost is the handler for Bet.
func (m *Repository) MobilePokerBetPost(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	bet, _ := strconv.Atoi(r.FormValue("Bet"))
	err := player.SetBet(bet)
	if err != nil {
		render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &models.TemplateData{Error: "You can't bet more than your stack.\n"})
		return
	}
	td := models.TemplateData{}
	if msg, ok := betFunc(player); ok {
		// sucess message
		td.Flash = msg
		next := nextPlayer(player.PlayerSeat())
		log.Println(next.ToString())
		if isDeal(next) {
			// Next Phase.
		} else {
			Repo.PokerRepo.PlayersBet[next] <- player.Bet()
		}
	} else {
		// error message
		td.Error = msg
	}
	td.IntMap = map[string]int{"stack": player.Stack()}
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &td)
}

// MobilePokerFoldPost is the handler for Fold.
func (m *Repository) MobilePokerFoldPost(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	player.Fold()
	td := models.TemplateData{}
	td.Flash = "You folded.\n"
	td.IntMap = map[string]int{"stack": player.Stack()}
	next := nextPlayer(player.PlayerSeat())
	if isDeal(next) {
		// Next Phase.
	} else {
		Repo.PokerRepo.PlayersBet[next] <- player.Bet()
	}
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &td)
}

// MobilePokerCheckPost is the handler for Check.
func (m *Repository) MobilePokerCheckPost(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	player.Check()
	td := models.TemplateData{}
	td.Flash = "You checked.\n"
	td.IntMap = map[string]int{"stack": player.Stack()}
	next := nextPlayer(player.PlayerSeat())
	if isDeal(next) {
		// Next Phase.
	} else {
		Repo.PokerRepo.PlayersBet[next] <- player.Bet()
	}
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &td)
}
