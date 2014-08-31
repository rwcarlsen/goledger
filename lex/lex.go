package ledger

import (
	"bytes"
	"fmt"
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

func (t *Trans) String() string {
	var b bytes.Buffer
	tm := t.Date.Format("2006/1/2")
	fmt.Fprintf(&b, "%v %v %v", tm, t.Status, t.Descrip)
	if t.Note != "" {
		fmt.Fprint(&b, "   ;", t.Note)
	}

	for _, item := range t.Items {
		fmt.Fprintf(&b, "\n    %v", item)
	}
	return b.String()
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

func (i *Item) String() string {
	s := ""
	if i.Commod == "$" {
		s = fmt.Sprintf("%v %v        $%v", i.Status, i.Account, i.Amount.FloatString(2))
	} else {
		s = fmt.Sprintf("%v %v        %v %v", i.Status, i.Account, i.Amount.FloatString(2), i.Commod)
	}

	if i.ExAmount != nil {
		if i.ExCommod == "$" {
			s += fmt.Sprintf("@ $%v", i.ExAmount.FloatString(2))
		} else {
			s += fmt.Sprintf("@ %v %v", i.ExAmount.FloatString(2), i.ExCommod)
		}
	}

	if i.Note != "" {
		s += fmt.Sprintf("   ;%v", i.Note)
	}
	return s
}

const (
	space      = " \t"
	lineend    = "\r\n"
	whitespace = space + lineend
	digit      = "0123456789"
	statuss    = "*!"
)

const (
	meta = ";"
	atat = "@@"
	at   = "@"
)

// Decode
func Decode(r io.Reader) ([]*Trans, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	l := &Lexer{
		Input: string(data),
	}
	decode(l)
	return l.Journal, nil
}

func decode(l *Lexer) {
	for {
		switch r := l.Peek(); {
		case unicode.IsDigit(r):
			trans, abort := lexTrans(l)
			if !abort {
				l.Journal = append(l.Journal, trans)
			}
		case r == EOF:
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

	items := lexItems(l)

	// build transaction
	return &Trans{
		Date:    tm,
		Status:  status,
		Descrip: descrip,
		Note:    note,
		Items:   items,
	}, false
}

func lexItems(l *Lexer) []*Item {
	items := []*Item{}
	for {
		if skipBlankLines(l) > 0 {
			break
		} else if l.Peek() == EOF {
			break
		} else if skipSpace(l) == 0 {
			break
		}

		if l.Accept(meta) {
			l.Emit()
			l.AcceptRunNot(lineend)
			if len(items) > 0 {
				items[len(items)-1].Note += l.Emit()
			}
			skipLine(l)
			continue
		}

		status := ""
		if l.Accept(statuss) {
			status = l.Emit()
			skipSpace(l)
		}

		account := lexAccount(l)
		skipSpace(l)

		// primary commod and amount
		amount, commod := lexAmount(l)

		skipSpace(l)

		// exchange commod and amount
		var examount *big.Rat
		excommod := ""
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
			note += l.Emit()
		}

		skipLine(l)

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

func lexNumber(l *Lexer) string {
	l.Accept("+-")
	l.AcceptRun(digit + ",")
	l.Accept(".")
	l.AcceptRun(digit)
	if l.Accept("Ee") {
		l.Accept("+-")
		l.AcceptRun(digit)
	}
	return l.Emit()
}

func lexAmount(l *Lexer) (amount *big.Rat, commod string) {
	if l.Accept("$") {
		commod = "$"
		l.Emit()
	}

	var success bool
	amt := lexNumber(l)
	amount = &big.Rat{}
	if len(amt) > 0 {
		amount, success = amount.SetString(amt)
		if !success {
			log.Fatalf("invalid amount on line %v: %v", l.Line(), amt)
		}
	}

	skipSpace(l)
	if l.AcceptRunNot(whitespace+meta+at) > 0 {
		commod = l.Emit()
	}
	return amount, commod
}

func lexAccount(l *Lexer) string {
	for {
		r := l.Next()
		nr := l.Peek()
		if isSpace(r) && isSpace(nr) || isNewline(r) || r == EOF {
			l.Backup()
			return l.Emit()
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

func skipSpace(l *Lexer) int {
	l.AcceptRun(space)
	return len(l.Emit())
}

func skipBlankLines(l *Lexer) int {
	count := 0
	for {
		l.AcceptRun(space)
		if l.AcceptRun(lineend) == 0 {
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
