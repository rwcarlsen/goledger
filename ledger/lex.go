
package ledger

import (
	"io"
)

type tokType int

const (
	tokDate tokType = iota
	tokComment
	tokStatus // e.g. cleared (*)
	tokName // e.g. transaction description or account name
	tokCommod
	tokQty
	tokEOF
)

type token struct {
	typ tokType
	val string
}

func (t *token) String() string {
	switch t.typ {
	case tokDate:
		return t.val
	case tokComment:
		return t.val
	// ... etc
	}
}

type lexer {
	
}

type lexFunc func(*lexer) lexFunc

func Parse(r io.Reader) ([]error, *Journal) {
	
}

