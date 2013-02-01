
package ledger

import (
	"io"
	"fmt"
	"time"
	"text/tabwriter"
	"bytes"
	"strings"

	"github.com/rwcarlsen/goledger/lex"
)

const (
	Minwidth = 4
	Tabwidth = 4
	Padding = 2
	Padchar = ' '
)

const (
	TransEntry = "Transaction"
	UnknownEntry = "Unknown"
)

const dateFmt = "01/02/2006"

type Journal struct {
	Entries []Entry
}

func (j *Journal) Add(e ...Entry) {
	j.Entries = append(j.Entries, e...)
}

func Decode(name string, data []byte) (*Journal, error) {
	l := lex.New(name, trans2, lexStart)
	return parse(l)
}

type Entry interface {
	Print(io.Writer) error
	Type() string
}

type MiscEntry struct {
	Content string
}

func (e *MiscEntry) Print(w io.Writer) error {
	_, err := fmt.Fprint(w, e.Content)
	return err
}

func (e *MiscEntry) Type() string {
	return UnknownEntry
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
	indent := strings.Repeat(" ", Tabwidth)
	_, err := fmt.Fprintf(w, "%v %v %v\n", t.Date.Format(dateFmt), t.Status, t.Payee)
	if err != nil {
		return err
	}

	for _, c := range t.Comments {
		if _, err := fmt.Fprintf(w, "%s; %v\n", indent, c); err != nil {
			return err
		}
	}

	tw := tabwriter.NewWriter(w, Minwidth, Tabwidth, Padding, Padchar, 0)
	for _, p := range t.Posts {
		s := strings.Replace(p.String(), "\n", "\n" + indent, -1)
		if _, err := fmt.Fprintf(tw, "%s%s\n", indent, s); err != nil {
			return err
		}
	}

	return tw.Flush()
}

type Post struct {
	trans *Trans
	Account string
	status string
	Value *Price
	Comments []string
}

func (p *Post) Status() string {
	if p.status != "" {
		return p.status
	}
	return p.trans.Status
}

func (p *Post) String() string {
	var buf bytes.Buffer
	if p.status != "" {
		fmt.Fprintf(&buf, "%s %s\t%s", p.status, p.Account, p.Value)
	} else {
		fmt.Fprintf(&buf, "%s\t%s", p.Account, p.Value)
	}

	indent := strings.Repeat(" ", Tabwidth)
	for _, c := range p.Comments {
		fmt.Fprintf(&buf, "\n%s; %v", indent, c)
	}
	return buf.String()
}

type Price struct {
	Commod string
	Prefix bool // is commodity prefix or postfix
	Qty float64
}

func (p *Price) String() string {
	if p.Prefix {
		return fmt.Sprintf("%v%v", p.Commod, p.Qty)
	}
	return fmt.Sprintf("%v %v", p.Qty, p.Commod)
}

