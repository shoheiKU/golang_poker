package models

import (
	"log"
)

var ()

// PlayerData is a data structure for each user
type PlayerData struct {
	PlayerId PlayerId
	Stack    int
	Bet      int
}

// PlayerId is the number of Player
type PlayerId int

const (
	Player1 PlayerId = iota
	Player2
	Player3
	Player4
	Player5
	Player6
	Player7
	Player8
	Player9
	MaxPlayer
)

// ItoPlayerId converts int to type PlayerId.
func ItoPlayerId(i int) PlayerId {
	switch i {
	case 0:
		return Player1
	case 1:
		return Player2
	case 2:
		return Player3
	case 3:
		return Player4
	case 4:
		return Player5
	case 5:
		return Player6
	case 6:
		return Player7
	case 7:
		return Player8
	case 8:
		return Player9
	default:
		log.Println("This number is incorrect.")
		return MaxPlayer
	}
}

// ItoPlayerId converts string to type PlayerId.
func AtoPlayerId(s string) PlayerId {
	switch s {
	case "Player1":
		return Player1
	case "Player2":
		return Player2
	case "Player3":
		return Player3
	case "Player4":
		return Player4
	case "Player5":
		return Player5
	case "Player6":
		return Player6
	case "Player7":
		return Player7
	case "Player8":
		return Player8
	case "Player9":
		return Player9
	default:
		log.Println("This number is incorrect.")
		return MaxPlayer
	}
}

// Tostring returns string according to PlayerId.
func (p PlayerId) ToString() string {
	switch p {
	case Player1:
		return "player1"
	case Player2:
		return "player2"
	case Player3:
		return "player3"
	case Player4:
		return "player4"
	case Player5:
		return "player5"
	case Player6:
		return "player6"
	case Player7:
		return "player7"
	case Player8:
		return "player8"
	case Player9:
		return "player9"
	default:
		log.Println("This number is incorrect.")
		return "This number is incorrect."
	}
}

// NowPlayer returns PlayerId. I will change this later.
func (r PlayerId) NowPlayer() PlayerId {
	return r
}

// MaxPlayer returns MaxPlayerId.
func (r PlayerId) MaxPlayer() PlayerId {
	return MaxPlayer
}

// NextPlayer returns next PlayerId.
func (r PlayerId) NextPlayer() PlayerId {
	return PlayerId(int(r)+1) % r.MaxPlayer()
}
