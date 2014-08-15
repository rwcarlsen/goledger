package parse

import "github.com/rwcarlsen/goledger/lex"

type StateFn func(p *Parser) StateFn

type Parser struct {
	l      *lex.Lexer
	toks   []lex.Token
	pos    int
	states []StateFn
}

func New(l *lex.Lexer, start StateFn) *Parser {
	p := &Parser{
		l: l,
	}
	go p.run(start)
	return p
}

func (p *Parser) run(start StateFn) {
	p.Push(start)
	for len(p.states) > 0 {
		state := p.pop()
		state = state(p)
		if state != nil {
			p.Push(state)
		}
	}
}

func (p *Parser) Push(fn StateFn) { p.states = append(p.states, fn) }

func (p *Parser) pop() StateFn {
	end := len(p.states) - 1
	fn := p.states[end]
	p.states = p.states[:end]
	return fn
}

func (p *Parser) Next() lex.Token {
	p.fill()
	tok := p.toks[p.pos]
	p.pos++
	return tok
}

func (p *Parser) Backup() {
	if p.pos > 0 {
		p.pos--
	}
}

func (p *Parser) Peek() lex.Token {
	p.fill()
	return p.toks[p.pos+1]
}

// fill adds pulls tokens from the lexer so that the nth token after the
// current one is ready.
func (p *Parser) fill() {
	if p.pos+1 >= len(p.toks) {
		need := p.pos - len(p.toks) + 10
		for i := 0; i < need; i++ {
			tok := <-p.l.Tokens
			blank := lex.Token{}
			if tok == blank {
				panic("eof - decide what to do")
			}
			p.toks = append(p.toks, tok)
		}
	}
}
