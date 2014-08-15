package ledger

import (
	"testing"

	"github.com/rwcarlsen/goledger/lex"
	"github.com/rwcarlsen/goledger/parse"
)

func TestParse(t *testing.T) {
	l := lex.New("testlex", journal1, lexStart)
	pp := &Parser{}
	p := parse.New(l, pp.Start)
	p.Run()
	for _, trans := range pp.Journal {
		t.Logf("%+v", trans)
	}
}
