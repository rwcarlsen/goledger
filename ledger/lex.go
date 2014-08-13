package ledger

import (
	"unicode"

	"github.com/rwcarlsen/goledger/lex"
)

const (
	tokNewline lex.TokType = iota
	tokIndent              // tab, multispace, etc
	tokDate                //
	tokText                // trans header, comment text, etc.
	tokMeta                // metadata/notes/comments
	tokCleared
	tokMultispace
	tokUnit
	tokCommod
	tokNumber
	tokAt
	tokAtAt
)

var tokNames = map[lex.TokType]string{
	lex.TokError:  "Error",
	lex.TokEOF:    "EOF",
	tokNewline:    "Newline",
	tokIndent:     "Indent",
	tokDate:       "Date",
	tokText:       "Text",
	tokMeta:       "Meta",
	tokCleared:    "Cleared",
	tokMultispace: "Multispace",
}

/////////////////// state functions ///////////////////////

const (
	indent     = " \t"
	lineend    = "\r\n"
	whitespace = indent + lineend
	cleared    = "*" + indent
	digit      = "0123456789"
)

const (
	meta = ";"
	atat = "@@"
	at   = "@"
)

// lexStart looks for a comment or a transaction, it emits everything
// inbetween as tokText
func lexStart(l *lex.Lexer) lex.StateFn {
	switch r := l.Peek(); {
	case string(r) == meta:
		return lexMeta
	case unicode.IsDigit(r):
		return lexTrans
	case isWhitespace(r):
		l.Push(lexStart)
		return lexBlankLine
	case isSpace(r):
		return lexIndent
	case r == lex.EOF:
		l.Emit(lex.TokEOF)
		return nil
	default:
		l.Errorf("Unexpected text on line %v", l.LineNumber())
		l.Push(lexStart)
		return lexSkipLine
	}
}

func lexSkipLine(l *lex.Lexer) lex.StateFn {
	l.AcceptRunNot(lineend)
	l.AcceptRun(lineend)
	l.Ignore()
	return nil
}

func lexBlankLine(l *lex.Lexer) lex.StateFn {
	l.AcceptRun(indent)
	l.Ignore()

	if l.AcceptRun(lineend) == 0 {
		l.Errorf("unexpected non-blank line at line %v", l.LineNumber())
		return lexSkipLine
	}
	return nil
}

func lexDate(l *lex.Lexer) lex.StateFn {
	l.AcceptRunNot(whitespace + meta)
	l.Emit(tokDate)
	return lexCleared
}

func lexCleared(l *lex.Lexer) lex.StateFn {
	if l.AcceptRun(cleared) > 0 {
		l.Emit(tokCleared)
	}
	return lexText
}

func lexText(l *lex.Lexer) lex.StateFn {
	if l.AcceptRunNot(lineend+meta) > 0 {
		l.Emit(tokText)
	}
	return lexStart
}

func lexMeta(l *lex.Lexer) lex.StateFn {
	l.Pos += len(meta)
	l.Emit(tokMeta)
	return lexText
}

func lexLineEnd(l *lex.Lexer) lex.StateFn {
	l.AcceptRun(whitespace)
	l.Emit(tokNewline)
	return lexStart
}

func lexIndent(l *lex.Lexer) lex.StateFn {
	l.AcceptRun(indent)
	l.Emit(tokIndent)
	return lexIndented
}

func lexMultispace(l *lex.Lexer) lex.StateFn {
	l.AcceptRun(indent)
	l.Emit(tokMultispace)
	return lexAmount
}

func lexUnit(l *lex.Lexer) lex.StateFn {
	if l.Accept("$") {
		l.Emit(tokUnit)
	}
	return lexAmount
}

func lexAt(l *lex.Lexer) lex.StateFn {
	if n := l.AcceptRun(at); n == 2 {
		l.Emit(tokAtAt)
	} else if n == 1 {
		l.Emit(tokAt)
	} else if n == 0 {
		if string(l.Peek()) == meta {
			return lexMeta
		} else {
			return lexLineEnd
		}
	} else {
		l.Emit(lex.TokError)
	}
	return lexAmount
}

func lexCommod(l *lex.Lexer) lex.StateFn {
	l.AcceptRun(whitespace)
	l.Ignore()
	if l.AcceptRunNot(whitespace+meta+at) > 0 {
		l.Emit(tokCommod)
	}
	return lexAt
}

func lexAmount(l *lex.Lexer) lex.StateFn {
	l.AcceptRun(digit + ",")
	l.Accept(".")
	l.AcceptRun(digit)
	l.Emit(tokNumber)
	return lexCommod
}

func lexIndented(l *lex.Lexer) lex.StateFn {
	for {
		nr := l.Next()
		nnr := l.Peek()
		l.Backup()
		if isSpace(nr) && isSpace(nnr) {
			return lexMultispace
		}

		if l.AcceptRunNot(whitespace+meta) == 0 {
			if string(l.Peek()) == meta {
				return lexMeta
			} else {
				return lexLineEnd
			}
		}
	}
	return lexText
}

// isSpace reports whether r is a spacing character
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isWhitespace(r rune) bool {
	return isSpace(r) || isLineEnd(r)
}

// isLineEnd
func isLineEnd(r rune) bool {
	return r == '\r' || r == '\n'
}
