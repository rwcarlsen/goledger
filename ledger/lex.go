package ledger

import (
	"unicode"

	"github.com/rwcarlsen/goledger/lex"
)

const (
	tokNewline lex.TokType = iota
	tokBeginTrans
	tokDate //
	tokStatus
	tokPayee
	tokMeta // metadata/notes/comments
	tokText // trans header, comment text, etc.
	tokAccount
	tokUnit
	tokCommod
	tokAmount
	tokAt
	tokAtAt
	tokEndTrans
)

var tokNames = map[lex.TokType]string{
	lex.TokError:  "Error",
	lex.TokEOF:    "EOF",
	tokNewline:    "Newline",
	tokDate:       "Date",
	tokText:       "Text",
	tokMeta:       "Meta",
	tokStatus:     "Status",
	tokPayee:      "Payee",
	tokBeginTrans: "BeginTrans",
	tokEndTrans:   "EndTrans",
	tokAccount:    "Account",
	tokUnit:       "Unit",
	tokCommod:     "Commod",
	tokAmount:     "Amount",
	tokAt:         "At",
	tokAtAt:       "AtAt",
}

/////////////////// state functions ///////////////////////

const (
	indent     = " \t"
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

// lexStart looks for a comment or a transaction, it emits everything
// inbetween as tokText
func lexStart(l *lex.Lexer) lex.StateFn {
	switch r := l.Peek(); {
	case string(r) == meta:
		l.Push(lexStart)
		return lexMeta
	case unicode.IsDigit(r):
		l.Push(lexStart)
		return lexTrans
	case isSpace(r) || isNewline(r):
		l.Push(lexStart)
		return lexBlankLine
	case r == lex.EOF:
		l.Emit(lex.TokEOF)
		return nil
	default:
		l.Errorf("unexpected text on line %v", l.LineNumber())
		l.Push(lexStart)
		return lexSkipLine
	}
}

func lexTrans(l *lex.Lexer) lex.StateFn {
	l.Emit(tokBeginTrans)

	l.Push(lexEndTrans)
	l.Push(lexItems)
	l.Push(lexMeta)
	l.Push(lexPayee)
	l.Push(lexStatus)
	return lexDate
}

func lexEndTrans(l *lex.Lexer) lex.StateFn {
	l.Emit(tokEndTrans)
	return nil
}

func lexDate(l *lex.Lexer) lex.StateFn {
	fail := false
	if l.AcceptRun(digit) == 0 {
		fail = true
	} else if !l.Accept("/") {
		fail = true
	} else if l.AcceptRun(digit) == 0 {
		fail = true
	} else if !l.Accept("/") {
		fail = true
	} else if l.AcceptRun(digit) == 0 {
		fail = true
	}

	if fail {
		l.AcceptRunNot(whitespace + meta)
		l.Errorf("invalid date on line %v", l.LineNumber())
		l.Ignore()
	} else {
		l.Emit(tokDate)
	}
	return nil
}

func lexStatus(l *lex.Lexer) lex.StateFn {
	l.AcceptRun(indent)
	l.Ignore()
	if l.Accept(statuss) {
		l.Emit(tokStatus)
	}
	return nil
}

func lexPayee(l *lex.Lexer) lex.StateFn {
	if l.AcceptRunNot(lineend+meta) > 0 {
		l.Emit(tokPayee)
	}
	return nil
}

func lexItems(l *lex.Lexer) lex.StateFn {
	if l.AcceptRun(indent) == 0 {
		return nil
	} else if string(l.Peek()) == meta {
		l.Push(lexItems)
		return lexMeta
	}

	l.Ignore()

	l.Push(lexItems)
	return lexItem
}

func lexItem(l *lex.Lexer) lex.StateFn {
	l.Push(lexMeta)
	l.Push(lexAmount)
	l.Push(lexAt)
	l.Push(lexAmount)
	l.Push(lexAccount)
	return lexStatus
}

func lexAccount(l *lex.Lexer) lex.StateFn {
	for {
		nr := l.Next()
		nnr := l.Peek()
		if isSpace(nr) && isSpace(nnr) || isNewline(nr) {
			l.Backup()
			l.Emit(tokAccount)
			l.AcceptRun(indent)
			l.Ignore()
			return nil
		}
	}
}

func lexAmount(l *lex.Lexer) lex.StateFn {
	l.AcceptRun(indent)
	l.Ignore()

	if l.Accept("$") {
		l.Emit(tokUnit)
	}

	l.AcceptRun(digit + ",")
	l.Accept(".")
	l.AcceptRun(digit)
	if l.Pos > l.Start {
		l.Emit(tokAmount)
	}
	return lexCommod
}

func lexCommod(l *lex.Lexer) lex.StateFn {
	l.AcceptRun(indent)
	l.Ignore()
	if l.AcceptRunNot(whitespace+meta+at) > 0 {
		l.Emit(tokCommod)
	}
	return nil
}

func lexAt(l *lex.Lexer) lex.StateFn {
	if n := l.AcceptRun(at); n > 2 {
		l.Errorf("invalid token on line %v", l.LineNumber())
		return lexSkipLine
	} else if n == 1 {
		l.Emit(tokAt)
	} else if n == 2 {
		l.Emit(tokAtAt)
	}
	return nil
}

func lexBlankLine(l *lex.Lexer) lex.StateFn {
	l.AcceptRun(indent)
	l.Ignore()

	if l.AcceptRun(lineend) == 0 {
		l.Errorf("unexpected non-blank line at line %v", l.LineNumber())
		return lexSkipLine
	}
	l.Ignore()
	return nil
}

func lexSkipLine(l *lex.Lexer) lex.StateFn {
	l.AcceptRunNot(lineend)
	l.AcceptRun(lineend)
	l.Ignore()
	return nil
}

func lexNewline(l *lex.Lexer) lex.StateFn {
	l.AcceptRun(indent)
	if l.AcceptRun(lineend) == 0 {
		l.Errorf("lexer error - missing expected newline")
	} else {
		l.Emit(tokNewline)
	}
	return nil
}

func lexMeta(l *lex.Lexer) lex.StateFn {
	l.AcceptRun(indent)
	l.Ignore()
	if l.Accept(";") {
		l.Emit(tokMeta)
		l.AcceptRun(indent)
		l.Ignore()
		return lexText
	}
	return lexNewline
}

func lexText(l *lex.Lexer) lex.StateFn {
	l.AcceptRunNot(lineend + meta)
	l.Emit(tokText)
	return lexNewline
}

// isSpace reports whether r is a spacing character
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isNewline
func isNewline(r rune) bool {
	return r == '\r' || r == '\n'
}
