
package ledger

import (
	"io"
	"fmt"
	"time"
	"text/tabwriter"
	"bytes"
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

const dateFmt = "01/02/2006"

var indent = bytes.Repeat([]byte(" "), tabwidth)

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

func (t *Trans) AddPost(p *Post) {
	p.trans = t
	t.Posts = append(t.Posts, p)
}

func (t *Trans) Print(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%v %v %v\n", t.Date.Format(dateFmt), t.Status, t.Payee)
	if err != nil {
		return err
	}

	for _, c := range t.Comments {
		if _, err := fmt.Fprintf(w, "%s; %v\n", indent, c); err != nil {
			return err
		}
	}

	tw := tabwriter.NewWriter(w, minwidth, tabwidth, padding, padchar, 0)
	for _, p := range t.Posts {
		if err := p.Print(tw); err != nil {
			return err
		}
	}

	return tw.Flush()
}

type Post struct {
	trans *Trans
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
	if _, err := w.Write(indent); err != nil {
		return err
	}

	if p.status != "" {
		fmt.Fprintf(w, "%v ", p.status)
	}
	fmt.Fprintf(w, "%v\t%v%v\n", p.Account, p.Commod, p.Qty)

	for _, c := range p.Comments {
		if _, err := fmt.Fprintf(w, "%s%s; %v\n", indent, indent, c); err != nil {
			return err
		}
	}
	return nil
}

