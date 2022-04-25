package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/shoheiKU/golang_poker/pkg/models"
	"github.com/shoheiKU/golang_poker/pkg/poker"
)

// BetsizeAjax is the handler for the betsize.
func (m *Repository) BetsizeAjax(w http.ResponseWriter, r *http.Request) {
	betdata := map[string]int{}
	a := 0
	player := m.getPlayerFromSession(r)
	go func() {
		fmt.Println("Waiting Scan")
		fmt.Scan(&a)
		m.PokerRepo.PlayersCh[player.PlayerSeat()] <- a
	}()
	betdata["betsize"] = <-m.PokerRepo.PlayersCh[player.PlayerSeat()]
	betdata["potsize"] = 100

	betdataJson, err := json.Marshal(betdata)
	if err != nil {
		log.Println(err)
	}
	w.Write(betdataJson)
}

// WaitingTurnAjax is the function for waiting the player's turn.
func (m *Repository) WaitingTurnAjax(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}
	player := m.getPlayerFromSession(r)
	log.Println(player.PlayerSeat().ToString(), "is waiting")
	select {
	// request.Context is cancelled.
	case <-r.Context().Done():
		log.Println("Context Done")
		return
	// Get data from the former player.
	case signal := <-m.PokerRepo.PlayersCh[player.PlayerSeat()]:
		log.Println("Get data")
		switch signal {
		// signal 0 indicates Prefrop phase
		case 0:
			data["func"] = "prefrop"
		// signal 1 indicates Frop phase
		case 1:
			data["func"] = "frop"
		// signal 2 indicates Turn phase
		case 2:
			data["func"] = "turn"
		// signal 3 indicates River phase
		case 3:
			data["func"] = "river"
		// signal 4 indicates ShowDown phase
		case 4:
			data["func"] = "Result"
			data["url"] = "/mobilepoker/result"
		// signal -1 is used to reset and redirect
		case -1:
			log.Println("-1")
			data["func"] = "reset"
			data["redirect"] = "/mobilepoker"
		// signal -2 is used to popup
		case -2:
			data["func"] = "popup"
		}
		// Write a json as a return to ajax.
		dataJson, err := json.Marshal(data)
		if err != nil {
			log.Println(err)
		}
		w.Write(dataJson)
	}
}

// WhoPlayAjax is the function for getting data of the player who is making decisions.
func (m *Repository) WaitingDataAjax(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}
	log.Println("Who Play")
	player := m.getPlayerFromSession(r)
	select {
	// request.Context is cancelled.
	case <-r.Context().Done():
		log.Println("WaitingDataAjax is closed")
		return
	// Get data from the former player.
	case p := <-m.PokerRepo.DecisionMakerCh[player.PlayerSeat()]:
		log.Println("Get data in WaitingDataAjax")
		data["decisionMaker"] = p.ToString()
		data["betSize"] = m.PokerRepo.Bet
		if m.PokerRepo.OriginalRaiser != models.PresetPlayer {
			originalraiser := m.PokerRepo.OriginalRaiser
			data["originalRaiser"] = originalraiser.ToString()
		}
		// Write a json as a return to ajax.
		dataJson, err := json.Marshal(data)
		if err != nil {
			log.Println(err)
		}
		w.Write(dataJson)
	}
}

// WaitingPhaseAjax is the function for waiting next Phase.
func (m *Repository) WaitingPhaseAjax(w http.ResponseWriter, r *http.Request) {
	log.Println("WaitingPhaseAjax is called")
	select {
	// request.Context is cancelled.
	case <-r.Context().Done():
		log.Println("WaitingPhaseAjax context done")
		return
	// Get a data from the former player.
	case phase := <-m.PokerRepo.PhaseCh:
		log.Println("Get a data in WaitingPhaseAjax phase: ", phase)
		switch phase {
		case 1:
			m.Frop(w, r)
		case 2:
			m.Turn(w, r)
		case 3:
			m.River(w, r)
		case 4:
			m.ToResultPage(w, r)
		}
	}
}

// Frop is the handler for Frop.
func (m *Repository) Frop(w http.ResponseWriter, r *http.Request) {
	s := struct {
		Function string
		Cards    [3]poker.Card
	}{Function: "frop", Cards: [3]poker.Card{}}
	for i := 0; i < 3; i++ {
		s.Cards[i] = poker.Deck.DrawACard()
		m.PokerRepo.CommunityCards[i] = s.Cards[i]
	}

	sJson, err := json.Marshal(s)
	if err != nil {
		log.Println(err)
	}
	w.Write(sJson)
}

// Turn is the handler for Turn.
func (m *Repository) Turn(w http.ResponseWriter, r *http.Request) {
	s := struct {
		Function string
		Card     poker.Card
	}{Function: "turn", Card: poker.Card{}}
	s.Card = poker.Deck.DrawACard()
	m.PokerRepo.CommunityCards[3] = s.Card
	sJson, err := json.Marshal(s)
	if err != nil {
		log.Println(err)
	}
	w.Write(sJson)
}

// River is the handler for River.
func (m *Repository) River(w http.ResponseWriter, r *http.Request) {
	s := struct {
		Function string
		Card     poker.Card
	}{Function: "river", Card: poker.Card{}}
	s.Card = poker.Deck.DrawACard()
	m.PokerRepo.CommunityCards[4] = s.Card
	sJson, err := json.Marshal(s)
	if err != nil {
		log.Println(err)
	}
	w.Write(sJson)
}

// ToResultPage shows a pop-up that navigates to ResultPage.
func (m *Repository) ToResultPage(w http.ResponseWriter, r *http.Request) {
	s := struct {
		Function string
		URL      string
	}{Function: "result", URL: "/poker/result"}
	sJson, err := json.Marshal(s)
	if err != nil {
		log.Println(err)
	}
	w.Write(sJson)
}
