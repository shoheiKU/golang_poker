package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/shoheiKU/golang_poker/pkg/models"
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
	ActivePlayersList []*models.Player
	AllInPlayersList  []*models.Player
	Winners           []*models.Player
	ButtonPlayer      models.PlayerSeat
	DecisionMaker     models.PlayerSeat
	DecisionMakerCh   [models.MaxPlayer]chan models.PlayerSeat
	Phase             int
	PhaseCh           chan int
	CommunityCards    [5]models.Card
	SB                int
	BB                int
	SidePots          []models.SidePot
	Pot               int
	Bet               int
	OriginalRaiser    models.PlayerSeat
	IsPlaying         bool
}

// NewPokerRepo make a PokerRepository
func NewPokerRepo() *PokerRepository {
	return &PokerRepository{
		PlayersCh:       [models.MaxPlayer]chan int{},
		PlayersList:     [models.MaxPlayer]*models.Player{},
		Winners:         []*models.Player{},
		ButtonPlayer:    models.PresetPlayer,
		DecisionMaker:   models.PresetPlayer,
		DecisionMakerCh: [models.MaxPlayer]chan models.PlayerSeat{},
		Phase:           0,
		PhaseCh:         make(chan int),
		CommunityCards:  [5]models.Card{},
		SB:              50,
		BB:              100,
		SidePots:        []models.SidePot{},
		Pot:             0,
		Bet:             0,
		OriginalRaiser:  models.PresetPlayer,
		IsPlaying:       false,
	}
}

func (r *PokerRepository) repoTemplateData() map[string]interface{} {
	data := make(map[string]interface{})
	data["playersList"] = r.PlayersList                // [models.MaxPlayer]*models.Player{},
	data["winners"] = r.Winners                        // []*models.Player
	data["buttonPlayer"] = r.ButtonPlayer.ToString()   // string
	data["decisionMaker"] = r.DecisionMaker.ToString() // string
	data["phase"] = PhaseString[r.Phase]               // string
	data["communityCards"] = r.CommunityCards          // [5]models.Card
	data["sb"] = r.SB                                  // int
	data["bb"] = r.BB                                  // int
	data["pot"] = r.Pot                                // int
	data["bet"] = r.Bet                                // int
	data["sidepots"] = r.SidePots                      // []SidePot
	if r.OriginalRaiser == models.PresetPlayer {
		bbplayer := r.nextPlayer(r.nextPlayer(r.ButtonPlayer))
		data["originalRaiser"] = "(Big Blind) " + bbplayer.ToString() // string
	} else {
		data["originalRaiser"] = r.OriginalRaiser.ToString() // string
	}
	return data
}

// reset resets a PokerRepository
func (r *PokerRepository) reset(nextButtonPlayer models.PlayerSeat) {
	r.Winners = []*models.Player{}
	r.ButtonPlayer = nextButtonPlayer
	for _, p := range r.PlayersList {
		if p == nil {
			continue
		}
		p.Reset()
	}
	r.Phase = 0
	r.CommunityCards = [5]models.Card{}
	r.Pot = 0
	r.Bet = 0
	r.OriginalRaiser = models.PresetPlayer
	utg := r.ButtonPlayer
	// set under the gun player
	for i := 0; i < 3; i++ {
		utg = r.nextPlayer(utg)
	}
	r.DecisionMaker = utg
	r.SidePots = []models.SidePot{}
}

// init initializes a PokerRepository
func (r *PokerRepository) init() {
	r.Winners = []*models.Player{}
	r.ButtonPlayer = models.PresetPlayer
	// set button player and initialize all players
	for _, p := range r.PlayersList {
		if p == nil {
			continue
		}
		if r.ButtonPlayer == models.PresetPlayer {
			r.ButtonPlayer = p.PlayerSeat()
		}
		p.Init()
	}
	r.Phase = 0
	r.CommunityCards = [5]models.Card{}
	r.Pot = 0
	r.Bet = 0
	r.OriginalRaiser = models.PresetPlayer
	utg := r.ButtonPlayer
	// set under the gun player
	for i := 0; i < 3; i++ {
		utg = r.nextPlayer(utg)
	}
	r.DecisionMaker = utg
	r.IsPlaying = false
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
	if len(m.PokerRepo.AllInPlayersList) > 0 {
		sort.Slice(m.PokerRepo.AllInPlayersList, func(i, j int) bool {
			return m.PokerRepo.AllInPlayersList[i].Bet() < m.PokerRepo.AllInPlayersList[j].Bet()
		})
		// make SidePot
		for _, allinp := range m.PokerRepo.AllInPlayersList {
			allInBet := allinp.Bet()
			sidePot := models.SidePot{}
			if allInBet == 0 {
				continue
			}
			for _, p := range m.PokerRepo.PlayersList {
				if p == nil {
					continue
				}
				if p.Bet() < allInBet {
					m.PokerRepo.Pot += p.Deal()
				} else {
					bet := p.Bet()
					p.SetBet(allInBet)
					m.PokerRepo.Pot += p.Deal()
					p.SetBet(bet - allInBet)
					sidePot.Players = append(sidePot.Players, p)
				}
			}
			// set side pot
			sidePot.Pot = m.PokerRepo.Pot
			m.PokerRepo.Pot = 0
			m.PokerRepo.SidePots = append(m.PokerRepo.SidePots, sidePot)
		}
	} else {
		for _, p := range m.PokerRepo.PlayersList {
			if p == nil {
				continue
			} else {
				m.PokerRepo.Pot += p.Deal()
			}
		}
	}
}

// isDeal returns whether or not all plaayers bet same amount of chips as a bool.
func (m *Repository) isDeal() bool {
	nextplayer := m.PokerRepo.nextPlayer(m.PokerRepo.DecisionMaker)
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
func (m *Repository) nextPhase(present models.PlayerSeat) {
	m.PokerRepo.Phase = (m.PokerRepo.Phase + 1) % 5
	phase := m.PokerRepo.Phase
	switch phase {
	case 1:
		m.frop()
	case 2:
		m.turn()
	case 3:
		m.river()
	case 4:
		m.setWinPot()
	}
	log.Println("send data-----------------------------------------------------------")
	go func() {
		m.PokerRepo.PhaseCh <- phase
	}()
	log.Println("Phase :", m.PokerRepo.Phase)
	// Set Original Raiser
	first := m.PokerRepo.nextPlayer(m.PokerRepo.ButtonPlayer)
	m.PokerRepo.OriginalRaiser = first
	// send phase data to each cliant
	m.playerPhaseChange(present, phase)
	m.PokerRepo.Bet = 0
}

// finalizeGame proceeds the phase to result phase.
// This function sends the number of result phase to PhaseCh.
func (m *Repository) finalizeGame(present models.PlayerSeat) {
	log.Println("finalizeGame")
	m.PokerRepo.Phase = (m.PokerRepo.Phase + 1) % 5
	phase := m.PokerRepo.Phase
	m.PokerRepo.Phase = 4
	switch phase {
	case 1:
		m.frop()
		fallthrough
	case 2:
		m.turn()
		fallthrough
	case 3:
		m.river()
		fallthrough
	case 4:
		m.setWinPot()
	}
	log.Println("send data-----------------------------------------------------------")
	go func() {
		m.PokerRepo.PhaseCh <- phase
	}()
	log.Println("Phase :", m.PokerRepo.Phase)
	// send phase data to each cliant
	m.playerPhaseChange(present, phase)
}

// finalizeGame proceeds the phase to result phase.
// This function sends the number of result phase to PhaseCh.
func (m *Repository) endGame(present models.PlayerSeat) {
	log.Println("endGame")
	m.PokerRepo.Phase = 4
	m.PokerRepo.Winners = []*models.Player{}
	for _, p := range m.PokerRepo.PlayersList {
		if p == nil || !p.IsPlaying() {
			continue
		}
		m.PokerRepo.Winners = append(m.PokerRepo.Winners, p)
	}
	if len(m.PokerRepo.Winners) != 1 {
		log.Panicln("func endGame error: There are more than two active players")
	}
	m.PokerRepo.Winners[0].AddWinPot(m.PokerRepo.Pot)
	log.Println("send data-----------------------------------------------------------")
	go func() {
		m.PokerRepo.PhaseCh <- m.PokerRepo.Phase
	}()
	log.Println("Phase :", m.PokerRepo.Phase)
	// send phase data to each cliant
	m.playerPhaseChange(present, m.PokerRepo.Phase)
}

// Frop is the handler for Frop.
func (m *Repository) frop() {
	for i := 0; i < 3; i++ {
		m.PokerRepo.CommunityCards[i] = models.Deck.DrawACard()
	}
}

// Turn is the handler for Turn.
func (m *Repository) turn() {
	m.PokerRepo.CommunityCards[3] = models.Deck.DrawACard()
}

// River is the handler for River.
func (m *Repository) river() {
	m.PokerRepo.CommunityCards[4] = models.Deck.DrawACard()
}

// nextPlayer returns a next player.
func (r *PokerRepository) nextPlayer(s models.PlayerSeat) models.PlayerSeat {
	if !r.IsPlaying {
		return models.PresetPlayer
	}
	next := s.NextSeat()
	for {
		if r.PlayersList[next] != nil && r.PlayersList[next].IsPlaying() {
			break
		}
		next = next.NextSeat()
	}
	return next
}

// setWinPot sets winpot for winners.
func (m *Repository) setWinPot() {
	m.PokerRepo.SidePots = append(m.PokerRepo.SidePots, models.SidePot{Pot: m.PokerRepo.Pot, Players: m.PokerRepo.ActivePlayersList})
	for i := len(m.PokerRepo.SidePots) - 1; i >= 0; i-- {
		winners := m.playHand(m.PokerRepo.SidePots[i].Players)
		pot := m.PokerRepo.SidePots[i].Pot / len(winners)
		remainder := m.PokerRepo.SidePots[i].Pot % len(winners)
		for _, player := range winners {
			if player == winners[0] {
				player.AddWinPot(pot + remainder)
			} else {
				player.AddWinPot(pot)
			}
		}
		m.PokerRepo.Winners = append(m.PokerRepo.Winners, winners...)
	}
}

// playHand returns a slice of pointers to models.Player who won the game.
func (m *Repository) playHand(players []*models.Player) []*models.Player {
	var winners []*models.Player
	var winner *models.Player
	for _, player := range players {
		if player == nil || !player.IsPlaying() {
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
	if m.PokerRepo.Bet > player.Bet() {
		// Bet less than the originalraiser's bet
		msg += fmt.Sprintf("You have to bet at least %d dollars.\n", m.PokerRepo.Bet)
		ok = false
	} else if player.Bet() == player.Stack() {
		// All in
		msg += fmt.Sprintf("You all in %d dollars.\n", player.Bet())
		if m.PokerRepo.Bet < player.Bet() || m.PokerRepo.OriginalRaiser == models.PresetPlayer {
			m.PokerRepo.Bet = player.Bet()
			m.PokerRepo.OriginalRaiser = player.PlayerSeat()
		}

	} else {
		// Bet
		msg = fmt.Sprintf("You bet %d dollars\n", player.Bet())
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
	sbp := m.PokerRepo.nextPlayer(btn)
	bbp := m.PokerRepo.nextPlayer(sbp)
	m.PokerRepo.PlayersList[sbp].SetBet(m.PokerRepo.SB)
	m.PokerRepo.PlayersList[bbp].SetBet(m.PokerRepo.BB)
	m.PokerRepo.OriginalRaiser = models.PresetPlayer
	m.PokerRepo.Bet = m.PokerRepo.BB
}

// playerChange sends each player the next player data with goroutine.
// This function sends the next player's data to DecisionMakerCh.
// This function should be used to send data in ajax connection.
func (m *Repository) playerChange(present, next models.PlayerSeat) {
	m.PokerRepo.DecisionMaker = next
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
func (m *Repository) initFunc(r *http.Request) *models.Player {
	playerSeat := models.AtoPlayerSeat(r.FormValue("PlayerSeat"))
	if m.PokerRepo.PlayersList[playerSeat] != nil {
		return nil
	}

	player := models.NewPlayer(
		playerSeat,
		models.InitialStack,
		0,
		true,
		&[2]models.Card{},
	)
	m.PokerRepo.PlayersCh[player.PlayerSeat()] = make(chan int)
	m.PokerRepo.PlayersList[player.PlayerSeat()] = player
	m.PokerRepo.DecisionMakerCh[player.PlayerSeat()] = make(chan models.PlayerSeat)
	m.setPlayerInSession(r, player)
	return player
}

// result
func (m *Repository) resultFunc(r *http.Request) map[string]interface{} {
	data := map[string]interface{}{}
	winners := []map[string]interface{}{}
	for _, p := range m.PokerRepo.Winners {
		winners = append(winners, p.PlayerTemplateData())
	}
	data["winners"] = winners
	showdown := []map[string]interface{}{}
	for _, p := range m.PokerRepo.PlayersList {
		if p == nil || !p.IsPlaying() {
			continue
		}
		showdown = append(showdown, p.PlayerTemplateData())
	}
	data["showdown"] = showdown
	return data
}

func (m *Repository) postActionFunc(td *models.TemplateData, player *models.Player) {
	if m.isDeal() {
		m.deal()
		if len(m.PokerRepo.ActivePlayersList) >= 2 {
			log.Println("a---------------------")
			// Next Phase
			m.nextPhase(player.PlayerSeat())
		} else {
			if len(m.PokerRepo.AllInPlayersList)+len(m.PokerRepo.ActivePlayersList) >= 2 {
				log.Println("b---------------------")
				// there are one or more all-in players
				m.finalizeGame(player.PlayerSeat())
			} else {
				log.Println("c---------------------")
				// the other players folded
				m.endGame(player.PlayerSeat())
			}
		}
		td.NotieInfo = PhaseString[m.PokerRepo.Phase]
		if m.PokerRepo.Phase <= 3 {
			// send poker data to each cliant
			// Next player of the button player is going to play.
			m.playerChange(player.PlayerSeat(), m.PokerRepo.nextPlayer(m.PokerRepo.ButtonPlayer))
		}
	} else {
		m.playerChange(player.PlayerSeat(), m.PokerRepo.nextPlayer(player.PlayerSeat()))
	}
}

func (m *Repository) generalBetFunc(td *models.TemplateData, player *models.Player) {
	if msg, ok := m.isBettable(player); ok {
		// sucess message
		td.Success = msg
		m.postActionFunc(td, player)
	} else {
		// error message
		td.Error = msg
	}
	td.Data = make(map[string]interface{})
	td.Data["player"] = player.PlayerTemplateData()
	td.Data["repo"] = m.PokerRepo.repoTemplateData()
}

// check
func (m *Repository) checkFunc(r *http.Request, player *models.Player) *models.TemplateData {
	player.Check()
	td := &models.TemplateData{}
	td.Success = "You checked.\n"
	td.Data = make(map[string]interface{})
	m.postActionFunc(td, player)
	td.Data["player"] = player.PlayerTemplateData()
	td.Data["repo"] = m.PokerRepo.repoTemplateData()
	return td
}

// call
func (m *Repository) callFunc(r *http.Request, player *models.Player) *models.TemplateData {
	err := player.SetBet(m.PokerRepo.Bet)
	td := &models.TemplateData{}
	if err != nil {
		// Can't bet correctly.
		td.Error = "You should All in.\n"
		return td
	}
	m.generalBetFunc(td, player)
	return td
}

// bet
func (m *Repository) betFunc(r *http.Request, player *models.Player) *models.TemplateData {
	bet, _ := strconv.Atoi(r.FormValue("Bet"))
	err := player.SetBet(bet)
	td := &models.TemplateData{}
	if err != nil {
		m.playerChange(player.PlayerSeat(), player.PlayerSeat())
		td.Error = "You can't bet more than your stack.\n"
		return td
	}
	m.generalBetFunc(td, player)
	return td
}

// allin
func (m *Repository) allInFunc(r *http.Request, player *models.Player) *models.TemplateData {
	player.AllIn()
	td := &models.TemplateData{}
	m.generalBetFunc(td, player)
	return td
}

// fold
func (m *Repository) foldFunc(r *http.Request, player *models.Player) *models.TemplateData {
	player.Fold()
	td := &models.TemplateData{}
	td.Success = "You folded.\n"
	td.Data = make(map[string]interface{})
	td.Data["player"] = player.PlayerTemplateData()
	td.Data["repo"] = m.PokerRepo.repoTemplateData()
	m.postActionFunc(td, player)
	return td
}

// start
func (m *Repository) startFunc(r *http.Request) {
	models.Deck.Reset()
	m.PokerRepo.IsPlaying = true
	for _, player := range m.PokerRepo.PlayersList {
		if player == nil {
			continue
		}
		m.PokerRepo.reset(player.PlayerSeat())
		break
	}
	// Blind Bet
	m.blindBet()
	for _, ch := range m.PokerRepo.PlayersCh {
		// skip nil player
		if ch == nil {
			continue
		}
		// redirect each page
		ch <- -1
	}
}

// reset
func (m *Repository) resetFunc(r *http.Request) {
	models.Deck.Reset()
	m.PokerRepo.init()
	// Blind Bet
	m.blindBet()
	for _, ch := range m.PokerRepo.PlayersCh {
		// skip nil player
		if ch == nil {
			continue
		}
		// redirect each page
		ch <- -1
	}
}

// next
func (m *Repository) nextFunc(r *http.Request) {
	models.Deck.Reset()
	m.PokerRepo.reset(m.PokerRepo.nextPlayer(m.PokerRepo.ButtonPlayer))
	// Blind Bet
	m.blindBet()
	for _, ch := range m.PokerRepo.PlayersCh {
		// skip nil player
		if ch == nil {
			continue
		}
		// redirect each page
		ch <- -1
	}
}

//----------------------------------------- Get Handlers-----------------------------------------//

//.
//└── data
//    ├── player
//    │   └── player.PlayerTemplateData
//    ├── repo
//    │   └── m.PokerRepo.repoTemplateData
//    ├── winners
//    │   └── []player.PlayerTemplateData
//    └── showdown
//        └── []player.PlayerTemplateData

// RemotePoker is the handler for the remote poker page.
func (m *Repository) RemotePoker(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	if player == nil {
		player = models.NewPlayer(models.PresetPlayer, 0, 0, false, &[2]models.Card{})
	}
	data := map[string]interface{}{}
	data["player"] = player.PlayerTemplateData()
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", &models.TemplateData{Data: data})
}

// RemotePokerStartGame is the handler to start a new game.
func (m *Repository) RemotePokerStartGame(w http.ResponseWriter, r *http.Request) {
	m.startFunc(r)
	data := map[string]interface{}{}
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", &models.TemplateData{Data: data})
}

// RemotePokerResetGame is the handler to reset the existing game.
func (m *Repository) RemotePokerResetGame(w http.ResponseWriter, r *http.Request) {
	m.resetFunc(r)
	data := map[string]interface{}{}
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", &models.TemplateData{Data: data})
}

// RemotePokerNextGame is the handler to proceed to next game.
func (m *Repository) RemotePokerNextGame(w http.ResponseWriter, r *http.Request) {
	m.nextFunc(r)
	data := map[string]interface{}{}
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", &models.TemplateData{Data: data})
}

func (m *Repository) RemotePokerResult(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	data := m.resultFunc(r)
	data["player"] = player.PlayerTemplateData()
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "remote_poker_result.page.tmpl", &models.TemplateData{Data: data})
}

//----------------------------------------- Post Handlers-----------------------------------------//
// RemotePokerInitPost is the handler to initialize the remote porker page
func (m *Repository) RemotePokerInitPost(w http.ResponseWriter, r *http.Request) {
	player := m.initFunc(r)
	if player == nil {
		player = models.NewPlayer(models.PresetPlayer, 0, 0, false, &[2]models.Card{})
		playerSeat := r.FormValue("PlayerSeat")
		td := &models.TemplateData{
			NotieError: playerSeat + " is already used.",
			Data:       map[string]interface{}{"player": player.PlayerTemplateData(), "repo": m.PokerRepo.repoTemplateData()},
		}
		render.RenderTemplate(w, r, "remote_poker.page.tmpl", td)
		return
	}
	m.PokerRepo.ActivePlayersList = append(m.PokerRepo.ActivePlayersList, player)
	data := make(map[string]interface{})
	data["player"] = player.PlayerTemplateData()
	data["repo"] = m.PokerRepo.repoTemplateData()
	render.RenderTemplate(w, r, "remote_poker.page.tmpl",
		&models.TemplateData{Data: data})
}

// RemotePokerCallPost is the handler for Call.
func (m *Repository) RemotePokerCallPost(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	td := m.callFunc(r, player)
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", td)
}

// RemotePokerAllInPost is the handler for All in.
func (m *Repository) RemotePokerAllInPost(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	td := m.allInFunc(r, player)
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", td)
}

// RemotePokerBetPost is the handler for Bet.
func (m *Repository) RemotePokerBetPost(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	td := m.betFunc(r, player)
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", td)
}

// RemotePokerFoldPost is the handler for Fold.
func (m *Repository) RemotePokerFoldPost(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	td := m.foldFunc(r, player)
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", td)
}

// RemotePokerCheckPost is the handler for Check.
func (m *Repository) RemotePokerCheckPost(w http.ResponseWriter, r *http.Request) {
	player := m.getPlayerFromSession(r)
	td := m.checkFunc(r, player)
	m.setPlayerInSession(r, player)
	render.RenderTemplate(w, r, "remote_poker.page.tmpl", td)
}
