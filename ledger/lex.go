
package ledger

import (
	"unicode"
	"github.com/rwcarlsen/goledger/lex"
)

const (
	tokNewline lex.TokType = iota
	tokIndent // tab, multispace, etc
	tokDate // 
	tokText // trans header, comment text, etc.
	tokMeta // comments
)

var tokNames = map[lex.TokType]string{
	lex.TokError: "Error",
	lex.TokEOF: "EOF",
	tokNewline: "Newline",
	tokIndent: "Indent",
	tokDate: "Date",
	tokText: "Text",
	tokMeta: "Meta",
}

/////////////////// state functions ///////////////////////

const (
	indent = " \t"
	lineend = "\r\n"
	whitespace = indent + lineend
	metaDelim = ";"
)

// lexStart looks for a comment or a transaction, it emits everything
// inbetween as tokText
func lexStart(l *lex.Lexer) lex.StateFn {
	switch r := l.Peek(); {
	case string(r) == metaDelim:
		return lexMeta
	case isEndOfLine(r):
		return lexLineEnd
	case unicode.IsDigit(r):
		return lexDate
	case isSpace(r):
		return lexIndent
	case r == lex.EOF:
		break
	default:
		return lexText
	}
	l.Emit(lex.TokEOF)
	return nil
}

func lexDate(l *lex.Lexer) lex.StateFn {
	l.AcceptRunNot(whitespace + metaDelim)
	l.Emit(tokDate)
	return lexText
}

func lexText(l *lex.Lexer) lex.StateFn {
	l.AcceptRunNot(lineend + metaDelim)
	if l.Pos > l.Start {
		l.Emit(tokText)
	}
	return lexStart
}

func lexMeta(l *lex.Lexer) lex.StateFn {
	l.Pos += len(metaDelim)
	l.Emit(tokMeta)
	return lexText
}

func lexLineEnd(l *lex.Lexer) lex.StateFn {
	l.Pos += 1
	l.Emit(tokNewline)
	return lexStart
}

func lexIndent(l *lex.Lexer) lex.StateFn {
	l.AcceptRun(indent)
	l.Emit(tokIndent)
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


