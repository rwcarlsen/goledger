package lex

import (
	"io"
	"strings"
	"unicode/utf8"
)

// Decode
func Decode(r io.Reader) Journal {
	l := &Lexer{
		name:  name,
		Input: input,
	}
	return l
}

// Lexer holds the state of the scanner.
type Lexer struct {
	name  string // the name of the input; used only for error reports
	Input string // the string being scanned
	Pos   int    // current position in the input
	Start int    // start position of this Token
	width int    // width of last rune read from input
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

// Ignore skips over the pending input before this point.
func (l *Lexer) Ignore() { l.Start = l.Pos }

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
