package ledger

import (
	"fmt"
	"math/big"
	"time"

	"github.com/rwcarlsen/goledger/lex"
	"github.com/rwcarlsen/goledger/parse"
)

type Trans struct {
	Date    time.Time
	Status  string
	Descrip string
	Items   []*Item
	Note    string
}

type Item struct {
	Status   string
	Account  string
	Amount   *big.Rat
	Commod   string
	ExAmount *big.Rat
	ExCommod string
	Note     string
}

type Parser struct {
	Journal   []*Trans
	currTrans *Trans
	currItem  *Item
}

func (a *Parser) Start(p *parse.Parser) parse.StateFn {
	switch tok := p.Peek(); tok.Type {
	case lex.TokEOF:
		return nil
	case tokBeginTrans:
		fmt.Printf("type %v: %+v\n", tokNames[tok.Type], tok)
		return a.pTrans
	case tokNewline:
		return a.Start
	default:
		panic("unexpected token")
	}
}

func (a *Parser) pTrans(p *parse.Parser) parse.StateFn {
	tok := p.Next()
	fmt.Printf("type %v: %+v\n", tokNames[tok.Type], tok)
	if tok.Type != tokBeginTrans {
		panic("unexpected token")
	}

	a.currTrans = &Trans{}
	p.Push(a.pEndTrans)
	return a.pHeader
}
func (a *Parser) pEndTrans(p *parse.Parser) parse.StateFn {
	a.Journal = append(a.Journal, a.currTrans)
	return a.Start
}

func (a *Parser) pItem(p *parse.Parser) parse.StateFn {
	var err error
	tok := p.Next()

	item := &Item{}

	// check for status
	if tok.Type == tokStatus {
		item.Status = tok.Val
		tok = p.Next()
	}

	// check for account (required)
	if tok.Type == tokAccount {
		item.Account = tok.Val
		tok = p.Next()
	} else {
		panic("unexpected token")
	}

	// check for unit
	if tok.Type == tokUnit {
		item.Commod = tok.Val
		tok = p.Next()
	}

	// check for amount (required)
	if tok.Type == tokAmount {
		rat := big.NewRat(0, 0)
		var success bool
		item.Amount, success = rat.SetString(tok.Val)
		if !success {
			panic(err.Error())
		}
		tok = p.Next()
	} else {
		panic("unexpected token")
	}

	// check for commod
	if tok.Type == tokCommod {
		item.Commod = tok.Val
		tok = p.Next()
	}

	// check for exchange rate
	if tok.Type == tokAt || tok.Type == tokAtAt {
		// check for unit
		if tok.Type == tokUnit {
			item.ExCommod = tok.Val
			tok = p.Next()
		}

		// check for amount (required)
		if tok.Type == tokAmount {
			rat := big.NewRat(0, 0)
			var success bool
			item.ExAmount, success = rat.SetString(tok.Val)
			if !success {
				panic(err.Error())
			}
			tok = p.Next()
		} else {
			panic("unexpected token")
		}

		// check for commod
		if tok.Type == tokCommod {
			item.ExCommod = tok.Val
			tok = p.Next()
		}
	}

	// check for note
	if tok.Type == tokMeta {
		item.Note = p.Next().Val
		tok = p.Next()
	}

	if tok.Type != tokNewline {
		panic("unexpected token")
	}

	a.currTrans.Items = append(a.currTrans.Items, item)
	if tok.Type == tokEndTrans {
		return nil
	} else {
		return a.pItem
	}
}

func (a *Parser) pHeader(p *parse.Parser) parse.StateFn {
	tok := p.Next()

	// check for date (required)
	if tok.Type == tokDate {
		var err error
		a.currTrans.Date, err = time.Parse("__06/_1/_2", tok.Val)
		if err != nil {
			panic(err.Error())
		}
		tok = p.Next()
	} else {
		panic("unexpected token")
	}

	// check for status
	if tok.Type == tokStatus {
		a.currTrans.Status = tok.Val
		tok = p.Next()
	}

	// check for payee (required)
	if tok.Type == tokPayee {
		a.currTrans.Descrip = tok.Val
		tok = p.Next()
	} else {
		panic("unexpected token")
	}

	// check for note
	if tok.Type == tokMeta {
		a.currTrans.Note = p.Next().Val
		tok = p.Next()
	}

	if tok.Type != tokNewline {
		panic("unexpected token")
	}
	return a.pItem
}
