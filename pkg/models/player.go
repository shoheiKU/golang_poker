package models

import (
	"errors"
	"log"
)

// PlayerSeat is the number of Seat
type PlayerSeat int

const InitialStack = 1000

const (
	Player1 PlayerSeat = iota
	Player2
	Player3
	Player4
	Player5
	Player6
	Player7
	Player8
	Player9
	MaxPlayer
	PresetPlayer
)

// Player is a data structure for each player.
type Player struct {
	playerSeat  PlayerSeat
	stack       int
	bet         int
	totalwinpot int
	isPlaying   bool
	isAllIn     bool
	pocketCards [2]Card
	hand        *Hand
}

// NewPlayer is a constractor for Player.
func NewPlayer(
	playerSeat PlayerSeat,
	stack int,
	bet int,
	isPlaying bool,
	pocketCards *[2]Card,
) *Player {
	return &Player{
		playerSeat:  playerSeat,
		stack:       stack,
		bet:         bet,
		totalwinpot: 0,
		isPlaying:   isPlaying,
		isAllIn:     false,
		pocketCards: *pocketCards,
		hand:        nil,
	}
}

// ItoPlayerSeat converts int to type PlayerSeat.
func ItoPlayerSeat(i int) PlayerSeat {
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
	case 9:
		return MaxPlayer
	case 10:
		return PresetPlayer
	default:
		log.Println("This number is incorrect.")
		return MaxPlayer
	}
}

// AtoPlayerSeat converts string to type PlayerSeat.
func AtoPlayerSeat(s string) PlayerSeat {
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

// Tostring returns string according to PlayerSeat.
func (p PlayerSeat) ToString() string {
	switch p {
	case Player1:
		return "Player1"
	case Player2:
		return "Player2"
	case Player3:
		return "Player3"
	case Player4:
		return "Player4"
	case Player5:
		return "Player5"
	case Player6:
		return "Player6"
	case Player7:
		return "Player7"
	case Player8:
		return "Player8"
	case Player9:
		return "Player9"
	case PresetPlayer:
		return "Preset Player"
	default:
		log.Println("This number is incorrect.")
		return ""
	}
}

func (p *Player) PlayerTemplateData() map[string]interface{} {
	m := map[string]interface{}{}
	m["bet"] = p.Bet()
	m["stack"] = p.Stack()
	m["winPot"] = p.WinPot()
	m["playerSeat"] = p.PlayerSeat().ToString()
	m["isPlaying"] = p.isPlaying
	pocketCards := p.PocketCards()
	m["pocketCards"] = map[string]Card{"card1": pocketCards[0], "card2": pocketCards[1]}
	m["hand"] = p.HandTemplateData()
	return m
}

// NowPlayer returns PlayerSeat. I will change this later.
func (r PlayerSeat) NowPlayer() PlayerSeat {
	return r
}

// MaxPlayer returns MaxPlayerSeat.
func (r PlayerSeat) MaxPlayer() PlayerSeat {
	return MaxPlayer
}

// NextSeat returns next PlayerSeat.
func (r PlayerSeat) NextSeat() PlayerSeat {
	return PlayerSeat(int(r)+1) % r.MaxPlayer()
}

func (p *Player) SetPocketCards() {
	p.pocketCards[0] = Deck.DrawACard()
	p.pocketCards[1] = Deck.DrawACard()
}

func (p *Player) PocketCards() *[2]Card {
	return &p.pocketCards
}
func (p *Player) SetHand(communityCards *[5]Card) {
	hand := ToHands(communityCards, &p.pocketCards)
	p.hand = hand
}

func (p *Player) Hand() *Hand {
	return p.hand
}

func (p *Player) HandTemplateData() *HandTemplateData {
	if p.hand == nil {
		return nil
	}
	data := &HandTemplateData{}
	data.Val = p.Hand().Val.ToString()
	data.Cards = p.Hand().Cards
	return data
}

func (p *Player) SetBet(bet int) error {
	if p.stack < bet {
		return errors.New("bet over stack")
	} else {
		log.Println("SetBet ", p.PlayerSeat(), bet)
		p.bet = bet
		return nil
	}
}

func (p *Player) AllIn() {
	p.bet = p.stack
	p.isAllIn = true
}

func (p *Player) Bet() int {
	return p.bet
}

func (p *Player) Check() {
	p.bet = 0
}

func (p *Player) Stack() int {
	return p.stack
}

func (p *Player) Deal() int {
	bet := p.bet
	p.stack -= p.bet
	p.bet = 0
	return bet
}

func (p *Player) AddWinPot(i int) {
	p.stack += i
	p.totalwinpot += i
}

func (p *Player) WinPot() int {
	return p.totalwinpot
}
func (p *Player) Fold() {
	p.isPlaying = false
}

func (p *Player) PlayerSeat() PlayerSeat {
	return p.playerSeat
}

func (p *Player) IsPlaying() bool {
	return p.isPlaying
}

func (p *Player) IsAllIn() bool {
	return p.isAllIn
}

func (p *Player) Reset() {
	p.bet = 0
	p.totalwinpot = 0
	p.isPlaying = true
	p.isAllIn = false
	p.SetPocketCards()
	p.hand = nil
}

func (p *Player) Init() {
	p.stack = InitialStack
	p.bet = 0
	p.totalwinpot = 0
	p.isPlaying = true
	p.isAllIn = false
	p.SetPocketCards()
	p.hand = nil
}
