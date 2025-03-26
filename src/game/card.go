package game

import "strings"

type CardType int

type Card struct {
	Name string
	Link string
	Type CardType
}

const (
	NumberCard CardType = iota
	SkipCard
	ReverseCard
	DrawTwoCard
	WildCard
	WildDrawFourCard
	CardBack
)

// Card deck
var Cards = []Card{
	{"card-back", "https://i.ibb.co/gZC3QMvC/deck.png", CardBack},
	{"wild-draw", "https://i.ibb.co/k24CjT9F/wild-draw.png", WildDrawFourCard},
	{"wild", "https://i.ibb.co/gZJTh4qN/wild.png", WildCard},
	{"blue-0", "https://i.ibb.co/zH6jpmL7/blue-0.png", NumberCard},
	{"blue-1", "https://i.ibb.co/RTCrDFjD/blue-1.png", NumberCard},
	{"blue-2", "https://i.ibb.co/JFbg1XW4/blue-2.png", NumberCard},
	{"blue-3", "https://i.ibb.co/zhXX8WCL/blue-3.png", NumberCard},
	{"blue-4", "https://i.ibb.co/Ng3yC2s1/blue-4.png", NumberCard},
	{"blue-5", "https://i.ibb.co/q3cwHJ0Q/blue-5.png", NumberCard},
	{"blue-6", "https://i.ibb.co/ksnSvCkc/blue-6.png", NumberCard},
	{"blue-7", "https://i.ibb.co/bgvPH919/blue-7.png", NumberCard},
	{"blue-8", "https://i.ibb.co/SjRgqq6/blue-8.png", NumberCard},
	{"blue-9", "https://i.ibb.co/PGdCwZQ2/blue-9.png", NumberCard},
	{"blue-draw", "https://i.ibb.co/0p9BSpQ9/blue-draw.png", DrawTwoCard},
	{"blue-reverse", "https://i.ibb.co/99rq6LLK/blue-reverse.png", ReverseCard},
	{"blue-skip", "https://i.ibb.co/d4QFhfqy/blue-skip.png", SkipCard},
	{"green-0", "https://i.ibb.co/DHzXh6mc/green-0.png", NumberCard},
	{"green-1", "https://i.ibb.co/fVj1h1Fp/green-1.png", NumberCard},
	{"green-2", "https://i.ibb.co/d0P8PGnX/green-2.png", NumberCard},
	{"green-3", "https://i.ibb.co/0VZ0FgWW/green-3.png", NumberCard},
	{"green-4", "https://i.ibb.co/PZxVVzP5/green-4.png", NumberCard},
	{"green-5", "https://i.ibb.co/r2gPKCLJ/green-5.png", NumberCard},
	{"green-6", "https://i.ibb.co/4n931Ld1/green-6.png", NumberCard},
	{"green-7", "https://i.ibb.co/BHLH3Zc7/green-7.png", NumberCard},
	{"green-8", "https://i.ibb.co/93s8cJxg/green-8.png", NumberCard},
	{"green-9", "https://i.ibb.co/v6fYmLyZ/green-9.png", NumberCard},
	{"green-draw", "https://i.ibb.co/FkZf5T1f/green-draw.png", DrawTwoCard},
	{"green-reverse", "https://i.ibb.co/Cp98BmTs/green-reverse.png", ReverseCard},
	{"green-skip", "https://i.ibb.co/LXhfQ6Zd/green-skip.png", SkipCard},
	{"red-0", "https://i.ibb.co/35vjvXCB/red-0.png", NumberCard},
	{"red-1", "https://i.ibb.co/bjTHky2W/red-1.png", NumberCard},
	{"red-2", "https://i.ibb.co/gF7qKQdf/red-2.png", NumberCard},
	{"red-3", "https://i.ibb.co/V6wZjd1/red-3.png", NumberCard},
	{"red-4", "https://i.ibb.co/YBMShLkB/red-4.png", NumberCard},
	{"red-5", "https://i.ibb.co/QvLfhd1B/red-5.png", NumberCard},
	{"red-6", "https://i.ibb.co/s9QjsT2V/red-6.png", NumberCard},
	{"red-7", "https://i.ibb.co/39hd44c5/red-7.png", NumberCard},
	{"red-8", "https://i.ibb.co/PZLmj0BS/red-8.png", NumberCard},
	{"red-9", "https://i.ibb.co/WWhKHtdn/red-9.png", NumberCard},
	{"red-draw", "https://i.ibb.co/RTc4qcX0/red-draw.png", DrawTwoCard},
	{"red-reverse", "https://i.ibb.co/hRd3697R/red-reverse.png", ReverseCard},
	{"red-skip", "https://i.ibb.co/dJpxW3Tb/red-skip.png", SkipCard},
	{"yellow-0", "https://i.ibb.co/zh3RjBYh/yellow-0.png", NumberCard},
	{"yellow-1", "https://i.ibb.co/YFJGrzTs/yellow-1.png", NumberCard},
	{"yellow-2", "https://i.ibb.co/k2Zf65NY/yellow-2.png", NumberCard},
	{"yellow-3", "https://i.ibb.co/7tgLrxRY/yellow-3.png", NumberCard},
	{"yellow-4", "https://i.ibb.co/TMXLJ7tP/yellow-4.png", NumberCard},
	{"yellow-5", "https://i.ibb.co/Csb3cH73/yellow-5.png", NumberCard},
	{"yellow-6", "https://i.ibb.co/rGH6Rx5m/yellow-6.png", NumberCard},
	{"yellow-7", "https://i.ibb.co/bkF1j7G/yellow-7.png", NumberCard},
	{"yellow-8", "https://i.ibb.co/SXG5Xn5m/yellow-8.png", NumberCard},
	{"yellow-9", "https://i.ibb.co/KxyyP2Z4/yellow-9.png", NumberCard},
	{"yellow-draw", "https://i.ibb.co/jPFpFcjb/yellow-Draw.png", DrawTwoCard},
	{"yellow-reverse", "https://i.ibb.co/TM34FRC9/yellow-reverse.png", ReverseCard},
	{"yellow-skip", "https://i.ibb.co/r2Bkc4w9/yellow-skip.png", SkipCard},
}

func GenerateDeck() []Card {
	var deck []Card

	for _, card := range Cards {
		switch card.Type {
		case NumberCard:
			// Add "0" card once, others twice
			if !strings.Contains(card.Name, "-0") {
				deck = append(deck, card)
				deck = append(deck, card)
			} else {
				deck = append(deck, card)
			}
		case SkipCard, ReverseCard, DrawTwoCard:
			// Add each action card twice
			deck = append(deck, card)
			deck = append(deck, card)
		case WildCard, WildDrawFourCard:
			// Add wild cards four times
			deck = append(deck, card)
			deck = append(deck, card)
			deck = append(deck, card)
			deck = append(deck, card)
		}
	}

	return deck
}
