
package ledger

import (
	"os"
	"io"
	"fmt"
	"text/tabwriter"
	"strings"
)

const (
	minwidth = 8
	tabwidth = 4
	padding = 2
	padchar = ' '
)

const (
	TransEntry = "Transaction"
	UnknownEntry = "Unknown"
)

var indent = strings.Repeat(" ", tabwidth)

type Entry interface {
	Print(io.Writer) error
	Type() string
}

type Trans struct {
	Date time.Time
	Status string
	Payee string
	Comments []string
	Posts []*Post
}

func (t *Trans) Type() string {
	return TransEntry
}

func (t *Trans) Print(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%v\t%v\t%v\n", t.Date.Format(dateFmt), t.Status, t.Payee)
	if err != nil {
		return err
	}

	for _, c := range t.Comments {
		if _, err := fmt.Fprintf(w, "%v; %v\n", indent, c); err != nil {
			return err
		}
	}

	tw := tabwriter.NewWriter(w, minwidth, tabwidth, padding, padchar, 0)
	for _, p := range t.Posts {
		if _, err := tw.Write(indent); err != nil {
			return err
		} else if err := p.Print(tw); err != nil {
			return err
		} else _, err := fmt.Fprint(tw, "\n"); err != nil {
			return err
		}
	}

	return tw.Flush()
}

type Post struct {
	trans *Transaction
	Account string
	status string
	Commod string
	Qty float64
	Comments []string
}

func (p *Post) Status() string {
	if p.status != "" {
		return p.status
	}
	return p.trans.Status
}

func (p *Post) Print(w io.Writer) error {
	fmt.Fprintf(w, "%v\t%v\t%v%v")

	return nil
}

