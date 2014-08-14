package ledger

import (
	"testing"

	"github.com/rwcarlsen/goledger/lex"
)

const journal1 = `
2010/05/31 Just an example
    * Expenses:Some:Account                $100.00
    * Income:Another:Account
`

func TestLex(t *testing.T) {
	l := lex.New("testlex", journal1, lexStart)

	for tok := range l.Tokens {
		t.Logf("%v: '%v'", tokNames[tok.Type], tok.Val)
	}
}
