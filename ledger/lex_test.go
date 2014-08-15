package ledger

import (
	"io/ioutil"
	"testing"

	"github.com/rwcarlsen/goledger/lex"
)

const journal1 = `
10/05/31 Just an example
    * Expenses:Some foo:Account                $100.00
    * Income:Another bar:Account
`

func TestLex(t *testing.T) {
	l := lex.New("testlex", journal1, lexStart)
	for tok := range l.Tokens {
		t.Logf("%v: '%v'", tokNames[tok.Type], tok.Val)
	}
}

func BenchmarkHello(b *testing.B) {
	data, err := ioutil.ReadFile("/home/robert/git/money/rwc-finances.ledger")
	if err != nil {
		b.Fatal(err)
	}
	s := string(data)

	for i := 0; i < b.N; i++ {
		l := lex.New("testlex", s, lexStart)
		for _ = range l.Tokens {
		}
	}
}
