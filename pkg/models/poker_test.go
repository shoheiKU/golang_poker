package models

import (
	"reflect"
	"testing"
)

func TestToHands(t *testing.T) {
	type args struct {
		communityCards *[5]Card
		pocketCards    *[2]Card
	}
	tests := []struct {
		name     string
		args     args
		wantHand *Hand
	}{
		{
			name: "Royal flush",
			args: args{
				communityCards: &[5]Card{
					{10, Suits[0]},
					{11, Suits[0]},
					{12, Suits[0]},
					{4, Suits[1]},
					{5, Suits[2]},
				},
				pocketCards: &[2]Card{
					{1, Suits[0]},
					{13, Suits[0]},
				},
			},
			wantHand: &Hand{
				RoyalFlush,
				Cards{
					{1, Suits[0]},
					{13, Suits[0]},
					{12, Suits[0]},
					{11, Suits[0]},
					{10, Suits[0]},
				},
			},
		},
		{
			name: "Straight flush",
			args: args{
				communityCards: &[5]Card{
					{3, Suits[0]},
					{8, Suits[3]},
					{10, Suits[2]},
					{6, Suits[0]},
					{7, Suits[0]},
				},
				pocketCards: &[2]Card{
					{4, Suits[0]},
					{5, Suits[0]},
				},
			},
			wantHand: &Hand{
				StraightFlush,
				Cards{
					{7, Suits[0]},
					{6, Suits[0]},
					{5, Suits[0]},
					{4, Suits[0]},
					{3, Suits[0]},
				},
			},
		},
		{
			name: "Four of a kind",
			args: args{
				communityCards: &[5]Card{
					{3, Suits[0]},
					{3, Suits[1]},
					{4, Suits[0]},
					{5, Suits[0]},
					{3, Suits[2]},
				},
				pocketCards: &[2]Card{
					{3, Suits[3]},
					{1, Suits[0]},
				},
			},
			wantHand: &Hand{
				FourOfAKind,
				Cards{
					{3, Suits[0]},
					{3, Suits[1]},
					{3, Suits[2]},
					{3, Suits[3]},
					{1, Suits[0]},
				},
			},
		},
		{
			name: "Full house",
			args: args{
				communityCards: &[5]Card{
					{3, Suits[0]},
					{3, Suits[1]},
					{4, Suits[0]},
					{5, Suits[0]},
					{3, Suits[2]},
				},
				pocketCards: &[2]Card{
					{4, Suits[3]},
					{1, Suits[0]},
				},
			},
			wantHand: &Hand{
				FullHouse,
				Cards{
					{3, Suits[0]},
					{3, Suits[1]},
					{3, Suits[2]},
					{4, Suits[0]},
					{4, Suits[3]},
				},
			},
		},
		{
			name: "Flush",
			args: args{
				communityCards: &[5]Card{
					{1, Suits[2]},
					{3, Suits[0]},
					{9, Suits[0]},
					{10, Suits[0]},
					{4, Suits[3]},
				},
				pocketCards: &[2]Card{
					{5, Suits[0]},
					{1, Suits[0]},
				},
			},
			wantHand: &Hand{
				Flush,
				Cards{
					{1, Suits[0]},
					{10, Suits[0]},
					{9, Suits[0]},
					{5, Suits[0]},
					{3, Suits[0]},
				},
			},
		},
		{
			name: "Straight",
			args: args{
				communityCards: &[5]Card{
					{3, Suits[0]},
					{8, Suits[3]},
					{10, Suits[2]},
					{6, Suits[2]},
					{7, Suits[0]},
				},
				pocketCards: &[2]Card{
					{4, Suits[1]},
					{5, Suits[0]},
				},
			},
			wantHand: &Hand{
				Straight,
				Cards{
					{8, Suits[3]},
					{7, Suits[0]},
					{6, Suits[2]},
					{5, Suits[0]},
					{4, Suits[1]},
				},
			},
		},
		{
			name: "Three of a kind",
			args: args{
				communityCards: &[5]Card{
					{2, Suits[3]},
					{1, Suits[0]},
					{4, Suits[0]},
					{5, Suits[0]},
					{3, Suits[2]},
				},
				pocketCards: &[2]Card{
					{3, Suits[0]},
					{3, Suits[1]},
				},
			},
			wantHand: &Hand{
				ThreeOfAKind,
				Cards{
					{3, Suits[0]},
					{3, Suits[1]},
					{3, Suits[2]},
					{1, Suits[0]},
					{5, Suits[0]},
				},
			},
		},
		{
			name: "Two pairs",
			args: args{
				communityCards: &[5]Card{
					{2, Suits[3]},
					{1, Suits[0]},
					{4, Suits[1]},
					{5, Suits[0]},
					{3, Suits[2]},
				},
				pocketCards: &[2]Card{
					{4, Suits[0]},
					{3, Suits[1]},
				},
			},
			wantHand: &Hand{
				TwoPairs,
				Cards{
					{4, Suits[0]},
					{4, Suits[1]},
					{3, Suits[1]},
					{3, Suits[2]},
					{1, Suits[0]},
				},
			},
		},
		{
			name: "One pair",
			args: args{
				communityCards: &[5]Card{
					{2, Suits[3]},
					{1, Suits[0]},
					{4, Suits[1]},
					{5, Suits[0]},
					{13, Suits[2]},
				},
				pocketCards: &[2]Card{
					{4, Suits[0]},
					{3, Suits[1]},
				},
			},
			wantHand: &Hand{
				OnePair,
				Cards{
					{4, Suits[0]},
					{4, Suits[1]},
					{1, Suits[0]},
					{13, Suits[2]},
					{5, Suits[0]},
				},
			},
		},
		{
			name: "Highcard",
			args: args{
				communityCards: &[5]Card{
					{2, Suits[3]},
					{4, Suits[1]},
					{8, Suits[0]},
					{13, Suits[2]},
					{3, Suits[1]},
				},
				pocketCards: &[2]Card{
					{9, Suits[0]},
					{1, Suits[0]},
				},
			},
			wantHand: &Hand{
				Highcard,
				Cards{
					{1, Suits[0]},
					{13, Suits[2]},
					{9, Suits[0]},
					{8, Suits[0]},
					{4, Suits[1]},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotHand := ToHands(tt.args.communityCards, tt.args.pocketCards); !reflect.DeepEqual(gotHand, tt.wantHand) {
				t.Errorf("ToHands() = %v, want %v", gotHand, tt.wantHand)
			}
		})
	}
}
