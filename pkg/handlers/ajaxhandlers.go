package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// BetsizeAjax is the handler for the betsize.
func (m *Repository) BetsizeAjax(w http.ResponseWriter, r *http.Request) {
	betdata := map[string]int{}
	a := 0
	player := m.getPlayerFromSession(r)
	go func() {
		fmt.Println("Waiting Scan")
		fmt.Scan(&a)
		m.PokerRepo.PlayersBet[player.PlayerSeat()] <- a
	}()
	betdata["betsize"] = <-m.PokerRepo.PlayersBet[player.PlayerSeat()]
	betdata["potsize"] = 100

	betdataJson, err := json.Marshal(betdata)
	if err != nil {
		log.Println(err)
	}
	w.Write(betdataJson)
}

// WaitingTurnAjax is the function for Waiting Turn.
func (m *Repository) WaitingTurnAjax(w http.ResponseWriter, r *http.Request) {
	betdata := map[string]interface{}{}
	player := m.getPlayerFromSession(r)
	log.Println(player.PlayerSeat(), "is waiting")
	select {
	// request.Context is cancelled.
	case <-r.Context().Done():
		log.Println("Context Done")
		return
	// Get a data from the former player.
	case betsize := <-m.PokerRepo.PlayersBet[player.PlayerSeat()]:
		log.Println("Get a data")
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
