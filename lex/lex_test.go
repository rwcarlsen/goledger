package ledger

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

const journal1 = `
10/05/31 Just an example ; a note
    * Expenses:Some foo:Account                $100.00
    * Income:Another bar:Account
`

const journal2 = `
10/05/31 Just an example
    Expenses:Some foo:Account                $100.00
    Income:Another bar:Account               $-100
`

const journal3 = `
10/05/31 Just an example
	; late note
    Expenses:Some foo:Account                $100.00
    Income:Another bar:Account               $-100
`

func TestLex(t *testing.T) {
	journal, err := Decode(bytes.NewBufferString(journal1))
	if err != nil {
		t.Fatal(err)
	}
	for _, trans := range journal {
		fmt.Printf("%+v\n", trans)
	}

	journal, err = Decode(bytes.NewBufferString(journal2))
	if err != nil {
		t.Fatal(err)
	}
	for _, trans := range journal {
		fmt.Printf("%+v\n", trans)
	}

	journal, err = Decode(bytes.NewBufferString(journal3))
	if err != nil {
		t.Fatal(err)
	}
	for _, trans := range journal {
		fmt.Printf("%+v\n", trans)
	}
}

func BenchmarkHello(b *testing.B) {
	data, err := ioutil.ReadFile("/home/robert/git/money/rwc-finances.ledger")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		_, err := Decode(bytes.NewBuffer(data))
		if err != nil {
			b.Fatal(err)
		}
	}
}
