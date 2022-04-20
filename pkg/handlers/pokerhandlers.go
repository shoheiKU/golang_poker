package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/shoheiKU/golang_poker/pkg/models"
	"github.com/shoheiKU/golang_poker/pkg/poker"
	"github.com/shoheiKU/golang_poker/pkg/render"
)

// PokerRepository is the repository type for poker.
type PokerRepository struct {
	PlayersCh         [models.MaxPlayer]chan int
	PlayersList       [models.MaxPlayer]*models.Player
	ButtonPlayer      models.PlayerSeat
	DecisionMakerCh   [models.MaxPlayer]chan models.PlayerSeat
	Phase             int
	PhaseCh           chan int
	CommunityCards    [5]poker.Card
	numOfActivePlayer int
	SB                int
	BB                int
	Pot               int
	Bet               int
	OriginalRaiser    models.PlayerSeat
}

// NewPokerRepo make a PokerRepository
func NewPokerRepo() *PokerRepository {
	return &PokerRepository{
		PlayersCh:         [models.MaxPlayer]chan int{},
		PlayersList:       [models.MaxPlayer]*models.Player{},
		ButtonPlayer:      models.PlayerSeat(models.PresetPlayer),
		DecisionMakerCh:   [models.MaxPlayer]chan models.PlayerSeat{},
		Phase:             0,
		PhaseCh:           make(chan int),
		CommunityCards:    [5]poker.Card{},
		numOfActivePlayer: 0,
		SB:                50,
		BB:                100,
		Pot:               0,
		Bet:               0,
		OriginalRaiser:    models.PlayerSeat(models.PresetPlayer),
	}
}

//----------------------------------------- Functions -----------------------------------------//
// setPlayerInSession sets Player to Session.
func (m *Repository) setPlayerInSession(r *http.Request, player *models.Player) {
	m.App.Session.Put(r.Context(), "playerSeat", int(player.PlayerSeat()))
}

// getPlayerFromSession gets Player from Session.
func (m *Repository) getPlayerFromSession(r *http.Request) *models.Player {
	ok := m.App.Session.Exists(r.Context(), "playerSeat")
	if ok {
		seat := m.App.Session.GetInt(r.Context(), "playerSeat")
		return m.PokerRepo.PlayersList[seat]
	} else {
		return nil
	}
}

// deal is a function for dealing.
func (m *Repository) deal() {
	for _, p := range m.PokerRepo.PlayersList {
		if p == nil {
			continue
		} else {
			p.Deal()
		}
	}
}

// isDeal returns whether or not all plaayers bet same amount of chips as a bool.
func (m *Repository) isDeal(nextplayer models.PlayerSeat) bool {
	log.Println("Original Raiser:", m.PokerRepo.OriginalRaiser.ToString())
	log.Println("Next Player:", nextplayer.ToString())
	if m.PokerRepo.OriginalRaiser == nextplayer {
		log.Println("isDeal is True")
		return true
	} else {
		log.Println("isDeal is False")
		return false
	}
}

// nextPhase sends Phase to PhaseCh.
func (m *Repository) nextPhase() {
	m.PokerRepo.PhaseCh <- m.PokerRepo.Phase
	m.PokerRepo.Phase = (m.PokerRepo.Phase + 1) % 3
	log.Println("Phase :", m.PokerRepo.Phase)
	first := m.nextPlayer(m.PokerRepo.ButtonPlayer)
	m.PokerRepo.OriginalRaiser = first
	m.playerChange(first)
	m.PokerRepo.Bet = 0
}

// nextPlayer returns a next player.
func (m *Repository) nextPlayer(s models.PlayerSeat) models.PlayerSeat {
	next := s.NextSeat()
	if m.PokerRepo.PlayersList[next] == nil || !m.PokerRepo.PlayersList[next].IsPlaying() || m.PokerRepo.PlayersList[next].Stack() == 0 {
		return m.nextPlayer(next)
	}
	return next
}

// playHand returns a slice of pointers to models.Player who won the game.
func (m *Repository) playHand() []*models.Player {
	var winners []*models.Player
	var winner *models.Player
	for _, player := range m.PokerRepo.PlayersList {
		if player == nil {
			continue
		} else {
			player.SetHand(&m.PokerRepo.CommunityCards)
			winner = m.compareHands(winner, player)
			if winner == nil {
				// Tie
				winners = append(winners, player)
				winner = player
			} else {
				winners = []*models.Player{winner}
			}
		}
	}
	return winners
}

// compareHands returns a pointer to winner models.Player.
func (m *Repository) compareHands(p1, p2 *models.Player) *models.Player {
	// When p1 or p2 is nil, this function returns the other player.
	if p1 == nil {
		return p2
	}
	if p2 == nil {
		return p1
	}
	// Compare players hands
	h1 := p1.Hand()
	h2 := p2.Hand()
	if h1.Val < h2.Val {
		return p2
	}
	if h1.Val > h2.Val {
		return p1
	}

	// Compare cards' values
	for i := 0; i < 5; i++ {
		if h1.Cards[i].Num < h2.Cards[i].Num {
			return p2
		} else if h1.Cards[i].Num > h2.Cards[i].Num {
			return p1
		}
	}

	// Draw
	return nil
}

// betFunc handles bet and returns msg and bool.
func (m *Repository) betFunc(player *models.Player) (msg string, ok bool) {
	ok = true
	if m.PokerRepo.Bet > player.Bet() {
		// Bet less than the originalraiser's bet
		msg += fmt.Sprintf("You have to bet at least %d dollars.\n", m.PokerRepo.Bet)
		ok = false
	} else if player.Bet() == player.Stack() {
		// All in
		msg += fmt.Sprintf("You all in %d dollars.\n", player.Bet())
		m.PokerRepo.Pot += player.Bet()
		if m.PokerRepo.Bet < player.Bet() || m.PokerRepo.OriginalRaiser == models.PresetPlayer {
			m.PokerRepo.Bet = player.Bet()
			m.PokerRepo.OriginalRaiser = player.PlayerSeat()
		}

	} else {
		// Bet
		msg = fmt.Sprintf("You bet %d dollars\n", player.Bet())
		m.PokerRepo.Pot += player.Bet()
		if m.PokerRepo.Bet < player.Bet() || m.PokerRepo.OriginalRaiser == models.PresetPlayer {
			m.PokerRepo.Bet = player.Bet()
			m.PokerRepo.OriginalRaiser = player.PlayerSeat()
		}
	}
	return
}

// blindBet bets small blind and big blind.
func (m *Repository) blindBet() {
	btn := m.PokerRepo.ButtonPlayer
	sbp := m.nextPlayer(btn)
	bbp := m.nextPlayer(sbp)
	m.PokerRepo.PlayersList[sbp].SetBet(m.PokerRepo.SB)
	m.PokerRepo.PlayersList[bbp].SetBet(m.PokerRepo.BB)
	m.PokerRepo.OriginalRaiser = models.PresetPlayer
	m.PokerRepo.Bet = m.PokerRepo.BB
}

func (m *Repository) playerChange(next models.PlayerSeat) {
	for _, p := range m.PokerRepo.PlayersList {
		if p == nil {
			continue
		}
		m.PokerRepo.DecisionMakerCh[p.PlayerSeat()] <- next
	}
}

//----------------------------------------- Get Handlers-----------------------------------------//
// MobilePoker is the handler for the mobile poker page.
func (m *Repository) MobilePoker(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	if player == nil {
		player = models.NewPlayer(models.PresetPlayer, 0, 0, false, &[2]poker.Card{})
	}
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl",
		&models.TemplateData{Data: player.PlayerTemplateData()})
}

// MobilePokerStart is the handler to start a new game.
func (m *Repository) MobilePokerStartGame(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	player.Reset()
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl",
		&models.TemplateData{Data: player.PlayerTemplateData()})
}

// Porker is the handler for the porker page.
func (m *Repository) Poker(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "poker.page.tmpl", &models.TemplateData{})
}

// PokerStartGame is the handler to start a new game.
func (m *Repository) PokerStartGame(w http.ResponseWriter, r *http.Request) {
	poker.Deck.Reset()
	m.PokerRepo.Phase = 0
	p := models.PlayerSeat(models.PresetPlayer)
	for _, player := range m.PokerRepo.PlayersList {
		// skip nil player
		if player == nil {
			continue
		}

		if p == models.PresetPlayer {
			p = player.PlayerSeat()
		}
		player.Reset()
	}
	m.PokerRepo.ButtonPlayer = p
	m.blindBet()
	utg := m.PokerRepo.ButtonPlayer
	for i := 0; i < 3; i++ {
		utg = m.nextPlayer(utg)
	}
	m.playerChange(utg)
	m.PokerRepo.PlayersCh[utg] <- m.PokerRepo.BB
	render.RenderTemplate(w, r, "poker.page.tmpl", &models.TemplateData{})
}

// PokerResetGame is the handler to reset the existing game.
func (m *Repository) PokerResetGame(w http.ResponseWriter, r *http.Request) {
	poker.Deck.Reset()
	m.PokerRepo.Phase = 0
	var p models.PlayerSeat
	for _, player := range m.PokerRepo.PlayersList {
		if player == nil {
			continue
		} else {
			p = player.PlayerSeat()
			break
		}
	}
	m.PokerRepo.ButtonPlayer = p
	for _, ch := range m.PokerRepo.PlayersCh {
		if ch == nil {
			continue
		} else {
			// Send data to WaitingTurnAjax.
			// JS calls "/mobilepoker/start" -> (func MobilepokerStartGame).
			ch <- -1
		}
	}
	utg := m.PokerRepo.ButtonPlayer
	for i := 0; i < 3; i++ {
		utg = m.nextPlayer(utg)
	}
	m.playerChange(utg)
	m.PokerRepo.PlayersCh[utg] <- m.PokerRepo.BB
	render.RenderTemplate(w, r, "poker.page.tmpl", &models.TemplateData{})
}

// PokerNextGame is the handler to proceed to next game.
func (m *Repository) PokerNextGame(w http.ResponseWriter, r *http.Request) {
	poker.Deck.Reset()
	m.PokerRepo.Phase = 0
	m.PokerRepo.ButtonPlayer = m.nextPlayer(m.PokerRepo.ButtonPlayer)
	for _, ch := range m.PokerRepo.PlayersCh {
		if ch == nil {
			continue
		} else {
			ch <- -1
		}
	}
	m.blindBet()
	utg := m.PokerRepo.ButtonPlayer
	for i := 0; i < 3; i++ {
		utg = m.nextPlayer(utg)
	}
	m.playerChange(utg)
	m.PokerRepo.PlayersCh[utg] <- m.PokerRepo.BB
	render.RenderTemplate(w, r, "poker.page.tmpl", &models.TemplateData{})
}

// PokerResult is the handler for the result page.
func (m *Repository) PokerResult(w http.ResponseWriter, r *http.Request) {
	player := models.NewPlayer(0, 500, 0, true, &[2]poker.Card{})
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "poker_result.page.tmpl",
		&models.TemplateData{Data: player.PlayerTemplateData()})
}

// ServeCards is the handler for serving player cards.
func (m *Repository) ServeCards(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	player.SetPocketCards()
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl",
		&models.TemplateData{Data: player.PlayerTemplateData()})
}

//----------------------------------------- Post Handlers-----------------------------------------//
// MobilePokerInitPost is the handler to initialize the mobile porker page
func (m *Repository) MobilePokerInitPost(w http.ResponseWriter, r *http.Request) {
	player := models.NewPlayer(
		models.AtoPlayerSeat(r.FormValue("PlayerSeat")),
		500,
		0,
		true,
		&[2]poker.Card{},
	)
	m.PokerRepo.PlayersCh[player.PlayerSeat()] = make(chan int)
	m.PokerRepo.PlayersList[player.PlayerSeat()] = player
	m.PokerRepo.DecisionMakerCh[player.PlayerSeat()] = make(chan models.PlayerSeat)
	if m.PokerRepo.ButtonPlayer == models.PresetPlayer {
		m.PokerRepo.ButtonPlayer = player.PlayerSeat()
	}
	m.PokerRepo.numOfActivePlayer++
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl",
		&models.TemplateData{Data: player.PlayerTemplateData()})
}

// MobilePokerCallPost is the handler for Call.
func (m *Repository) MobilePokerCallPost(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	err := player.SetBet(m.PokerRepo.Bet)
	if err != nil {
		// Can't bet correctly.
		render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &models.TemplateData{Error: "You should All in.\n"})
		return
	}
	td := models.TemplateData{}
	if msg, ok := m.betFunc(player); ok {
		// sucess message
		td.Flash = msg
		next := m.nextPlayer(player.PlayerSeat())
		if m.isDeal(next) {
			// Next Phase.
			m.deal()
			m.nextPhase()
			m.PokerRepo.PlayersCh[m.nextPlayer(m.PokerRepo.ButtonPlayer)] <- m.PokerRepo.Bet
		} else {
			m.playerChange(next)
			Repo.PokerRepo.PlayersCh[next] <- player.Bet()
		}
	} else {
		// error message
		td.Error = msg
	}
	td.Data = player.PlayerTemplateData()
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &td)
}

// MobilePokerBetPost is the handler for All in.
func (m *Repository) MobilePokerAllInPost(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	player.AllIn()
	td := models.TemplateData{}
	if msg, ok := m.betFunc(player); ok {
		// sucess message
		td.Flash = msg
		next := m.nextPlayer(player.PlayerSeat())
		if m.isDeal(next) {
			// Next Phase.
			m.deal()
			m.nextPhase()
			m.PokerRepo.PlayersCh[m.nextPlayer(m.PokerRepo.ButtonPlayer)] <- m.PokerRepo.Bet
		} else {
			m.playerChange(next)
			Repo.PokerRepo.PlayersCh[next] <- player.Bet()
		}
	} else {
		// error message
		td.Error = msg
	}
	td.Data = player.PlayerTemplateData()
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
	if msg, ok := m.betFunc(player); ok {
		// sucess message
		td.Flash = msg
		next := m.nextPlayer(player.PlayerSeat())
		if m.isDeal(next) {
			// Next Phase.
			m.deal()
			m.nextPhase()
			m.PokerRepo.PlayersCh[m.nextPlayer(m.PokerRepo.ButtonPlayer)] <- m.PokerRepo.Bet
		} else {
			m.playerChange(next)
			Repo.PokerRepo.PlayersCh[next] <- player.Bet()
		}
	} else {
		// error message
		td.Error = msg
	}
	td.Data = player.PlayerTemplateData()
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &td)
}

// MobilePokerFoldPost is the handler for Fold.
func (m *Repository) MobilePokerFoldPost(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	player.Fold()
	td := models.TemplateData{}
	td.Flash = "You folded.\n"
	td.Data = player.PlayerTemplateData()
	next := m.nextPlayer(player.PlayerSeat())
	if m.isDeal(next) {
		// Next Phase.
		m.deal()
		m.nextPhase()
		m.PokerRepo.PlayersCh[m.nextPlayer(m.PokerRepo.ButtonPlayer)] <- m.PokerRepo.Bet
	} else {
		m.playerChange(next)
		Repo.PokerRepo.PlayersCh[next] <- player.Bet()
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
	td.Data = player.PlayerTemplateData()
	next := m.nextPlayer(player.PlayerSeat())
	if m.isDeal(next) {
		// Next Phase.
		m.deal()
		m.nextPhase()
		m.PokerRepo.PlayersCh[m.nextPlayer(m.PokerRepo.ButtonPlayer)] <- m.PokerRepo.Bet
	} else {
		m.playerChange(next)
		Repo.PokerRepo.PlayersCh[next] <- player.Bet()
	}
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &td)
}
