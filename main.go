package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Deck struct {
	flashcards []Flashcard
	numCards   int
	log        []string
	reader     *bufio.Reader
}

type Flashcard struct {
	Term       string `json:"term"`
	Definition string `json:"definition"`
	Mistakes   int    `json:"mistakes"`
}

func (f *Flashcard) String() string {
	return fmt.Sprintf("Card:\n%s\nDefinition:\n%s", f.Term, f.Definition)
}

func NewDeck() *Deck {
	return &Deck{
		flashcards: make([]Flashcard, 0),
		reader:     bufio.NewReader(os.Stdin),
	}
}

func (d *Deck) printlnAndLog(str string) {
	fmt.Println(str)
	d.log = append(d.log, str)
}

func (d *Deck) hasDuplicatedDefinition(definition string) (bool, int) {
	for idx, f := range d.flashcards {
		if definition == f.Definition {
			return true, idx
		}
	}
	return false, 0
}

func (d *Deck) ReadInputLineAndLog() (str string, err error) {
	str, err = d.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	str = strings.TrimSpace(str)
	d.log = append(d.log, str)
	return
}

func (d *Deck) addFlashcard() (int, error) {
	var term string
	var definition string
	var err error
	var duplicated bool

	d.printlnAndLog("The card:")
	for duplicated = true; duplicated; {
		duplicated = false
		term, err = d.ReadInputLineAndLog()
		if err != nil {
			return 0, err
		}
		for _, f := range d.flashcards {
			if f.Term == term {
				duplicated = true
				d.printlnAndLog(fmt.Sprintf("The card \"%s\" already exists. Try again:", term))
				break
			}
		}
	}

	d.printlnAndLog("The definition of the card:")
	for duplicated = true; duplicated; {
		duplicated = false
		definition, err = d.ReadInputLineAndLog()
		if err != nil {
			return 0, err
		}
		for _, f := range d.flashcards {
			if f.Definition == definition {
				duplicated = true
				d.printlnAndLog(fmt.Sprintf("The definition \"%s\" already exists. Try again:", definition))
				break
			}
		}
	}
	d.flashcards = append(d.flashcards, Flashcard{term, definition, 0})
	d.numCards++
	d.printlnAndLog(fmt.Sprintf("The pair (\"%s\":\"%s\") has been added.", term, definition))
	return 1, nil
}

func (d *Deck) askFlashcardAnswers() (int, error) {
	// Seed the random number generator
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	d.printlnAndLog("How many times to ask?")
	numAskStr, err := d.ReadInputLineAndLog()
	if err != nil {
		return 0, err
	}
	numAsk, err := strconv.Atoi(numAskStr)
	if err != nil {
		return 0, err
	}
	if d.numCards == 0 {
		d.printlnAndLog("There are not currently no flashcards. First add some and try again.")
		return 0, nil
	}
	for i := 0; i < numAsk; i++ {
		_, err := d.checkAnswer(r.Intn(d.numCards))
		if err != nil {
			return 0, err
		}
	}
	return 1, nil
}

func (d *Deck) checkAnswer(idx int) (int, error) {
	d.printlnAndLog(fmt.Sprintf("Print the definition of \"%s\":", d.flashcards[idx].Term))

	answer, err := d.ReadInputLineAndLog()
	if err != nil {
		return 0, err
	}

	fd := d.flashcards[idx].Definition
	if answer == fd {
		d.printlnAndLog("Correct!")
	} else {
		d.flashcards[idx].Mistakes++
		wrongStr := fmt.Sprintf("Wrong. The right answer is \"%s\"", fd)
		if match, dupIdx := d.hasDuplicatedDefinition(answer); match {
			wrongStr += fmt.Sprintf(", but your definition is correct for \"%s\".",
				d.flashcards[dupIdx].Term)
		}
		d.printlnAndLog(wrongStr)
	}

	return 1, nil
}

func (d *Deck) removeFlashcard() (bool, error) {
	d.printlnAndLog("Which card?")
	card, err := d.ReadInputLineAndLog()
	if err != nil {
		return false, err
	}

	for idx, flashcard := range d.flashcards {
		if card == flashcard.Term {
			// Remove the card at the specified index by creating a new slice that
			// excludes the card at that index
			d.flashcards = append(d.flashcards[:idx], d.flashcards[idx+1:]...)
			d.numCards--
			d.printlnAndLog("The card has been removed.")
			return true, nil
		}
	}

	d.printlnAndLog(fmt.Sprintf("Can't remove \"%s\": there is no such card.", card))
	return false, nil
}

func (d *Deck) exportToFile(filename string) error {
	file, err := json.MarshalIndent(d.flashcards, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, file, 0644)
	if err != nil {
		return err
	}

	d.printlnAndLog(fmt.Sprintf("%d cards have been saved.\n", d.numCards))

	return nil
}

func (d *Deck) importFromFile(filename string) error {
	file, err := os.Open(filename)
	if file == nil {
		d.printlnAndLog("File not found.")
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var flashcards []Flashcard
	err = decoder.Decode(&flashcards)
	if err != nil {
		return err
	}

	// Create a map to track existing Flashcards by term
	existingFlashcards := make(map[string]int)
	for i, f := range d.flashcards {
		existingFlashcards[f.Term] = i
	}

	// Update existing Flashcards and add new ones
	for _, f := range flashcards {
		if idx, ok := existingFlashcards[f.Term]; ok {
			// If the Flashcard already exists, update its definition
			d.flashcards[idx].Definition = f.Definition
		} else {
			// If the Flashcard doesn't exist, add it to the deck
			d.flashcards = append(d.flashcards, f)
			d.numCards++
		}
	}

	d.printlnAndLog(fmt.Sprintf("%d cards have been loaded", len(flashcards)))

	return nil
}

func (d *Deck) resetStats() {
	for idx, _ := range d.flashcards {
		d.flashcards[idx].Mistakes = 0
	}
	d.printlnAndLog("Card statistics have been reset.")
}

func (d *Deck) findHardestCards() {
	var hardestCards []string
	maxMistakes := 0

	for _, fc := range d.flashcards {
		if fc.Mistakes > 0 && fc.Mistakes >= maxMistakes {
			if fc.Mistakes > maxMistakes {
				hardestCards = []string{}
				maxMistakes = fc.Mistakes
			}
			hardestCards = append(hardestCards, fc.Term)
		}
	}

	if len(hardestCards) == 0 {
		d.printlnAndLog("There are no cards with errors.")
	} else if len(hardestCards) == 1 {
		d.printlnAndLog(fmt.Sprintf("The hardest card is \"%s\". You have %d errors answering it.",
			hardestCards[0], maxMistakes))
	} else {
		d.printlnAndLog(fmt.Sprintf("The hardest cards are \"%s\". You have %d errors answering it.",
			strings.Join(hardestCards, "\", \""), maxMistakes))
	}
}

func (d *Deck) writeToLogFile() error {
	d.printlnAndLog("File name:")
	filename, err := d.ReadInputLineAndLog()
	if err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, str := range d.log {
		_, err := fmt.Fprintln(file, str)
		if err != nil {
			return err
		}
	}

	d.printlnAndLog("The log has been saved.")

	return nil
}

func main() {
	importFromFilename := flag.String("import_from", "", "Filename to import flashcards from.")
	exportToFilename := flag.String("export_to", "", "Filename to export flashcards to.")
	flag.Parse()

	deck := NewDeck()

	if *importFromFilename != "" {
		if err := deck.importFromFile(*importFromFilename); err != nil {
			deck.printlnAndLog(fmt.Sprintf("Error importing flashcards from file: %s", err))
			return
		}
	}

	for {
		deck.printlnAndLog("Input the action (add, remove, import, export, ask, exit, log, hardest card, reset stats):")
		action, err := deck.ReadInputLineAndLog()
		if err != nil {
			deck.printlnAndLog(fmt.Sprintf("Error reading the action: %s", err))
			return
		}
		switch action {
		case "add":
			_, err = deck.addFlashcard()
			if err != nil {
				deck.printlnAndLog(fmt.Sprintf("Error reading flashcard: %s", err))
				return
			}
		case "remove":
			_, err = deck.removeFlashcard()
			if err != nil {
				deck.printlnAndLog(fmt.Sprintf("Error removing flashcard: %s", err))
				return
			}
		case "import":
			deck.printlnAndLog("File name:")
			filename, err := deck.ReadInputLineAndLog()
			if err != nil {
				deck.printlnAndLog(fmt.Sprintf("Error reading filename: %s", err))
			}
			if err := deck.importFromFile(filename); err != nil {
				deck.printlnAndLog(fmt.Sprintf("Error importing flashcards from file: %s", err))
				return
			}
		case "export":
			deck.printlnAndLog("File name:")
			filename, err := deck.ReadInputLineAndLog()
			if err != nil {
				deck.printlnAndLog(fmt.Sprintf("Error reading filename: %s", err))
			}
			if err := deck.exportToFile(filename); err != nil {
				deck.printlnAndLog(fmt.Sprintf("Error exporting flashcards to file: %s", err))
				return
			}
		case "ask":
			_, err = deck.askFlashcardAnswers()
			if err != nil {
				deck.printlnAndLog(fmt.Sprintf("Error asking for Flashcard answers: %s", err))
				return
			}
		case "log":
			if err := deck.writeToLogFile(); err != nil {
				deck.printlnAndLog(fmt.Sprintf("Error writing log to file: %s", err))
				return
			}
		case "hardest card":
			deck.findHardestCards()
		case "reset stats":
			deck.resetStats()
		case "exit":
			deck.printlnAndLog("Bye bye!")
			if *exportToFilename != "" {
				if err := deck.exportToFile(*exportToFilename); err != nil {
					deck.printlnAndLog(fmt.Sprintf("Error exporting flashcards to file: %s", err))
				}
			}
			return
		default:
			deck.printlnAndLog(fmt.Sprintf("Action %s is not implemented. Try again.", action))
		}
	}
}
