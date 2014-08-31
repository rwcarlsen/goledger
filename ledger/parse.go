package ledger

import (
	"fmt"
	"math/big"
	"time"

	"github.com/rwcarlsen/goledger/lex"
	"github.com/rwcarlsen/goledger/parse"
)

type Parser struct {
	Journal   []*Trans
	currTrans *Trans
	currItem  *Item
	currNote  string
	currAmt   *big.Rat
}

func (a *Parser) pNote(p *parse.Parser) parse.StateFn {
	switch tok := p.Peek(); tok.Type {
	case tokMeta:
		p.Next()
		tok = p.Next()
		a.currNote = tok.Val
		if tok = p.Next(); tok.Type != tokNewline {
			panic(fmt.Sprintf("unexpected token %v: '%v'", tokNames[tok.Type], tok.Val))
		}
		return nil
	default:
		return nil
	}
}

func (a *Parser) Start(p *parse.Parser) parse.StateFn {
	fmt.Println("Start")
	a.currNote = ""
	switch tok := p.Peek(); tok.Type {
	case lex.TokEOF:
		return nil
	case tokBeginTrans:
		fmt.Printf("type %v: %+v\n", tokNames[tok.Type], tok)
		return a.pTrans
	case tokNewline:
		return a.Start
	case tokMeta:
		p.Push(a.Start)
		return a.pNote
	default:
		panic(fmt.Sprintf("unexpected token %v: '%v'", tokNames[tok.Type], tok.Val))
	}
}

func (a *Parser) pTrans(p *parse.Parser) parse.StateFn {
	fmt.Println("Trans")
	tok := p.Next()
	if tok.Type != tokBeginTrans {
		panic(fmt.Sprintf("unexpected token %v: '%v'", tokNames[tok.Type], tok.Val))
	}

	a.currTrans = &Trans{}
	p.Push(a.pEndTrans)
	return a.pHeader
}
func (a *Parser) pEndTrans(p *parse.Parser) parse.StateFn {
	fmt.Println("EndTrans")
	a.Journal = append(a.Journal, a.currTrans)
	return a.Start
}

func (a *Parser) pItem(p *parse.Parser) parse.StateFn {
	fmt.Println("Item")
	tok := p.Next()

	a.currItem = &Item{}

	// check for status
	if tok.Type == tokStatus {
		fmt.Println("status")
		a.currItem.Status = tok.Val
		tok = p.Next()
	}

	// check for account (required)
	if tok.Type == tokAccount {
		fmt.Println("account: ", tok.Val)
		a.currItem.Account = tok.Val
		tok = p.Next()
	} else {
		panic(fmt.Sprintf("unexpected token %v: '%v'", tokNames[tok.Type], tok.Val))
	}

	p.Push(a.pEndItem)
	p.Push(a.pNote)
	p.Push(a.pExchange)
	return a.pAmount
}

func (a *Parser) pEndItem(p *parse.Parser) parse.StateFn {
	fmt.Println("enditem")
	a.currItem.Note = a.currNote
	a.currNote = ""
	a.currTrans.Items = append(a.currTrans.Items, a.currItem)
	if tok := p.Peek(); tok.Type == tokEndTrans {
		return nil
	}
	return a.pItem
}

func (a *Parser) pAmount(p *parse.Parser) parse.StateFn {
	fmt.Println("amount")
	tok := p.Peek()

	if tok.Type == tokUnit {
		tok = p.Next()
		a.currItem.Commod = tok.Val
	}

	switch tok := p.Peek(); tok.Type {
	case tokNewline, tokEndTrans:
		p.Next()
	case tokAmount:
		tok = p.Next()
		rat := big.NewRat(0, 1)
		var success bool
		a.currItem.Amount, success = rat.SetString(tok.Val)
		if !success {
			panic("invalid amount")
		}

		if tok = p.Peek(); tok.Type == tokCommod {
			tok = p.Next()
			a.currItem.Commod = tok.Val
		}
	}
	return nil
}

func (a *Parser) pExAmount(p *parse.Parser) parse.StateFn {
	fmt.Println("examount")
	tok := p.Peek()

	if tok.Type == tokUnit {
		tok = p.Next()
		a.currItem.ExCommod = tok.Val
	}

	switch tok := p.Peek(); tok.Type {
	case tokNewline, tokEndTrans:
		p.Next()
	case tokAmount:
		tok = p.Next()
		rat := big.NewRat(0, 1)
		var success bool
		a.currItem.ExAmount, success = rat.SetString(tok.Val)
		if !success {
			panic("invalid amount")
		}
		if tok = p.Peek(); tok.Type == tokCommod {
			tok = p.Next()
			a.currItem.ExCommod = tok.Val
		}
	}
	return nil
}

func (a *Parser) pExchange(p *parse.Parser) parse.StateFn {
	tok := p.Peek()
	if tok.Type != tokAt || tok.Type != tokAtAt {
		return a.pExAmount
	}
	return nil
}

func (a *Parser) pHeader(p *parse.Parser) parse.StateFn {
	tok := p.Next()

	// check for date (required)
	if tok.Type == tokDate {
		var err error
		a.currTrans.Date, err = time.Parse("06/1/2", tok.Val)
		if err != nil {
			panic(err.Error())
		}
		tok = p.Next()
	} else {
		panic(fmt.Sprintf("unexpected token %v: '%v'", tokNames[tok.Type], tok.Val))
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
		panic(fmt.Sprintf("unexpected token %v: '%v'", tokNames[tok.Type], tok.Val))
	}

	// check for note
	if tok.Type == tokMeta {
		a.currTrans.Note = p.Next().Val
		tok = p.Next()
	}

	if tok.Type != tokNewline {
		panic(fmt.Sprintf("unexpected token %v: '%v'", tokNames[tok.Type], tok.Val))
	}
	return a.pItem
}
