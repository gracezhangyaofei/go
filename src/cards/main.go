package main

func main() {
	cards := newDeck()

	hand, remainingDeck := deal(cards, 5)
	hand.print()
	remainingDeck.print()

	cards.saveToFile("deck.txt")

	cards1 := newDeckFromFile("deck.txt")
	cards1.print()

	cards.shuffle()
	cards.print()

	// Print err
	newDeckFromFile("deck")
}
