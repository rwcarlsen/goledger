package lex

import (
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const EOF = -1

type Trans struct {
	Date    time.Time
	Status  string
	Descrip string
	Items   []*Item
	Note    string
}

type Item struct {
	Status   string
	Account  string
	Amount   *big.Rat
	Commod   string
	ExAmount *big.Rat
	ExCommod string
	Note     string
}

const (
	space      = " \t"
	lineend    = "\r\n"
	whitespace = indent + lineend
	digit      = "0123456789"
	statuss    = "*!"
)

const (
	meta = ";"
	atat = "@@"
	at   = "@"
)

// Decode
func Decode(r io.Reader) (Journal, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	l := &Lexer{
		Input: string(data),
	}
	decode(l * Lexer)
	return l.Journal, nil
}

func decode(l *Lexer) {
	for {
		switch r := l.Peek(); {
		case unicode.IsDigit(r):
			if abort := lexTrans(l); abort {
				return
			}
		case EOF:
			return
		case unicode.IsSpace(r):
			skipBlankLines(l)
		case string(r) == meta:
			skipLine(l)
		default:
			l.AcceptRunNot(lineend)
			log.Printf("unexpected token on line %v: %v", l.Line(), l.Emit())
			l.AcceptRun(lineend)
			l.Emit()
		}
	}
}

func lexTrans(l *Lexer) (t *Trans, abort bool) {
	// date
	l.AcceptRunNot(whitespace)
	s := l.Emit()
	tm, err := time.Parse("06/1/2", s)
	if err != nil {
		log.Printf("invalid date on line %v: %v", l.Line(), err)
		return nil, true
	}

	skipSpace(l)

	// status
	status := ""
	if l.Accept(statuss) {
		status = l.Emit()
	}

	// description
	l.AcceptRunNot(meta + lineend)
	descrip := l.Emit()

	// note
	note := ""
	if l.Accept(meta) {
		l.Emit()
		l.AcceptRunNot(lineend)
		note = l.Emit()
	}

	skipLine(l)

	items, abort := lexItems()
	if abort {
		return nil, abort
	}

	// build transaction
	return &Trans{
		Date:    tm,
		Status:  status,
		Descrip: descrip,
		Note:    note,
		Items:   items,
	}, false
}

func lexItems(l *Lexer) []Item {
	for {
		if r := l.Peek(); r == EOF || r == '\n' || r == '\r' {
			break
		}

		skipSpace(l)

		status := ""
		if l.Accept(statuss) {
			status = l.Emit()
			skipSpace(l)
		}

		account := lexAccount(l)
		skipSpace(l)

		// unit (prefix comomd)
		amount, commod := lexAmount(l)

		skipSpace(l)

		examount, excommod := nil, ""
		if n := l.AcceptRun(at); n > 2 {
			log.Fatalf("invalid token on line %v: %v", l.Line(), l.Emit())
		} else if n == 1 {
			examount, excommod = lexAmount(l)
			examount.Mul(examount, amount)
		} else if n == 2 {
			examount, excommod = lexAmount(l)
		}

		skipSpace(l)

		note := ""
		if l.Accept(meta) {
			l.Emit()
			l.AcceptRunNot(lineend)
			note = l.Emit()
		}

		items = append(items, &Item{
			status,
			account,
			amount,
			commod,
			examount,
			excommod,
			note,
		})
	}
	return items
}

func lexAmount(l) (amount *bit.Rat, commod string) {
	if l.Accept("$") {
		commod = "$"
	}

	amt := lexAmount
	var err error
	amount = big.NewRat(0, 1)
	amount, err = big.SetString(amt)
	if err != nil {
		log.Fatalf("invalid amount on line %v: %v", l.Line(), amount)
	}

	skipSpace(l)
	if l.AcceptRunNot(whitespace+meta+at) > 0 {
		commod = l.Emit()
	}
	return amount, commod
}

func lexCommod(l *Lexer) string {
}

func lexAmount(l *Lexer) *big.Rat {
	l.AcceptRun(digit + ",")
	l.Accept(".")
	l.AcceptRun(digit)
	if l.Pos > l.Start {
		l.Emit(tokAmount)
	}
}

func lexAccount(l *Lexer) string {
	for {
		r := l.Next()
		nr := l.Peek()
		if isSpace(r) && isSpace(nr) || isNewline(r) || r == EOF {
			l.Backup()
			account := l.Emit()
			return account
		}
	}
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isNewline(r rune) bool {
	return r == '\n' || r == '\r'
}

func skipLine(l *Lexer) {
	l.AcceptRunNot(lineend)
	l.Accept(lineend)
	l.Emit()
}

func skipSpace(l *Lexer) {
	l.AcceptRun(space)
	l.Emit()
}

func skipBlankLines(l *Lexer) int {
	count := 0
	for {
		if l.AcceptRun(space) == 0 && l.Acceptrun(lineend) == 0 {
			l.Reset()
			return count
		}
		l.Emit()
		count++
	}
	return count
}

// Lexer holds the state of the scanner.
type Lexer struct {
	Input   string // the string being scanned
	Pos     int    // current position in the input
	Start   int    // start position of this Token
	width   int    // width of last rune read from input
	Journal []*Trans
}

// Next returns the next rune in the input.
func (l *Lexer) Next() rune {
	if int(l.Pos) >= len(l.Input) {
		l.width = 0
		return EOF
	}
	r, w := utf8.DecodeRuneInString(l.Input[l.Pos:])
	l.width = w
	l.Pos += l.width
	return r
}

func (l *Lexer) AcceptLit(s string) bool {
	if strings.HasPrefix(l.Input[l.Start:], s) {
		l.Pos += len(s)
		return true
	}
	return false
}

// Peek returns but does not consume the next rune in the input.
func (l *Lexer) Peek() rune {
	if int(l.Pos) >= len(l.Input) {
		return EOF
	}
	r := l.Next()
	l.Backup()
	return r
}

// Backup steps back one rune. Can only be called once per call of next.
func (l *Lexer) Backup() { l.Pos -= l.width }

func (l *Lexer) Reset() { l.Pos = l.Start }

// emit passes a Token back to the client.
func (l *Lexer) Emit() string {
	s := l.Input[l.Start:l.Pos]
	l.Start = l.Pos
	return s
}

// Accept consumes the next rune if it's from the valid set and returns true
// if a rune was consumed.
func (l *Lexer) Accept(valid string) bool {
	if strings.IndexRune(valid, l.Next()) >= 0 {
		return true
	}
	l.Backup()
	return false
}

// AcceptNot consumes the next rune if it's not from the invalid set and
// returns true if a rune was consumed.
func (l *Lexer) AcceptNot(invalid string) bool {
	r := l.Next()
	if strings.IndexRune(invalid, r) >= 0 || r == EOF {
		l.Backup()
		return false
	}
	return true
}

// AcceptRun consumes a run of runes from the valid set and returns the number
// of runes accepted.
func (l *Lexer) AcceptRun(valid string) int {
	x := l.Pos
	for strings.IndexRune(valid, l.Next()) >= 0 {
	}
	l.Backup()
	return l.Pos - x
}

// AcceptRunNot consumes a run of runes that aren't from the invalid set and
// returns the number of runes accepted.
func (l *Lexer) AcceptRunNot(invalid string) int {
	x := l.Pos
	r := l.Next()
	for strings.IndexRune(invalid, r) < 0 && r != EOF {
		r = l.Next()
	}
	l.Backup()
	return l.Pos - x
}

// Line reports which line we're on, based on the current position.
func (l *Lexer) Line() int {
	return 1 + strings.Count(l.Input[:l.Pos], "\n")
}
