package ledger

import (
	"errors"

	"github.com/rwcarlsen/goledger/lex"
)

type node struct {
	tok      lex.Token
	children []*node
}

type parser struct {
	tokens chan lex.Token
	bucket []lex.Token
	tree   *node
	Journ  *Journal
}

func (p *parser) Next() lex.Token {
	return <-p.tokens
}

func (p *parser) Push(tok lex.Token) {
	p.bucket = append(p.bucket, tok)
}

func (p *parser) Pop() lex.Token {
	end := len(p.bucket) - 1
	tok := p.bucket[end]
	p.bucket = p.bucket[:end]
	return tok
}

type parseFn func(p *parser) (parseFn, node, error)

func parse(l *lex.Lexer) (*Journal, error) {
	var err error
	p := &parser{
		tokens: l.Tokens,
		bucket: make([]lex.Token, 0),
	}
	stack := []parseFn{parseOuter}
	root := lex.Token{99, 0, "ROOT"}
	treestack := []*node{&node{root}}
	for len(stack) > 0 {
		end := len(stack) - 1
		state := stack[end]
		stack = stack[:end]

		if state, node, err = state(p); err != nil {
			return nil, error
		} else if state != nil {
			stack = append(stack, state)
		}
		if node != nil {
			p.tree
		}
	}
	return j, nil
}

func parseOuter(p *parser) (parseFn, error) {
	switch tok := p.Next(); tok {
	case tokNewline, tokIndent, tokText, tokMeta: // ignore
		return parseLine, nil
	case tokDate:
		p.Push(tok)
		return parseTrans, nil
	case lex.TokEOF:
		return nil, nil
	case lex.Error:
		return nil, errors.New(tok.Val)
	}
	panic("not reached")
}

func parseTrans(p *parser) (parseFn, error) {
	date := p.Pop()

	switch tok := p.Next(); tok {
	case tokText:
		p.Push(tok)
		return parseText, nil
	case tokNewline:
	}
}
