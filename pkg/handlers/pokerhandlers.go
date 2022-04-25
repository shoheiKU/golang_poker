package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/shoheiKU/golang_poker/pkg/models"
	"github.com/shoheiKU/golang_poker/pkg/poker"
	"github.com/shoheiKU/golang_poker/pkg/render"
)

// PokerRepository is the repository type for poker.
type PokerRepository struct {
	PlayersCh         [models.MaxPlayer]chan int
	PlayersList       [models.MaxPlayer]*models.Player
	Winners           []*models.Player
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
		Winners:           []*models.Player{},
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
	if m.PokerRepo.Phase == 4 {
		num := len(m.PokerRepo.Winners)
		pot := m.PokerRepo.Pot / num
		remainder := m.PokerRepo.Pot % num
		for _, player := range m.PokerRepo.Winners {
			if player == m.PokerRepo.Winners[0] {
				player.SetWinPot(pot + remainder)
			} else {
				player.SetWinPot(pot)
			}
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

// nextPhase sends the number of phase to PhaseCh.
func (m *Repository) nextPhase() {
	m.PokerRepo.Phase = (m.PokerRepo.Phase + 1) % 5
	if m.PokerRepo.numOfActivePlayer == 1 {
		// Result Phase(4)
		m.PokerRepo.Phase = 4
	}
	phase := m.PokerRepo.Phase
	log.Println("send data-----------------------------------------------------------")
	m.PokerRepo.PhaseCh <- phase
	log.Println("Phase :", m.PokerRepo.Phase)
	// Set Original Raiser
	first := m.nextPlayer(m.PokerRepo.ButtonPlayer)
	m.PokerRepo.OriginalRaiser = first
	// send phase data to each cliant
	m.playerPhaseChange(phase)
	m.PokerRepo.Bet = 0
}

// nextPlayer returns a next player.
func (m *Repository) nextPlayer(s models.PlayerSeat) models.PlayerSeat {
	// If the number of active players is less than 2, return Preset Player.
	if m.PokerRepo.numOfActivePlayer <= 1 {
		return models.PresetPlayer
	}
	next := s.NextSeat()
	if m.PokerRepo.PlayersList[next] == nil || !m.PokerRepo.PlayersList[next].IsPlaying() || m.PokerRepo.PlayersList[next].Stack() == 0 {
		return m.nextPlayer(next)
	}
	return next
}

// playHand returns a slice of pointers to models.Player who won the game.
func (m *Repository) setWinners() {
	m.PokerRepo.Winners = m.playHand()
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
	sort.Slice(winners, func(i, j int) bool {
		si := (winners[i].PlayerSeat() + models.MaxPlayer) % models.MaxPlayer
		sj := (winners[j].PlayerSeat() + models.MaxPlayer) % models.MaxPlayer
		return si < sj
	})
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
		h1num := h1.Cards[i].Num
		h2num := h2.Cards[i].Num
		if h1num == 1 {
			h1num += 13
		}
		if h2num == 1 {
			h2num += 13
		}
		if h1num < h2num {
			return p2
		} else if h1num > h2num {
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

// playerChange sends each player the next player data with goroutine.
func (m *Repository) playerChange(next models.PlayerSeat) {
	for _, p := range m.PokerRepo.PlayersList {
		if p == nil {
			continue
		}
		go func(player *models.Player) {
			m.PokerRepo.DecisionMakerCh[player.PlayerSeat()] <- next
		}(p)
	}
}

// playerPhaseChange sends each player the number of phase.
func (m *Repository) playerPhaseChange(phase int) {
	for _, p := range m.PokerRepo.PlayersList {
		if p == nil {
			continue
		}
		go func(player *models.Player) {
			m.PokerRepo.PlayersCh[player.PlayerSeat()] <- phase
		}(p)
	}
}

//----------------------------------------- Get Handlers-----------------------------------------//
// MobilePoker is the handler for the mobile poker page.
func (m *Repository) MobilePoker(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	if player == nil {
		player = models.NewPlayer(models.PresetPlayer, 0, 0, false, &[2]poker.Card{})
	}
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &models.TemplateData{Data: player.PlayerTemplateData()})
}

//.
//└── data
//    ├── bet
//    │   └── int
//    ├── stack
//    │   └── int
//    ├── playerSeat
//    │   └── string
//    ├── isPlaying
//    │   └── bool
//    ├── pocketCards
//    │   ├── card1
//    │   │   └── poker.Card
//    │   └── card2
//    │       └── poker.Card
//    ├── hand
//    │   └── poker.HandTemplateData
//    ├── communityCards
//    │   └── [5]poker.Card
//    ├── winners
//    │   └── []struct{}
//    │       ├── PlayerSeat
//    │       │   └── string
//    │       ├── WinPot
//    │       │   └── int
//    │       └── Hand
//    │           └── poker.HandTemplateData
//    └── showdowns
//        └── []struct{}
//            ├── PlayerSeat
//            │   └── string
//            └── Hand
//                └── poker.HandTemplateData

// MobilePokerResult is the handler for the mobile result page.
func (m *Repository) MobilePokerResult(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	data := map[string]interface{}{}
	for key, val := range player.PlayerTemplateData() {
		data[key] = val
	}
	data["hand"] = player.HandTemplateData()
	data["communityCards"] = m.PokerRepo.CommunityCards

	winnersstructs := make([]struct {
		PlayerSeat string
		WinPot     int
		Hand       *poker.HandTemplateData
	}, 0, len(m.PokerRepo.Winners))
	showdowns := make([]struct {
		PlayerSeat string
		Hand       *poker.HandTemplateData
	}, 0)

	for _, p := range m.PokerRepo.Winners {
		winnersstructs = append(winnersstructs, struct {
			PlayerSeat string
			WinPot     int
			Hand       *poker.HandTemplateData
		}{PlayerSeat: p.PlayerSeat().ToString(), WinPot: p.WinPot(), Hand: p.HandTemplateData()})
	}
	data["winners"] = winnersstructs

	for _, p := range m.PokerRepo.PlayersList {
		if p == nil || !p.IsPlaying() {
			continue
		}
		showdowns = append(showdowns, struct {
			PlayerSeat string
			Hand       *poker.HandTemplateData
		}{PlayerSeat: p.PlayerSeat().ToString(), Hand: p.HandTemplateData()})

	}
	data["showdowns"] = showdowns
	render.RenderTemplate(w, r, "mobile_poker_result.page.tmpl",
		&models.TemplateData{Data: data})
}

// Porker is the handler for the porker page.
func (m *Repository) Poker(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "poker.page.tmpl", &models.TemplateData{})
}

// PokerStartGame is the handler to start a new game.
func (m *Repository) PokerStartGame(w http.ResponseWriter, r *http.Request) {
	poker.Deck.Reset()
	m.PokerRepo.Phase = -1
	m.PokerRepo.ButtonPlayer = models.PlayerSeat(models.PresetPlayer)
	for _, player := range m.PokerRepo.PlayersList {
		// skip nil player
		if player == nil {
			continue
		}

		// set ButtonPlayer
		if m.PokerRepo.ButtonPlayer == models.PresetPlayer {
			m.PokerRepo.ButtonPlayer = player.PlayerSeat()
		}

		player.Reset()
		// reset each page
		log.Println("Before-------------------------------------")
		m.PokerRepo.PlayersCh[player.PlayerSeat()] <- -1
		log.Println("After-------------------------------------")
	}
	// Blind Bet
	m.blindBet()
	utg := m.PokerRepo.ButtonPlayer
	// wait for redirect
	time.Sleep(time.Second)
	for i := 0; i < 3; i++ {
		utg = m.nextPlayer(utg)
	}
	m.nextPhase()
	m.playerChange(utg)
	render.RenderTemplate(w, r, "poker.page.tmpl", &models.TemplateData{Info: "Prefrop Phase"})
}

// PokerResetGame is the handler to reset the existing game.
func (m *Repository) PokerResetGame(w http.ResponseWriter, r *http.Request) {
	poker.Deck.Reset()
	m.PokerRepo.Phase = 0
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
	// wait for redirect
	time.Sleep(time.Second)
	for i := 0; i < 3; i++ {
		utg = m.nextPlayer(utg)
	}
	m.playerChange(utg)
	render.RenderTemplate(w, r, "poker.page.tmpl", &models.TemplateData{})
}

// PokerResult is the handler for the result page.
func (m *Repository) PokerResult(w http.ResponseWriter, r *http.Request) {
	player := models.NewPlayer(0, 500, 0, true, &[2]poker.Card{})
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "poker_result.page.tmpl",
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
		td.NotieFlash = msg
		next := m.nextPlayer(player.PlayerSeat())
		if m.isDeal(next) {
			// Next Phase
			m.nextPhase()
			if m.PokerRepo.Phase == 4 {
				m.setWinners()
			} else {
				// send poker data to each cliant
				// Next player of the button player is going to play.
				m.playerChange(m.nextPlayer(m.PokerRepo.ButtonPlayer))
			}
			m.deal()
		} else {
			m.playerChange(next)
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
		td.NotieFlash = msg
		next := m.nextPlayer(player.PlayerSeat())
		if m.isDeal(next) {
			// Next Phase.
			m.nextPhase()
			if m.PokerRepo.Phase == 4 {
				m.setWinners()
			} else {
				// send poker data to each cliant
				// Next player of the button player is going to play.
				m.playerChange(m.nextPlayer(m.PokerRepo.ButtonPlayer))
			}
			m.deal()
		} else {
			m.playerChange(next)
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
		m.playerChange(player.PlayerSeat())
		render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &models.TemplateData{Error: "You can't bet more than your stack.\n"})
		return
	}
	td := models.TemplateData{}
	if msg, ok := m.betFunc(player); ok {
		// sucess message
		td.NotieFlash = msg
		next := m.nextPlayer(player.PlayerSeat())
		if m.isDeal(next) {
			// Next Phase.
			m.nextPhase()
			if m.PokerRepo.Phase == 4 {
				m.setWinners()
			} else {
				// send poker data to each cliant
				// Next player of the button player is going to play.
				m.playerChange(m.nextPlayer(m.PokerRepo.ButtonPlayer))
			}
			m.deal()
		} else {
			m.playerChange(next)
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
	td.NotieFlash = "You folded.\n"
	td.Data = player.PlayerTemplateData()
	next := m.nextPlayer(player.PlayerSeat())
	if m.isDeal(next) {
		// Next Phase.
		m.nextPhase()
		if m.PokerRepo.Phase == 4 {
			m.setWinners()
		} else {
			// send poker data to each cliant
			// Next player of the button player is going to play.
			m.playerChange(m.nextPlayer(m.PokerRepo.ButtonPlayer))
		}
		m.deal()
	} else {
		m.playerChange(next)
	}
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &td)
}

// MobilePokerCheckPost is the handler for Check.
func (m *Repository) MobilePokerCheckPost(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	player.Check()
	td := models.TemplateData{}
	td.NotieFlash = "You checked.\n"
	td.Data = player.PlayerTemplateData()
	next := m.nextPlayer(player.PlayerSeat())
	if m.isDeal(next) {
		// Next Phase.
		m.nextPhase()
		if m.PokerRepo.Phase == 4 {
			m.setWinners()
		} else {
			// send poker data to each cliant
			// Next player of the button player is going to play.
			m.playerChange(m.nextPlayer(m.PokerRepo.ButtonPlayer))
		}
		m.deal()
	} else {
		m.playerChange(next)
	}
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &td)
}
