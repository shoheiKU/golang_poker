package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/shoheiKU/golang_poker/pkg/models"
	"github.com/shoheiKU/golang_poker/pkg/poker"
	"github.com/shoheiKU/golang_poker/pkg/render"
)

var PhaseString = [5]string{
	"Prefrop Betting Phase",
	"Frop Betting Phase",
	"Turn Betting Phase",
	"River Betting Phase",
	"Show Down",
}

// PokerRepository is the repository type for poker.
type PokerRepository struct {
	PlayersCh         [models.MaxPlayer]chan int
	PlayersList       [models.MaxPlayer]*models.Player
	Winners           []*models.Player
	ButtonPlayer      *models.PlayerSeat
	DecisionMaker     *models.PlayerSeat
	DecisionMakerCh   [models.MaxPlayer]chan models.PlayerSeat
	Phase             *int
	PhaseCh           chan int
	CommunityCards    [5]poker.Card
	numOfActivePlayer *int
	SB                int
	BB                int
	Pot               *int
	Bet               *int
	OriginalRaiser    *models.PlayerSeat
}

// NewPokerRepo make a PokerRepository
func NewPokerRepo() *PokerRepository {
	button := models.PlayerSeat(models.PresetPlayer)
	decisionmaker := models.PlayerSeat(models.PresetPlayer)
	originalraiser := models.PlayerSeat(models.PresetPlayer)
	return &PokerRepository{
		PlayersCh:         [models.MaxPlayer]chan int{},
		PlayersList:       [models.MaxPlayer]*models.Player{},
		Winners:           []*models.Player{},
		ButtonPlayer:      &button,
		DecisionMaker:     &decisionmaker,
		DecisionMakerCh:   [models.MaxPlayer]chan models.PlayerSeat{},
		Phase:             new(int),
		PhaseCh:           make(chan int),
		CommunityCards:    [5]poker.Card{},
		numOfActivePlayer: new(int),
		SB:                50,
		BB:                100,
		Pot:               new(int),
		Bet:               new(int),
		OriginalRaiser:    &originalraiser,
	}
}

func (r *PokerRepository) repoTemplateData() map[string]interface{} {
	data := map[string]interface{}{}
	data["playersList"] = r.PlayersList                // [models.MaxPlayer]*models.Player{},
	data["winners"] = r.Winners                        // []*models.Player
	data["buttonPlayer"] = r.ButtonPlayer.ToString()   // string
	data["decisionMaker"] = r.DecisionMaker.ToString() // string
	data["phase"] = PhaseString[*r.Phase]              // string
	data["communityCards"] = r.CommunityCards          // [5]poker.Card
	data["NumOfActivePlayer"] = r.numOfActivePlayer    // *int
	data["sb"] = r.SB                                  // int
	data["bb"] = r.BB                                  // int
	data["pot"] = r.Pot                                // *int
	data["bet"] = r.Bet                                // *int
	if *r.OriginalRaiser == models.PresetPlayer {
		bbplayer := r.nextPlayer(r.nextPlayer(*r.ButtonPlayer))
		data["originalRaiser"] = "(Big Blind) " + bbplayer.ToString() // string
	} else {
		data["originalRaiser"] = r.OriginalRaiser.ToString() // string
	}
	return data
}

// NewPokerRepo make a PokerRepository
func (r *PokerRepository) reset() {
	r.Winners = []*models.Player{}
	*r.numOfActivePlayer = 0
	*r.ButtonPlayer = models.PresetPlayer
	for _, p := range r.PlayersList {
		if p == nil {
			continue
		}
		if *r.ButtonPlayer == models.PresetPlayer {
			*r.ButtonPlayer = p.PlayerSeat()
		}
		*r.numOfActivePlayer++
		p.Reset()
	}
	*r.Phase = 0
	r.CommunityCards = [5]poker.Card{}
	*r.Pot = 0
	*r.Bet = 0
	*r.OriginalRaiser = models.PresetPlayer
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
	if *m.PokerRepo.Phase == 4 {
		num := len(m.PokerRepo.Winners)
		pot := *m.PokerRepo.Pot / num
		remainder := *m.PokerRepo.Pot % num
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
	if *m.PokerRepo.OriginalRaiser == nextplayer {
		log.Println("isDeal is True")
		return true
	} else {
		log.Println("isDeal is False")
		return false
	}
}

// nextPhase sends the number of phase to PhaseCh.
func (m *Repository) nextPhase(present models.PlayerSeat) {
	*m.PokerRepo.Phase = (*m.PokerRepo.Phase + 1) % 5
	if *m.PokerRepo.numOfActivePlayer == 1 {
		// Result Phase(4)
		*m.PokerRepo.Phase = 4
	}
	phase := m.PokerRepo.Phase
	switch *phase {
	case 1:
		m.frop()
	case 2:
		m.turn()
	case 3:
		m.river()
	case 4:
		m.setWinners()
	}
	log.Println("send data-----------------------------------------------------------")
	m.PokerRepo.PhaseCh <- *phase
	log.Println("Phase :", m.PokerRepo.Phase)
	// Set Original Raiser
	first := m.PokerRepo.nextPlayer(*m.PokerRepo.ButtonPlayer)
	*m.PokerRepo.OriginalRaiser = first
	// send phase data to each cliant
	m.playerPhaseChange(present, *phase)
	*m.PokerRepo.Bet = 0
}

// Frop is the handler for Frop.
func (m *Repository) frop() {
	for i := 0; i < 3; i++ {
		m.PokerRepo.CommunityCards[i] = poker.Deck.DrawACard()
		log.Println(m.PokerRepo.CommunityCards[i])
	}
}

// Turn is the handler for Turn.
func (m *Repository) turn() {
	m.PokerRepo.CommunityCards[3] = poker.Deck.DrawACard()
	log.Println(m.PokerRepo.CommunityCards[3])
}

// River is the handler for River.
func (m *Repository) river() {
	m.PokerRepo.CommunityCards[4] = poker.Deck.DrawACard()
	log.Println(m.PokerRepo.CommunityCards[4])
}

// nextPlayer returns a next player.
func (r *PokerRepository) nextPlayer(s models.PlayerSeat) models.PlayerSeat {
	// If the number of active players is less than 2, return Preset Player.
	if *r.numOfActivePlayer <= 1 {
		return models.PresetPlayer
	}
	next := s.NextSeat()
	if r.PlayersList[next] == nil || !r.PlayersList[next].IsPlaying() || r.PlayersList[next].Stack() == 0 {
		return r.nextPlayer(next)
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

// isBettable handles bet and returns msg and bool.
func (m *Repository) isBettable(player *models.Player) (msg string, ok bool) {
	ok = true
	if *m.PokerRepo.Bet > player.Bet() {
		// Bet less than the originalraiser's bet
		msg += fmt.Sprintf("You have to bet at least %d dollars.\n", *m.PokerRepo.Bet)
		ok = false
	} else if player.Bet() == player.Stack() {
		// All in
		msg += fmt.Sprintf("You all in %d dollars.\n", player.Bet())
		*m.PokerRepo.Pot += player.Bet()
		if *m.PokerRepo.Bet < player.Bet() || *m.PokerRepo.OriginalRaiser == models.PresetPlayer {
			*m.PokerRepo.Bet = player.Bet()
			*m.PokerRepo.OriginalRaiser = player.PlayerSeat()
		}

	} else {
		// Bet
		msg = fmt.Sprintf("You bet %d dollars\n", player.Bet())
		*m.PokerRepo.Pot += player.Bet()
		if *m.PokerRepo.Bet < player.Bet() || *m.PokerRepo.OriginalRaiser == models.PresetPlayer {
			*m.PokerRepo.Bet = player.Bet()
			*m.PokerRepo.OriginalRaiser = player.PlayerSeat()
		}
	}
	return
}

// blindBet bets small blind and big blind.
func (m *Repository) blindBet() {
	btn := *m.PokerRepo.ButtonPlayer
	sbp := m.PokerRepo.nextPlayer(btn)
	bbp := m.PokerRepo.nextPlayer(sbp)
	m.PokerRepo.PlayersList[sbp].SetBet(m.PokerRepo.SB)
	m.PokerRepo.PlayersList[bbp].SetBet(m.PokerRepo.BB)
	*m.PokerRepo.OriginalRaiser = models.PresetPlayer
	*m.PokerRepo.Bet = m.PokerRepo.BB
}

// playerChange sends each player the next player data with goroutine.
// This function sends the next player's data to DecisionMakerCh.
// This function should be used to send data in ajax connection.
func (m *Repository) playerChange(present, next models.PlayerSeat) {
	*m.PokerRepo.DecisionMaker = next
	for _, p := range m.PokerRepo.PlayersList {
		if p == nil || p.PlayerSeat() == present {
			continue
		}
		go func(player *models.Player) {
			m.PokerRepo.DecisionMakerCh[player.PlayerSeat()] <- next
		}(p)
	}
}

// playerPhaseChange sends each player the number of phase.
func (m *Repository) playerPhaseChange(present models.PlayerSeat, phase int) {
	for _, p := range m.PokerRepo.PlayersList {
		if p == nil || p.PlayerSeat() == present {
			continue
		}
		go func(player *models.Player) {
			m.PokerRepo.PlayersCh[player.PlayerSeat()] <- phase
		}(p)
	}
}

// init
func (m *Repository) initfunc(r *http.Request) *models.Player {
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
	*m.PokerRepo.numOfActivePlayer++
	m.setPlayerInSession(r, player)
	return player
}

// result
func (m *Repository) resultfunc(r *http.Request) map[string]interface{} {
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
	return data
}

// check
func (m *Repository) checkfunc(r *http.Request) *models.TemplateData {
	player := m.getPlayerFromSession(r)
	player.Check()
	td := &models.TemplateData{}
	td.Success = "You checked.\n"
	next := m.PokerRepo.nextPlayer(player.PlayerSeat())
	if m.isDeal(next) {
		// Next Phase.
		m.nextPhase(player.PlayerSeat())
		td.NotieInfo = PhaseString[*m.PokerRepo.Phase]
		if *m.PokerRepo.Phase <= 3 {
			// send poker data to each cliant
			// Next player of the button player is going to play.
			m.playerChange(player.PlayerSeat(), m.PokerRepo.nextPlayer(*m.PokerRepo.ButtonPlayer))
		}
		m.deal()
	} else {
		m.playerChange(player.PlayerSeat(), next)
	}
	m.setPlayerInSession(r, player)
	return td
}

// call
func (m *Repository) callfunc(r *http.Request) *models.TemplateData {
	player := m.getPlayerFromSession(r)
	err := player.SetBet(*m.PokerRepo.Bet)
	td := &models.TemplateData{}
	if err != nil {
		// Can't bet correctly.
		td.Error = "You should All in.\n"
		return td
	}
	if msg, ok := m.isBettable(player); ok {
		// sucess message
		td.Success = msg
		next := m.PokerRepo.nextPlayer(player.PlayerSeat())
		if m.isDeal(next) {
			// Next Phase
			m.nextPhase(player.PlayerSeat())
			td.NotieInfo = PhaseString[*m.PokerRepo.Phase]
			if *m.PokerRepo.Phase <= 3 {
				// send poker data to each cliant
				// Next player of the button player is going to play.
				m.playerChange(player.PlayerSeat(), m.PokerRepo.nextPlayer(*m.PokerRepo.ButtonPlayer))
			}
			m.deal()
		} else {
			m.playerChange(player.PlayerSeat(), next)
		}
	} else {
		// error message
		td.Error = msg
	}
	m.setPlayerInSession(r, player)
	return td
}

// bet
func (m *Repository) betfunc(r *http.Request) *models.TemplateData {
	player := m.getPlayerFromSession(r)
	bet, _ := strconv.Atoi(r.FormValue("Bet"))
	err := player.SetBet(bet)
	td := &models.TemplateData{}
	if err != nil {
		m.playerChange(player.PlayerSeat(), player.PlayerSeat())
		td.Error = "You can't bet more than your stack.\n"
		return td
	}
	if msg, ok := m.isBettable(player); ok {
		// sucess message
		td.Success = msg
		next := m.PokerRepo.nextPlayer(player.PlayerSeat())
		if m.isDeal(next) {
			// Next Phase.
			m.nextPhase(player.PlayerSeat())
			td.NotieInfo = PhaseString[*m.PokerRepo.Phase]
			if *m.PokerRepo.Phase <= 3 {
				// send poker data to each cliant
				// Next player of the button player is going to play.
				m.playerChange(player.PlayerSeat(), m.PokerRepo.nextPlayer(*m.PokerRepo.ButtonPlayer))
			}
			m.deal()
		} else {
			m.playerChange(player.PlayerSeat(), next)
		}
	} else {
		// error message
		td.Error = msg
	}
	m.setPlayerInSession(r, player)
	return td
}

// allin
func (m *Repository) allinfunc(r *http.Request) *models.TemplateData {
	player := m.getPlayerFromSession(r)
	player.AllIn()
	td := &models.TemplateData{}
	if msg, ok := m.isBettable(player); ok {
		// sucess message
		td.Success = msg
		next := m.PokerRepo.nextPlayer(player.PlayerSeat())
		if m.isDeal(next) {
			// Next Phase.
			m.nextPhase(player.PlayerSeat())
			td.NotieInfo = PhaseString[*m.PokerRepo.Phase]
			if *m.PokerRepo.Phase <= 3 {
				// send poker data to each cliant
				// Next player of the button player is going to play.
				m.playerChange(player.PlayerSeat(), m.PokerRepo.nextPlayer(*m.PokerRepo.ButtonPlayer))
			}
			m.deal()
		} else {
			m.playerChange(player.PlayerSeat(), next)
		}
	} else {
		// error message
		td.Error = msg
	}
	m.setPlayerInSession(r, player)
	return td
}

// fold
func (m *Repository) foldfunc(r *http.Request) *models.TemplateData {
	player := m.getPlayerFromSession(r)
	player.Fold()
	*m.PokerRepo.numOfActivePlayer--
	td := &models.TemplateData{}
	td.Success = "You folded.\n"
	next := m.PokerRepo.nextPlayer(player.PlayerSeat())
	if m.isDeal(next) {
		// Next Phase.
		m.nextPhase(player.PlayerSeat())
		td.NotieInfo = PhaseString[*m.PokerRepo.Phase]
		if *m.PokerRepo.Phase <= 3 {
			// send poker data to each cliant
			// Next player of the button player is going to play.
			m.playerChange(player.PlayerSeat(), m.PokerRepo.nextPlayer(*m.PokerRepo.ButtonPlayer))
		}
		m.deal()
	} else {
		m.playerChange(player.PlayerSeat(), next)
	}
	m.setPlayerInSession(r, player)
	return td
}

// start
func (m *Repository) startfunc(r *http.Request) {
	poker.Deck.Reset()
	m.PokerRepo.reset()
	// Blind Bet
	m.blindBet()
	for _, player := range m.PokerRepo.PlayersList {
		// skip nil player
		if player == nil {
			continue
		}
		// redirect each page
		log.Println("Before-------------------------------------")
		m.PokerRepo.PlayersCh[player.PlayerSeat()] <- -1
		log.Println("After-------------------------------------")

	}
	utg := *m.PokerRepo.ButtonPlayer
	for i := 0; i < 3; i++ {
		utg = m.PokerRepo.nextPlayer(utg)
	}
	*m.PokerRepo.DecisionMaker = utg
}

// reset
func (m *Repository) resetfunc(r *http.Request) {
	poker.Deck.Reset()
	*m.PokerRepo.Phase = 0
}

// next
func (m *Repository) nextfunc(r *http.Request) {
	poker.Deck.Reset()
	*m.PokerRepo.Phase = 0
	*m.PokerRepo.ButtonPlayer = m.PokerRepo.nextPlayer(*m.PokerRepo.ButtonPlayer)
	for _, ch := range m.PokerRepo.PlayersCh {
		if ch == nil {
			continue
		} else {
			ch <- -1
		}
	}
	m.blindBet()
	utg := *m.PokerRepo.ButtonPlayer
	for i := 0; i < 3; i++ {
		utg = m.PokerRepo.nextPlayer(utg)
	}
	*m.PokerRepo.DecisionMaker = utg
}

//----------------------------------------- Get Handlers-----------------------------------------//
// MobilePoker is the handler for the mobile poker page.
func (m *Repository) MobilePoker(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	if player == nil {
		player = models.NewPlayer(models.PresetPlayer, 0, 0, false, &[2]poker.Card{})
	}
	data := map[string]interface{}{}
	data["player"] = player.PlayerTemplateData()
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", &models.TemplateData{Data: data})
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
	data := m.resultfunc(r)
	render.RenderTemplate(w, r, "mobile_poker_result.page.tmpl",
		&models.TemplateData{Data: data})
}

// Porker is the handler for the porker page.
func (m *Repository) Poker(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "poker.page.tmpl", &models.TemplateData{Data: data})
}

// PokerStartGame is the handler to start a new game.
func (m *Repository) PokerStartGame(w http.ResponseWriter, r *http.Request) {
	m.startfunc(r)
	data := map[string]interface{}{}
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "poker.page.tmpl", &models.TemplateData{Info: "Prefrop Phase", Data: data})
}

// PokerResetGame is the handler to reset the existing game.
func (m *Repository) PokerResetGame(w http.ResponseWriter, r *http.Request) {
	m.resetfunc(r)
	data := map[string]interface{}{}
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "poker.page.tmpl", &models.TemplateData{Data: data})
}

// PokerNextGame is the handler to proceed to next game.
func (m *Repository) PokerNextGame(w http.ResponseWriter, r *http.Request) {
	m.nextfunc(r)
	data := map[string]interface{}{}
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "poker.page.tmpl", &models.TemplateData{Data: data})
}

// PokerResult is the handler for the result page.
func (m *Repository) PokerResult(w http.ResponseWriter, r *http.Request) {
	data := m.resultfunc(r)
	render.RenderTemplate(w, r, "poker_result.page.tmpl",
		&models.TemplateData{Data: data})
}

// RemotePoker is the handler for the remote poker page.
func (m *Repository) RemotePoker(w http.ResponseWriter, r *http.Request) {
	if m.PokerRepo.PlayersList[0] != nil {
		log.Println(m.PokerRepo.PlayersList[0].Bet())
	}
	player := m.getPlayerFromSession(r)
	if player == nil {
		player = models.NewPlayer(models.PresetPlayer, 0, 0, false, &[2]poker.Card{})
	}
	log.Println(player.PlayerTemplateData())
	data := map[string]interface{}{}
	data["player"] = player.PlayerTemplateData()
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", &models.TemplateData{Data: data})
}

// RemotePokerStartGame is the handler to start a new game.
func (m *Repository) RemotePokerStartGame(w http.ResponseWriter, r *http.Request) {
	m.startfunc(r)
	data := map[string]interface{}{}
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", &models.TemplateData{Data: data})
}

// RemotePokerResetGame is the handler to reset the existing game.
func (m *Repository) RemotePokerResetGame(w http.ResponseWriter, r *http.Request) {
	m.resetfunc(r)
	data := map[string]interface{}{}
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", &models.TemplateData{Data: data})
}

// RemotePokerNextGame is the handler to proceed to next game.
func (m *Repository) RemotePokerNextGame(w http.ResponseWriter, r *http.Request) {
	m.nextfunc(r)
	data := map[string]interface{}{}
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", &models.TemplateData{Data: data})
}

//----------------------------------------- Post Handlers-----------------------------------------//
// MobilePokerInitPost is the handler to initialize the mobile porker page
func (m *Repository) MobilePokerInitPost(w http.ResponseWriter, r *http.Request) {
	player := m.initfunc(r)
	data := map[string]interface{}{}
	data["player"] = player.PlayerTemplateData()
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl",
		&models.TemplateData{Data: player.PlayerTemplateData()})
}

// MobilePokerCallPost is the handler for Call.
func (m *Repository) MobilePokerCallPost(w http.ResponseWriter, r *http.Request) {
	td := m.callfunc(r)
	player := m.getPlayerFromSession(r)
	td.Data = map[string]interface{}{}
	td.Data["player"] = player.PlayerTemplateData()
	td.Data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", td)
}

// MobilePokerBetPost is the handler for All in.
func (m *Repository) MobilePokerAllInPost(w http.ResponseWriter, r *http.Request) {
	td := m.allinfunc(r)
	player := m.getPlayerFromSession(r)
	td.Data = map[string]interface{}{}
	td.Data["player"] = player.PlayerTemplateData()
	td.Data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", td)
}

// MobilePokerBetPost is the handler for Bet.
func (m *Repository) MobilePokerBetPost(w http.ResponseWriter, r *http.Request) {
	td := m.betfunc(r)
	player := m.getPlayerFromSession(r)
	td.Data = map[string]interface{}{}
	td.Data["player"] = player.PlayerTemplateData()
	td.Data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", td)
}

// MobilePokerFoldPost is the handler for Fold.
func (m *Repository) MobilePokerFoldPost(w http.ResponseWriter, r *http.Request) {
	td := m.foldfunc(r)
	player := m.getPlayerFromSession(r)
	td.Data = map[string]interface{}{}
	td.Data["player"] = player.PlayerTemplateData()
	td.Data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", td)
}

// MobilePokerCheckPost is the handler for Check.
func (m *Repository) MobilePokerCheckPost(w http.ResponseWriter, r *http.Request) {
	td := m.checkfunc(r)
	player := m.getPlayerFromSession(r)
	td.Data = map[string]interface{}{}
	td.Data["player"] = player.PlayerTemplateData()
	td.Data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "mobile_poker.page.tmpl", td)
}

// RemotePokerInitPost is the handler to initialize the mobile porker page
func (m *Repository) RemotePokerInitPost(w http.ResponseWriter, r *http.Request) {
	player := m.initfunc(r)
	data := map[string]interface{}{}
	data["player"] = player.PlayerTemplateData()
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "remote_poker.page.tmpl",
		&models.TemplateData{Data: data})
}

// RemotePokerCallPost is the handler for Call.
func (m *Repository) RemotePokerCallPost(w http.ResponseWriter, r *http.Request) {
	td := m.callfunc(r)
	player := m.getPlayerFromSession(r)
	td.Data = map[string]interface{}{}
	td.Data["player"] = player.PlayerTemplateData()
	td.Data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", td)
}

// RemotePokerAllInPost is the handler for All in.
func (m *Repository) RemotePokerAllInPost(w http.ResponseWriter, r *http.Request) {
	td := m.allinfunc(r)
	player := m.getPlayerFromSession(r)
	td.Data = map[string]interface{}{}
	td.Data["player"] = player.PlayerTemplateData()
	td.Data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", td)
}

// RemotePokerBetPost is the handler for Bet.
func (m *Repository) RemotePokerBetPost(w http.ResponseWriter, r *http.Request) {
	td := m.betfunc(r)
	player := m.getPlayerFromSession(r)
	td.Data = map[string]interface{}{}
	td.Data["player"] = player.PlayerTemplateData()
	td.Data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", td)
}

// RemotePokerFoldPost is the handler for Fold.
func (m *Repository) RemotePokerFoldPost(w http.ResponseWriter, r *http.Request) {
	td := m.foldfunc(r)
	player := m.getPlayerFromSession(r)
	td.Data = map[string]interface{}{}
	td.Data["player"] = player.PlayerTemplateData()
	td.Data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", td)
}

// RemotePokerCheckPost is the handler for Check.
func (m *Repository) RemotePokerCheckPost(w http.ResponseWriter, r *http.Request) {
	td := m.checkfunc(r)
	player := m.getPlayerFromSession(r)
	td.Data = map[string]interface{}{}
	td.Data["player"] = player.PlayerTemplateData()
	td.Data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", td)
}
