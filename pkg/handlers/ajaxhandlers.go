package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/shoheiKU/golang_poker/pkg/models"
)

// MobileWaitingPhaseAjax is the function for waiting next Phase in a cliant page.
func (m *Repository) MobileWaitingPhaseAjax(w http.ResponseWriter, r *http.Request) {
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
		data["bet"] = player.Bet()
		data["stack"] = player.Stack()
		log.Println("Get data")
		switch signal {
		// signal 0 indicates Prefrop phase
		case 0:
			data["func"] = "prefrop"
			data["text"] = PhaseString[0]
		// signal 1 indicates Frop phase
		case 1:
			data["func"] = "frop"
			data["cards"] = m.PokerRepo.CommunityCards[0:3]
			data["text"] = PhaseString[1]
		// signal 2 indicates Turn phase
		case 2:
			data["func"] = "turn"
			data["card"] = m.PokerRepo.CommunityCards[3]
			data["text"] = PhaseString[2]
		// signal 3 indicates River phase
		case 3:
			data["func"] = "river"
			data["card"] = m.PokerRepo.CommunityCards[4]
			data["text"] = PhaseString[3]
		// signal 4 indicates ShowDown phase
		case 4:
			data["func"] = "result"
			if r.FormValue("from") == "remotepoker" {
				data["URL"] = "/remotepoker/result"
			} else {
				data["URL"] = "/mobilepoker/result"
			}
			data["text"] = PhaseString[4]
		// signal -1 is used to reset and redirect
		case -1:
			log.Println("-1")
			data["func"] = "reset"
			if r.FormValue("from") == "remotepoker" {
				data["redirect"] = "/remotepoker"
			} else {
				data["redirect"] = "/mobilepoker"
			}
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

// WaitingDataAjax is the function for getting other players' betting data.
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
		if m.PokerRepo.OriginalRaiser == models.PresetPlayer {
			// Blind Bet
			bbplayer := m.PokerRepo.nextPlayer(m.PokerRepo.nextPlayer(m.PokerRepo.ButtonPlayer))
			data["originalRaiser"] = "(Big Blind) " + bbplayer.ToString()
		} else {
			originalraiser := m.PokerRepo.OriginalRaiser
			log.Println(m.PokerRepo.OriginalRaiser)
			data["originalRaiser"] = originalraiser.ToString()
		}
		// Write a json as a return to ajax.
		log.Println(data)
		dataJson, err := json.Marshal(data)
		if err != nil {
			log.Println(err)
		}
		w.Write(dataJson)
	}
}

// WaitingPhaseAjax is the function for waiting next Phase in a poker page.
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
		data := map[string]interface{}{}
		switch phase {
		case 1:
			// Frop
			data["func"] = "frop"
			data["cards"] = m.PokerRepo.CommunityCards[0:3]
			data["text"] = PhaseString[1]
		case 2:
			// Turn
			data["func"] = "turn"
			data["card"] = m.PokerRepo.CommunityCards[3]
			data["text"] = PhaseString[2]
		case 3:
			// Frop
			data["func"] = "frop"
			data["card"] = m.PokerRepo.CommunityCards[4]
			data["text"] = PhaseString[3]
		case 4:
			// Result
			data["func"] = "result"
			data["URL"] = "/poker/result"
		}
		returnjson, err := json.Marshal(data)
		if err != nil {
			log.Println(err)
		}
		w.Write(returnjson)
	}
}
