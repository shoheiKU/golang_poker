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
	betdata := map[string]interface{}{}
	player := m.getPlayerFromSession(r)
	log.Println(player.PlayerSeat().ToString(), "is waiting")
	select {
	// request.Context is cancelled.
	case <-r.Context().Done():
		return
	// Get data from the former player.
	case betsize := <-m.PokerRepo.PlayersCh[player.PlayerSeat()]:
		log.Println("Get data")
		betdata["BetSize"] = m.PokerRepo.Bet
		if m.PokerRepo.OriginalRaiser != models.PresetPlayer {
			originalraiser := m.PokerRepo.OriginalRaiser
			log.Println(betsize)
			betdata["OriginalRaiser"] = originalraiser.ToString()
		}
		// Write a json as a return to ajax.
		betdataJson, err := json.Marshal(betdata)
		if err != nil {
			log.Println(err)
		}
		w.Write(betdataJson)
	}
}

// WaitingPhaseAjax is the function for waiting next Phase.
func (m *Repository) WaitingPhaseAjax(w http.ResponseWriter, r *http.Request) {
	select {
	// request.Context is cancelled.
	case <-r.Context().Done():
		return
	// Get a data from the former player.
	case phase := <-m.PokerRepo.PhaseCh:
		log.Println("Get a data in WaitingPhaseAjax")
		switch phase {
		case 0:
			m.Frop(w, r)
		case 1:
			m.Turn(w, r)
		case 2:
			m.River(w, r)
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

// Result is the handler for Result.
func (m *Repository) Result(w http.ResponseWriter, r *http.Request) {
	s := struct {
		Function string
	}{Function: "result"}
	sJson, err := json.Marshal(s)
	if err != nil {
		log.Println(err)
	}
	w.Write(sJson)
}

// ToResultPage shows a pop-up that navigates to ResultPage.
func (m *Repository) ToResultPage(w http.ResponseWriter, r *http.Request) {
	val := map[string]string{"popupstr": "Check Result"}
	valJson, err := json.Marshal(val)
	if err != nil {
		log.Println(err)
	}
	w.Write(valJson)
}
