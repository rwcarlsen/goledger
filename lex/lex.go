package lex

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type TokType int

const (
	TokError TokType = -1*iota - 1
	TokEOF
)

const EOF = -1

type Token struct {
	Type TokType
	Pos  int
	Val  string
}

func (t *Token) String() string {
	switch {
	case t.Type == TokEOF:
		return "EOF"
	case t.Type == TokError:
		return t.Val
	case len(t.Val) > 10:
		return fmt.Sprintf("%.10q...", t.Val)
	}
	return fmt.Sprintf("%q", t.Val)
}

// StateFn represents the state of the scanner as a function that returns the next state.
type StateFn func(*Lexer) StateFn

// Lexer holds the state of the scanner.
type Lexer struct {
	name   string     // the name of the input; used only for error reports
	Input  string     // the string being scanned
	states []StateFn  // the next lexing function to enter
	Pos    int        // current position in the input
	Start  int        // start position of this Token
	width  int        // width of last rune read from input
	Tokens chan Token // channel of scanned Tokens
}

// New creates a new scanner for the input string and begins lexing imediately,
// concurrently.
func New(name, input string, start StateFn) *Lexer {
	l := &Lexer{
		name:   name,
		Input:  input,
		Tokens: make(chan Token, 100),
	}
	go l.run(start)
	return l
}

// run runs the state machine for the Lexer.
func (l *Lexer) run(start StateFn) {
	l.Push(start)
	for len(l.states) > 0 {
		state := l.pop()
		state = state(l)
		if state != nil {
			l.Push(state)
		}
	}
	close(l.Tokens)
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
func (l *Lexer) Backup() {
	l.Pos -= l.width
}

// emit passes a Token back to the client.
func (l *Lexer) Emit(t TokType) {
	tok := Token{t, l.Start, l.Input[l.Start:l.Pos]}
	l.Tokens <- tok
	l.Start = l.Pos
}

// Ignore skips over the pending input before this point.
func (l *Lexer) Ignore() {
	l.Start = l.Pos
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

// LineNumber reports which line we're on, based on the current position.
func (l *Lexer) LineNumber() int {
	return 1 + strings.Count(l.Input[:l.Pos], "\n")
}

// Errorf returns an error Token.
func (l *Lexer) Errorf(format string, args ...interface{}) StateFn {
	l.Tokens <- Token{TokError, l.Start, fmt.Sprintf(format, args...)}
	l.Ignore()
	return nil
}

func (l *Lexer) Push(s StateFn) {
	l.states = append(l.states, s)
}

func (l *Lexer) pop() StateFn {
	last := len(l.states) - 1
	s := l.states[last]
	l.states = l.states[:last]
	return s
}
