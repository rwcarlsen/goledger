
package ledger

import (
	"testing"
	"time"
	"bytes"
	"fmt"

	"github.com/rwcarlsen/goledger/lex"
)

func trans1() *Trans {
	trans := &Trans{
		Date: time.Now(),
		Status: "*",
		Payee: "walgreens pharmacy",
		Comments: []string{"stuff for esther", "another comment"},
	}
	trans.AddPost(&Post{
		Account: "assets:checking",
		Value: &Price{"$", true, -1.0},
	})
	trans.AddPost(&Post{
		Account: "assets:checking",
		Value: &Price{"$", true, -12.34},
		Comments: []string{"post comment 1", "post comment 2"},
	})
	trans.AddPost(&Post{
		Account: "expenses:food",
		Value: &Price{"$", true, 13.34},
	})
	return trans
}

const trans2 = `
2009/05/14  * Gas Station
	; blablabla a transaction comment
    Assets:Westmark Checking                  $-5.32
    ; used a debit card to pay
    Expenses:Transportation:Gas
    ; topped of tank
`

func DISABLE_TestTransPrint(t *testing.T) {
	trans := trans1()

	var buf bytes.Buffer
	if err := trans.Print(&buf); err != nil {
		t.Error(err)
	}

	fmt.Println(buf.String())
}

func TestJournalLex(t *testing.T) {
	l := lex.New("trans2", trans2, lexStart)
	for tok := range l.Tokens {
		t.Logf("------ token %v -------\n'''%v'''\n", tokNames[tok.Type], tok.Val)
	}
}
