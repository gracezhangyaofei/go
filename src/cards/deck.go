package main

import (
	"fmt"
	"strings"
	"io/ioutil"
	"os"
	"math/rand"
	"time"
)

type deck []string

// annotate the function with the type that it returns
// iterate through the list
func newDeck() deck {
	cards := deck{}
	cardSuits := []string{"Spades", "Diamonds", "Hearts", "Clubs"}
	cardValues := []string{"Ace", "Two", "Thres", "Four"}

	for _, suit := range cardSuits {
		for _, value := range cardValues {
			cards = append(cards, value+" of "+suit)
		}
	}

	return cards
}

// Just use the first character of receiver
// Avoid using self or this in go
func (d deck) print() {
	for i, card := range d {
		fmt.Println(i, card)
	}
}

// not put deck as receiver because we are not changing deck
func deal(d deck, handSize int) (deck, deck) {
	return d[:handSize], d[handSize:]
}

func newDeckFromFile(filename string) deck {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	s := strings.Split(string(bs), ",")
	return deck(s)
}

// we are changing deck here
func (d deck) shuffle() {
	// specify a seed for rand, otherwise rand will generate the same list
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	for i := range d {
		newPosition := r.Intn(len(d) - 1)

		d[i], d[newPosition] = d[newPosition], d[i]
	}
}

func (d deck) saveToFile(filename string) {
	ioutil.WriteFile(filename, []byte(d.toString()), os.ModePerm)
}

func (d deck) toString() string {
	// return []string(d)
	// return strings.Join(d, ",")
	return strings.Join([]string(d), ",")
}
