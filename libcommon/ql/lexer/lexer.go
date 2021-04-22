package lexer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"selfText/giligili_back/libcommon/ql/token"
)

// eof represents a marker rune for the end of the reader.
var eof = rune(0)

// Lexer is a token scanner
type Lexer struct {
	r   *bufio.Reader
	pos token.Pos //current position
	ch  rune      //current char
}

// NewLexer returns a new instance of Lexer.
func NewLexer(r io.Reader) *Lexer {
	l := &Lexer{r: bufio.NewReader(r), pos: token.Pos{Line: 0, Char: -1}}
	l.read()
	return l
}

// Next returns the next token.
func (p *Lexer) Next() (t token.Token, lit string, pos token.Pos) {
	for p.ch != eof {
		p.skipWS()
		if isLetter(p.ch) {
			return p.ident()
		} else if isQuote(p.ch) {
			return p.quoteIdent()
		} else if isDigital(p.ch) {
			return p.number()
		}
		pos = p.pos
		switch p.ch {
		case '\'':
			return p.string()
		case '+':
			p.read()
			return token.ADD, "", pos
		case '-':
			p.read()
			// if isDigital(p.ch) {
			// 	t, lit, _ = p.number()
			// 	return t, "-" + lit, pos
			// }
			return token.SUB, "", pos
		case '*':
			p.read()
			return token.MUL, "", pos
		case '/': //DIV: "/",
			p.read()
			return token.DIV, "", pos
		case '%': //MOD: "%",
			p.read()
			return token.MOD, "", pos
		case '=': //EQ:  "=",
			p.read()
			if p.ch == '~' {
				p.read()
				return token.EQREGEX, "", pos
			}

			return token.EQ, "", pos
		case '!': //NEQ: "!=",
			p.read()
			if p.ch == '=' {
				p.read()
				return token.NEQ, "", pos
			} else if p.ch == '~' {
				p.read()
				return token.NEQREGEX, "", pos
			}

			return token.ILLEGAL, "!", pos
		case '<': //LT:  "<",
			p.read()
			if p.ch == '=' {
				p.read()
				return token.LTE, "", pos
			} else if p.ch == '>' {
				p.read()
				return token.NEQ, "", pos
			}
			return token.LT, "", pos
		case '>': //GT:  ">",
			p.read()
			if p.ch == '=' {
				p.read()
				return token.GTE, "", pos
			}
			return token.GT, "", pos
		case '(': //LPAREN: "(",
			p.read()
			return token.LPAREN, "", pos
		case ')': //RPAREN: ")",
			p.read()
			return token.RPAREN, "", pos
		case ',': //COMMA:  ",",
			p.read()
			return token.COMMA, "", pos
		case '?': //QUESTION:  "?",
			p.read()
			return token.QUESTION, "", pos
		}

		return token.ILLEGAL, string(p.ch), pos
	}
	return token.EOF, "", pos
}

func (p *Lexer) skipWS() {
	for isWS(p.ch) {
		p.read()
	}
}

func (p *Lexer) ident() (t token.Token, lit string, pos token.Pos) {
	pos = p.pos
	var buf bytes.Buffer
	buf.WriteRune(p.ch)
	p.read()
	for p.ch != eof && (isLetter(p.ch) || isDigital(p.ch) || p.ch == '_') {
		buf.WriteRune(p.ch)
		p.read()
	}

	lit = buf.String()
	if t = token.Lookup(lit); t != token.IDENT {
		return t, "", pos
	}
	return token.IDENT, lit, pos
}

func (p *Lexer) quoteIdent() (t token.Token, lit string, pos token.Pos) {
	pos = p.pos
	var buf bytes.Buffer
	//skip `
	p.read()
	for p.ch != eof && !isQuote(p.ch) {
		buf.WriteRune(p.ch)
		p.read()
	}
	if isQuote(p.ch) {
		//skip `
		p.read()
	} else {
		return token.ILLEGAL, buf.String(), p.pos
	}

	return token.QIDENT, buf.String(), pos
}

func (p *Lexer) number() (t token.Token, lit string, pos token.Pos) {
	t = token.INTEGER
	pos = p.pos
	var buf bytes.Buffer
	buf.WriteRune(p.ch)
	p.read()

	for p.ch != eof {
		if t == token.INTEGER {
			if isDigital(p.ch) {
				buf.WriteRune(p.ch)
				p.read()
			} else if p.ch == '.' {
				t = token.FLOAT
				buf.WriteRune(p.ch)
				p.read()
			} else {
				return t, buf.String(), pos
			}
		} else if t == token.FLOAT {
			if isDigital(p.ch) {
				buf.WriteRune(p.ch)
				p.read()
			} else {
				return t, buf.String(), pos
			}
		}
	}

	return t, buf.String(), pos
}

func (p *Lexer) string() (t token.Token, lit string, pos token.Pos) {
	t = token.STRING
	pos = p.pos
	var buf bytes.Buffer
	buf.WriteRune(p.ch)
	p.read()

	for p.ch != eof {
		if p.ch < 0 {
			return token.ILLEGAL, buf.String(), p.pos
		} else if p.ch == '\'' {
			buf.WriteRune(p.ch)
			p.read()
			return t, buf.String(), pos
		}

		buf.WriteRune(p.ch)
		p.read()
	}
	return token.ILLEGAL, buf.String(), p.pos
}

// read reads the next rune from the buffered reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (p *Lexer) read() {
	p.pos.Char++
	ch, _, _ := p.r.ReadRune()
	p.ch = ch
}

//unread places the previously read rune back on the reader.
// func (p *Lexer) unread() {
// 	p.pos.Char--
// 	_ = p.r.UnreadRune()
// }

// func (p *Lexer) peek() rune {
// 	ch, _, _ := p.r.ReadRune()
// 	p.r.UnreadRune()
// 	return ch
// }

func (p *Lexer) match(x rune) error {
	if p.ch == x {
		p.read()
	}
	return fmt.Errorf("expecting %s ,found %s", string(x), string(p.ch))
}

func isWS(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigital(ch rune) bool {
	return (ch >= '0' && ch <= '9')

}

func isQuote(ch rune) bool {
	return ch == '`'
}
