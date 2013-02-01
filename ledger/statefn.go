
package ledger

import (
	"unicode"
)

/////////////////// state functions ///////////////////////

const (
	indent = " \t"
	lineend = "\r\n"
	whitespace = indent + lineend
	metaDelim = ";"
	eof = -1
)

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// lexLineStart looks for a comment or a transaction, it emits everything
// inbetween as tokText
func lexStart(l *lexer) stateFn {
	switch r := l.peek(); {
	case string(r) == metaDelim:
		return lexMeta
	case isEndOfLine(r):
		return lexLineEnd
	case unicode.IsDigit(r):
		return lexDate
	case isSpace(r):
		return lexIndent
	case r == eof:
		break
	default:
		return lexText
	}
	l.emit(tokEOF)
	return nil
}

func lexDate(l *lexer) stateFn {
	l.acceptRunNot(whitespace + metaDelim)
	l.emit(tokDate)
	return lexText
}

func lexText(l *lexer) stateFn {
	l.acceptRunNot(lineend + metaDelim)
	if l.pos > l.start {
		l.emit(tokText)
	}
	return lexStart
}

func lexMeta(l *lexer) stateFn {
	l.pos += len(metaDelim)
	l.emit(tokMeta)
	return lexText
}

func lexLineEnd(l *lexer) stateFn {
	l.pos += 1
	l.emit(tokNewline)
	return lexStart
}

func lexIndent(l *lexer) stateFn {
	l.acceptRun(indent)
	l.emit(tokIndent)
	return lexText
}

// isSpace reports whether r is a spacing character
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}


