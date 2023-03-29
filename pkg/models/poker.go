package models

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"
)

var (
	Deck       DeckArray
	numOfCards = 52
	Suits      = [...]string{"clubs", "diamonds", "hearts", "spades"}
)

const (
	Highcard HandValue = iota
	OnePair
	TwoPairs
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

type DeckArray [52]int
type Cards []Card
type HandValue int

type Card struct {
	Num  int
	Suit string
}

type Hand struct {
	Val   HandValue
	Cards Cards
}

type HandTemplateData struct {
	Val   string
	Cards Cards
}

func (v HandValue) ToString() string {
	switch v {
	case Highcard:
		return "Highcard"
	case OnePair:
		return "One pair"
	case TwoPairs:
		return "Two pairs"
	case ThreeOfAKind:
		return "Three of a kind"
	case Straight:
		return "Straight"
	case Flush:
		return "Flush"
	case FullHouse:
		return "Full house"
	case FourOfAKind:
		return "Four of a kind"
	case StraightFlush:
		return "Straight flush"
	case RoyalFlush:
		return "Royal flush"
	default:
		log.Println("This HandValue is incorrect.")
		return ""
	}
}

func (c Cards) Len() int {
	return len(c)
}
func (c Cards) Less(i, j int) bool {
	ci := c[i].Num
	cj := c[j].Num
	if ci == 1 {
		ci += 13
	}
	if cj == 1 {
		cj += 13
	}

	// return the Suits order if the numbers are the same
	if ci == cj {
		return c[i].Suit < c[j].Suit
	}

	// return the numbers order
	return ci < cj
}
func (c Cards) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// toSortedStruct returns a sorted slice of pointers to structs.
// This slice is sorted by the size first, then by the num in strength order which means 2 is the lowest, 3,4,...,13,1 is the highest.
func toSortedStruct(mapIntCards map[int]Cards) []*struct {
	Num   int
	Size  int
	Cards Cards
} {
	s := make([]*struct {
		Num   int
		Size  int
		Cards Cards
	}, 0, len(mapIntCards))
	for key, val := range mapIntCards {
		s = append(s, &struct {
			Num   int
			Size  int
			Cards Cards
		}{key, val.Len(), val})
	}
	sort.Slice(s, func(i, j int) bool {
		if s[i].Size == s[j].Size {
			si := s[i].Num
			sj := s[j].Num
			if si == 1 {
				si += 13
			}
			if sj == 1 {
				sj += 13
			}
			return si > sj
		} else {
			return s[i].Size > s[j].Size
		}
	})
	return s
}

// checkAllFlushHand returns a pointer to a Hand struct and ok as a bool.
// This function checks all hands associated with Flush, which means Flush, Straight Flush, and Royal Flush
// ok is true when the hand is Flush, otherwise it is false.
func checkAllFlushHand(mapSuitCards map[string]Cards) (hand *Hand, ok bool) {
	for _, sameSuitCards := range mapSuitCards {
		if sameSuitCards.Len() >= 5 {
			hand, ok = checkStraight(sameSuitCards)
			if ok { // Straight
				if hand.Cards[0].Num == 1 {
					// Royal Flush
					hand.Val = RoyalFlush
				} else {
					// Straight Flush
					hand.Val = StraightFlush
				}
			} else {
				// Flush
				hand = &Hand{}
				sort.Sort(sort.Reverse(sameSuitCards))
				hand.Val = Flush
				hand.Cards = sameSuitCards[0:5]
				ok = true
			}
		}
	}
	return
}

// checkStraight returns a pointer to a Hand struct and ok as a bool.
// ok is true when the hand is Straight, otherwise it is false.
func checkStraight(cards Cards) (hand *Hand, ok bool) {
	sort.Sort(sort.Reverse(cards))
	formerCard := cards[0]
	var returnCards Cards
	for _, card := range cards {
		if dif := formerCard.Num - card.Num; dif == 1 || dif == -12 {
			// Consecutive
			returnCards = append(returnCards, card)
		} else {
			// Not consecutive
			returnCards = Cards{card}
		}
		// Straight Check
		if len(returnCards) == 5 {
			// Straight
			ok = true
			hand = &Hand{Val: Straight, Cards: returnCards}
			break
		}
		// 1,2,3,4,5 Straight Check
		if len(returnCards) == 4 && card.Num == 2 {
			// 1,2,3,4,5 Straight
			if cards[0].Num == 1 {
				returnCards = append(returnCards, cards[0])
				fmt.Println("1,2,3,4,,5 Size ", returnCards.Len())
				hand = &Hand{Val: Straight, Cards: returnCards}
			}
			break
		}
		formerCard = card
	}
	return
}

// checkPairs returns a pointer to a Hand struct and ok as a bool.
// ok is true when the hand is something. If ok is false, something wrong happened.
func checkPairs(mapIntCards map[int]Cards) (hand *Hand, ok bool) {
	s := toSortedStruct(mapIntCards)
	for _, ss := range s {
		sort.Sort(ss.Cards)
	}
	switch s[0].Size {
	case 4:
		// FourOfAKind
		hand = &Hand{Val: FourOfAKind, Cards: append(s[0].Cards, s[1].Cards[0])}
		ok = true
	case 3:
		if s[1].Size >= 2 {
			// FullHouse
			hand = &Hand{Val: FullHouse, Cards: append(s[0].Cards, s[1].Cards[0:2]...)}
			ok = true
		} else {
			// ThreeOfAKind
			hand = &Hand{Val: ThreeOfAKind, Cards: append(s[0].Cards, s[1].Cards[0], s[2].Cards[0])}
			ok = true
		}

	case 2:
		if s[1].Size >= 2 {
			// TwoPairs
			hand = &Hand{Val: TwoPairs, Cards: append(s[0].Cards, append(s[1].Cards[0:2], s[2].Cards[0])...)}
			ok = true
		} else {
			// OnePair
			hand = &Hand{Val: OnePair, Cards: append(s[0].Cards, s[1].Cards[0], s[2].Cards[0], s[3].Cards[0])}
			ok = true
		}
	case 1:
		// Highcard
		returnCards := make(Cards, 0, 5)
		for i := 0; i < 5; i++ {
			returnCards = append(returnCards, s[i].Cards[0])
		}
		hand = &Hand{Val: Highcard, Cards: returnCards}
		ok = true
	}
	return
}

func ToHands(communityCards *[5]Card, pocketCards *[2]Card) (hand *Hand) {
	// Valuables
	var cards Cards = append(communityCards[:], pocketCards[:]...)
	mapIntCards := map[int]Cards{}
	mapSuitCards := map[string]Cards{}

	for _, card := range cards {
		mapIntCards[card.Num] = append(mapIntCards[card.Num], card)
		mapSuitCards[card.Suit] = append(mapSuitCards[card.Suit], card)
	}

	// Check Flush, Straight Flush, or Royal Flush
	hand, ok := checkAllFlushHand(mapSuitCards)
	if ok {
		return
	}
	// Straight Check
	hand, ok = checkStraight(cards)
	if ok {
		return
	}
	hand, ok = checkPairs(mapIntCards)
	if ok {
		return
	} else {
		log.Fatalln("There is something wrong.")
		return &Hand{}
	}
}

// init sets data to a Deck and sets a seed with the time.
func init() {
	for i := range Deck {
		Deck[i] = i
	}
	rand.Seed(time.Now().UnixNano())
}

// numSuit converts an int to a Card.
func numSuit(i int) Card {
	num := i%13 + 1
	suit := Suits[i/13]
	return Card{num, suit}
}

// DrawACard returns a Card randomly form the Deck.
func (C *DeckArray) DrawACard() Card {
	randnum := rand.Intn(numOfCards)
	numOfCards--
	num := C[randnum]
	C[randnum] = C[numOfCards]
	C[numOfCards] = num
	return numSuit(num)
}

// Reset resets the Deck.
func (C *DeckArray) Reset() {
	numOfCards = 52
}
