
package ledger

import (
	"testing"
	"time"
	"bytes"
	"fmt"
)

func TestTransPrint(t *testing.T) {
	trans := &Trans{
		Date: time.Now(),
		Status: "*",
		Payee: "walgreens pharmacy",
		Comments: []string{"stuff for esther", "another comment"},
	}
	trans.AddPost(&Post{
		Account: "assets:checking",
		Commod: "$",
		Qty: -1.0,
	})
	trans.AddPost(&Post{
		Account: "assets:checking",
		Commod: "$",
		Qty: -12.34,
		Comments: []string{"post comment 1", "post comment 2"},
	})
	trans.AddPost(&Post{
		Account: "expenses:food",
		Commod: "$",
		Qty: 13.34,
	})

	var buf bytes.Buffer

	if err := trans.Print(&buf); err != nil {
		t.Error(err)
	}

	fmt.Println(buf.String())
}
