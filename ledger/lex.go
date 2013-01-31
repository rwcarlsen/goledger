
package ledger

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type tokType int

const (
	tokError tokType = iota
	tokEOF
	tokTransHead
	tokComment // comments
	tokSub
	tokText // unsupported/unrecognized text
)

type token struct {
	typ tokType
	pos int
	val string
}

func (t *token) String() string {
	switch {
	case t.typ == tokEOF:
		return "EOF"
	case t.typ == tokError:
		return t.val
	case len(t.val) > 10:
		return fmt.Sprintf("%.10q...", t.val)
	}
	return fmt.Sprintf("%q", t.val)
}

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name       string    // the name of the input; used only for error reports
	input      string    // the string being scanned
	state      stateFn   // the next lexing function to enter
	pos        int       // current position in the input
	start      int       // start position of this token
	width      int       // width of last rune read from input
	lastPos    int       // position of most recent token returned by nextItem
	tokens     chan token // channel of scanned tokens
}

// lex creates a new scanner for the input string.
func newLexer(name, input string) *lexer {
	l := &lexer{
		name:       name,
		input:      input,
		tokens:      make(chan token),
	}
	go l.run()
	return l
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes a token back to the client.
func (l *lexer) emit(t tokType) {
	l.tokens <- token{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptNot consumes the next rune if it's not from the invalid set.
func (l *lexer) acceptNot(invalid string) bool {
	r := l.next()
	if strings.IndexRune(invalid, r) >= 0 || r == eof {
		l.backup()
		return false
	}
	return true
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// acceptRunNot consumes a run of runes that aren't from the invalid set.
func (l *lexer) acceptRunNot(invalid string) {
	r := l.next()
	for strings.IndexRune(invalid, r) < 0 && r != eof {
	}
	l.backup()
}

// lineNumber reports which line we're on, based on the position of
// the previous token returned by nextItem. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token{tokError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next token from the input.
func (l *lexer) nextItem() token {
	tok := <-l.tokens
	l.lastPos = tok.pos
	return tok
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	fmt.Println("running")
	for l.state = lexLineStart; l.state != nil; {
		fmt.Println("------- changed state -------")
		l.state = l.state(l)
	}
}

/////////////////// state functions ///////////////////////

const (
	digit = "0123456789"
	alpha = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	space = " \t"
	lineend = "\r\n"
	commentDelim = ";"
	eof = -1
)

// lexLineStart looks for a comment or a transaction, it emits everything
// inbetween as tokText
func lexLineStart(l *lexer) stateFn {
	for {
		fmt.Printf("lexing at line %v: %v\n", l.lineNumber(), l.input[l.pos:l.pos+10])
		switch r := l.peek(); {
		case l.input[l.pos:l.pos] == commentDelim: // comment
			if l.pos > l.start {
				l.emit(tokText)
			}
			return lexComment
		case unicode.IsDigit(r): // begin of new transaction
			if l.pos > l.start {
				l.emit(tokText)
			}
			return lexTransHead
		case isSpace(r):
			if l.pos > l.start {
				l.emit(tokText)
			}
			l.acceptRun(space) // ignore prefixed space to sub-indented lines
			l.ignore()
			return lexSub
		default: // each loop iteration begins with 1st char of new line
			l.acceptRunNot(lineend + commentDelim)
		}
		if l.next() == eof {
			break
		}
	}
	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(tokText)
	}
	l.emit(tokEOF)
	return nil
}

// lexComment scans a comment to the end of the line and ignores it
func lexComment(l *lexer) stateFn {
	l.pos += len(commentDelim)
	l.acceptNot(lineend)
	l.emit(tokComment)
	return lexLineStart
}

// lexTransRef scans a transaction
func lexTransHead(l *lexer) stateFn {
	l.acceptRunNot(lineend + commentDelim)
	l.emit(tokTransHead)
	return lexLineStart
}

// lexSub scans a sub/indented line (e.g. a transaction post)
func lexSub(l *lexer) stateFn {
	l.acceptRunNot(lineend + commentDelim)
	l.emit(tokSub)
	return lexLineStart
}

// isSpace reports whether r is a spacing character
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

