package parser

import (
	"fmt"
	"strconv"

	"selfText/giligili_back/libcommon/ql/ast"
	"selfText/giligili_back/libcommon/ql/lexer"
	"selfText/giligili_back/libcommon/ql/token"
)

type tok struct {
	t   token.Token
	lit string
	pos token.Pos
}

func (p tok) String() string {
	return fmt.Sprintf("<%s,%s,%s>", p.t, p.lit, p.pos)
}

// Parser parses tokens form the lexer returns an AST
type Parser struct {
	input  *lexer.Lexer
	tokens []tok
	k      int // LL(k) buf size
	i      int // current index of token in the buf
}

// NewParser create an instance of parser
func NewParser(input *lexer.Lexer) *Parser {
	k := 2
	p := &Parser{
		input:  input,
		tokens: make([]tok, k),
		k:      k,
	}
	for i := 0; i < k; i++ {
		p.read()
	}
	return p
}

//Select statement
func (p *Parser) Select() (*ast.Select, error) {
	var err error
	selectStmt := &ast.Select{}
	if err = p.match(token.SELECT); err != nil {
		return nil, err
	}
	if selectStmt.Fields, err = p.fields(); err != nil {
		return nil, err
	}
	if p.token(1).t == token.FROM {
		if selectStmt.Tables, err = p.from(); err != nil {
			return nil, err
		}
	}
	if p.token(1).t == token.WHERE {
		if selectStmt.Where, err = p.where(); err != nil {
			return nil, err
		}
	}
	if p.token(1).t == token.GROUP && p.token(2).t == token.BY {
		if selectStmt.GroupFields, err = p.groupBy(); err != nil {
			return nil, err
		}
	}
	if p.token(1).t == token.ORDER && p.token(2).t == token.BY {
		if selectStmt.SortFields, err = p.orderBy(); err != nil {
			return nil, err
		}
	}
	if p.token(1).t == token.LIMIT {
		selectStmt.Limit, err = p.limit()
	}
	if p.token(1).t == token.OFFSET {
		selectStmt.Offset, err = p.offset()
	}
	return selectStmt, nil
}

func (p *Parser) Delete() (*ast.Delete, error) {
	var (
		deleteStmt = &ast.Delete{}
		err error
	)
	if err = p.match(token.DELETE); err != nil {
		return nil, err
	}
	if p.token(1).t == token.FROM {
		if deleteStmt.Tables, err = p.from(); err != nil {
			return nil, err
		}
	}
	if p.token(1).t == token.WHERE {
		if err = p.match(token.WHERE); err != nil {
			return nil, err
		}

		if err = p.match(token.PRECISE); err == nil {
			deleteStmt.HasPreciseCondition = true
		}

		if deleteStmt.Where, err = p.expr(); err != nil {
			return nil, err
		}
	}
	return deleteStmt, nil
}

func (p *Parser) Drop() (*ast.Drop, error) {
	var (
		dropStmt = &ast.Drop{}
		err error
	)
	if err = p.match(token.DROP); err != nil {
		return nil, err
	}
	if p.token(1).t == token.TABLE {
		if dropStmt.Tables, err = p.table(); err != nil {
			return nil, err
		}
	}
	if p.token(1).t == token.WHERE {
		if err = p.match(token.WHERE); err != nil {
			return nil, err
		}

		if err = p.match(token.PRECISE); err == nil {
			dropStmt.HasPreciseCondition = true
		}

		if dropStmt.Where, err = p.expr(); err != nil {
			return nil, err
		}
	}
	return dropStmt, nil
}

//Filter returns a condition expression
func (p *Parser) Filter() (ast.Expr, error) {
	return p.expr()
}

//Expression return an expression. support expr libcommon
func (p *Parser) Expression() (root ast.Expr, err error) {
	return p.expr()
}

func (p *Parser) fields() (ast.Fields, error) {
	var (
		fields     = ast.Fields{}
		isDistinct = func(expr ast.Expr) bool {
			_, ok := expr.(*ast.Distinct)
			return ok
		}
		haveDistinct = false
	)

	field, err := p.field()
	if err != nil {
		return nil, err
	}
	if isDistinct(field.Expr) {
		haveDistinct = true
	}
	fields = append(fields, field)

	for p.token(1).t == token.COMMA {
		if err = p.match(token.COMMA); err != nil {
			return nil, err
		}
		field, err = p.field()
		if err != nil {
			return nil, err
		}
		if haveDistinct && isDistinct(field.Expr) {
			return nil, newSyntaxError(p.token(1))
		}
		fields = append(fields, field)
	}

	return fields, nil
}

func (p *Parser) field() (field *ast.Field, err error) {
	field = &ast.Field{}
	tok := p.token(1)
	if tok.t == token.MUL {
		p.match(token.MUL)
		field.Expr = &ast.AllField{}
	} else if tok.t == token.DISTINCT {
		field.Expr, err = p.distinctField()
		if err != nil {
			return nil, err
		}
	} else {
		field, err = p.readField()
		if err != nil {
			return nil, err
		}
	}

	return field, nil
}

func (p *Parser) readField() (field *ast.Field, err error) {
	field = &ast.Field{}
	if p.token(1).t == token.MUL {
		p.match(token.MUL)
		field.Expr = &ast.AllField{}
	} else {
		tok := p.token(1)
		if tok.t == token.IDENT {
			field.Expr, err = p.ident()
			if err != nil {
				return nil, err
			}
		} else if tok.t == token.QIDENT {
			field.Expr, err = p.quoteIdent()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, newSyntaxError(tok)
		}

		if p.token(1).t == token.AS {
			if err = p.match(token.AS); err != nil {
				return nil, err
			}
			as := p.token(1)
			if err = p.match(token.IDENT); err != nil {
				return nil, err
			}
			field.AS = as.lit
		}
	}
	return field, nil
}

func (p *Parser) distinctField() (*ast.Distinct, error) {
	if err := p.match(token.DISTINCT); err != nil {
		return nil, err
	}
	var (
		dis    = &ast.Distinct{Exprs: make([]ast.Expr, 0)}
		isCall = func(expr ast.Expr) bool {
			_, ok := expr.(*ast.CallExpr)
			return ok
		}
		hasCallFn = false
	)

	field, err := p.readField()
	if err != nil {
		return nil, err
	}
	if isCall(field.Expr) {
		hasCallFn = true
	}
	dis.Exprs = append(dis.Exprs, field)

	for p.token(1).t == token.COMMA {
		if err = p.match(token.COMMA); err != nil {
			return nil, err
		}

		if hasCallFn {
			return nil, newSyntaxError(p.token(1))
		}

		field, err := p.readField()
		if err != nil {
			return nil, err
		}

		if isCall(field.Expr) {
			hasCallFn = true
		}

		dis.Exprs = append(dis.Exprs, field)
	}

	if hasCallFn && len(dis.Exprs) > 1 {
		return nil, fmt.Errorf("invalid syntax: %s", dis.String())
	}

	return dis, nil
}

func (p *Parser) from() (ast.Tables, error) {
	if err := p.match(token.FROM); err != nil {
		return nil, err
	}
	return p.tables()
}

func (p *Parser) table() (ast.Tables, error) {
	if err := p.match(token.TABLE); err != nil {
		return nil, err
	}
	return p.tables()
}

func (p *Parser) tables() (ast.Tables, error) {
	tables := ast.Tables{}
	table, err := p.readTable()
	if err != nil {
		return nil, err
	}
	tables = append(tables, table)
	for p.token(1).t == token.COMMA {
		if err = p.match(token.COMMA); err != nil {
			return nil, err
		}
		table, err := p.readTable()
		if err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (p *Parser) readTable() (*ast.Table, error) {
	tok := p.token(1)
	if tok.t == token.IDENT {
		p.read()
		return &ast.Table{Name: tok.lit}, nil
	} else if tok.t == token.QIDENT {
		p.read()
		return &ast.Table{IsQuote: true, Name: tok.lit}, nil
	}
	return nil, newSyntaxError(tok)
}

func (p *Parser) where() (ast.Expr, error) {
	if err := p.match(token.WHERE); err != nil {
		return nil, err
	}
	return p.expr()
}

func (p *Parser) expr() (root ast.Expr, err error) {
	if root, err = p.unaryExpr(); err != nil {
		return nil, err
	}
	ft := p.token(1).t

	var (
		lastExpr *ast.BinaryExpr
		rootExpr *ast.BinaryExpr
	)
	for ft.IsOperator() {
		if ft == token.IN {
			if err := p.match(ft); err != nil {
				return nil, err
			}
			//LP
			ft = p.token(1).t
			if ft == token.LPAREN {
				if err := p.match(ft); err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("( front of select ecpected")
			}

			inExpr := &ast.InExpr{}
			if lastExpr != nil {
				inExpr.Left = lastExpr.Right
				lastExpr.Right = inExpr
			} else {
				inExpr.Left = root
			}

			if right, err := p.Select(); err == nil {
				tok := p.token(1)
				if tok.t == token.RPAREN {
					if err = p.match(token.RPAREN); err != nil {
						return nil, err
					}
					inExpr.Right = right
				}
			} else if right, err := p.inArgs(); err == nil {
				tok := p.token(1)
				if tok.t == token.RPAREN {
					if err = p.match(token.RPAREN); err != nil {
						return nil, err
					}
					inExpr.Right = right
				}
			} else {
				return nil, err
			}
			if lastExpr != nil {
				root = lastExpr
			} else {
				root = inExpr
			}

		} else {
			binaryExpr := &ast.BinaryExpr{}
			binaryExpr.Op = ft

			if err := p.match(ft); err != nil {
				return nil, err
			}
			if binaryExpr.Right, err = p.unaryExpr(); err != nil {
				return nil, err
			}
			if lastExpr != nil {
				if binaryExpr.Op.Precedence() <= lastExpr.Op.Precedence() {
					//new expr is the root
					binaryExpr.Left = rootExpr
					rootExpr = binaryExpr
					lastExpr = rootExpr
				} else {
					//new expr is a sub tree of the root's right leaf
					parent := lastExpr
					child, ok := parent.Right.(*ast.BinaryExpr)
					for ok && binaryExpr.Op.Precedence() > child.Op.Precedence() {
						parent = child
						child, ok = parent.Right.(*ast.BinaryExpr)
					}
					binaryExpr.Left = parent.Right
					parent.Right = binaryExpr
				}
			} else {
				lastExpr = binaryExpr
				binaryExpr.Left = root
				rootExpr = lastExpr
			}
		}
		ft = p.token(1).t

		if rootExpr != nil {
			root = rootExpr
		}
	}
	return root, nil
}

func (p *Parser) unaryExpr() (expr ast.Expr, err error) {
	tok1 := p.token(1)
	switch tok1.t {
	case token.LPAREN:
		return p.paren()
	case token.SUB:
		if err = p.match(token.SUB); err != nil {
			return nil, err
		}
		tok2 := p.token(1)
		switch tok2.t {
		case token.INTEGER:
			return p.integer(false)
		case token.FLOAT:
			return p.float(false)
		default:
			return nil, newSyntaxError(tok2)
		}
	case token.IDENT:
		return p.ident()
	case token.QIDENT:
		return p.quoteIdent()
	case token.STRING:
		return p.string()
	case token.INTEGER:
		return p.integer(true)
	case token.FLOAT:
		return p.float(true)
	case token.TRUE, token.FALSE:
		return p.bool()
	case token.NOT:
		return p.not()
	case token.QUESTION:
		return p.parameter()
	case token.DISTINCT:
		return p.distinctField()
	case token.BINARY:
		return p.binary()
	default:
		return nil, newSyntaxError(tok1)
	}
}

func (p *Parser) paren() (expr ast.Expr, err error) {
	if err = p.match(token.LPAREN); err != nil {
		return nil, err
	}
	if expr, err = p.expr(); err == nil {
		tok := p.token(1)
		if tok.t == token.RPAREN {
			if err = p.match(token.RPAREN); err != nil {
				return nil, err
			}
			return expr, nil
		}
		return nil, newSyntaxError(tok)
	}
	return nil, err
}

func (p *Parser) ident() (ast.Expr, error) {
	tok1 := p.token(1)
	tok2 := p.token(2)
	if tok1.t == token.IDENT {
		if tok2.t == token.LPAREN {
			return p.callFn()
		}
		if err := p.match(token.IDENT); err != nil {
			return nil, err
		}
		return &ast.FieldRef{Field: tok1.lit}, nil
	}
	return nil, newSyntaxError(tok1)
}

func (p *Parser) quoteIdent() (ast.Expr, error) {
	tok := p.token(1)
	//fmt.Println("quoteIdent")
	if err := p.match(token.QIDENT); err != nil {
		return nil, err
	}
	return &ast.FieldRef{IsQuote: true, Field: tok.lit}, nil
}

func (p *Parser) string() (*ast.StringLit, error) {
	lit := p.token(1).lit
	if err := p.match(token.STRING); err != nil {
		return nil, err
	}
	return &ast.StringLit{Val: lit}, nil
}

func (p *Parser) integer(positive bool) (*ast.IntegerLit, error) {
	tok := p.token(1)
	if err := p.match(token.INTEGER); err != nil {
		return nil, err
	}
	v, err := strconv.ParseInt(tok.lit, 10, 64)
	if err != nil {
		return nil, newSyntaxError(tok)
	}
	return &ast.IntegerLit{Val: v, Positive: positive}, nil
}

func (p *Parser) float(positive bool) (*ast.FloatLit, error) {
	tok := p.token(1)
	if err := p.match(token.FLOAT); err != nil {
		return nil, err
	}
	v, err := strconv.ParseFloat(tok.lit, 64)
	if err != nil {
		return nil, newSyntaxError(tok)
	}
	return &ast.FloatLit{Val: v, Positive: positive}, nil
}

func (p *Parser) bool() (*ast.BoolLit, error) {
	tok := p.token(1)
	if err := p.match(tok.t); err != nil {
		return nil, err
	}
	return &ast.BoolLit{Val: (tok.t == token.TRUE)}, nil
}

func (p *Parser) not() (*ast.NotExpr, error) {
	if err := p.match(token.NOT); err != nil {
		return nil, err
	}
	notExpr := &ast.NotExpr{}
	expr, err := p.expr()
	if err != nil {
		return nil, err
	}
	notExpr.Right = expr
	return notExpr, nil
}

func (p *Parser) parameter() (*ast.Parameter, error) {
	if err := p.match(token.QUESTION); err != nil {
		return nil, err
	}
	return &ast.Parameter{}, nil
}

func (p *Parser) binary() (*ast.BinaryLit, error) {
	if err := p.match(token.BINARY); err != nil {
		return nil, err
	}
	expr, err := p.unaryExpr()
	if err != nil {
		return nil, err
	}
	return &ast.BinaryLit{Val: expr}, nil
}

func (p *Parser) callFn() (expr *ast.CallExpr, err error) {
	tok1 := p.token(1)
	tok2 := p.token(2)
	if tok1.t == token.IDENT && tok2.t == token.LPAREN {
		p.match(token.IDENT)
		p.match(token.LPAREN)
		expr = &ast.CallExpr{}
		expr.Func = tok1.lit
		if p.token(1).t != token.RPAREN {
			if expr.Args, err = p.callArgs(); err != nil {
				return nil, err
			}
			if err = p.match(token.RPAREN); err != nil {
				return nil, err
			}
		} else if err = p.match(token.RPAREN); err != nil {
			return nil, err
		}
		return expr, nil
	}
	return nil, newSyntaxError(tok1)
}

func (p *Parser) inArgs() (expr *ast.ArgsExpr, err error) {
	argsExpr := ast.ArgsExpr{}
	var arg ast.Expr
	if arg, err = p.expr(); err != nil {
		return nil, err
	}
	argsExpr = append(argsExpr, arg)
	tok := p.token(1)
	for tok.t == token.COMMA {
		p.match(token.COMMA)
		if arg, err = p.expr(); err != nil {
			return nil, err
		}
		argsExpr = append(argsExpr, arg)
		tok = p.token(1)
	}
	return &argsExpr, nil
}

func (p *Parser) callArgs() (expr ast.CallArgs, err error) {
	expr = ast.CallArgs{}
	var arg ast.Expr
	if arg, err = p.expr(); err != nil {
		return nil, err
	}
	expr = append(expr, arg)
	tok := p.token(1)
	for tok.t == token.COMMA {
		p.match(token.COMMA)
		if arg, err = p.expr(); err != nil {
			return nil, err
		}
		expr = append(expr, arg)
		tok = p.token(1)
	}
	return expr, nil
}

func (p *Parser) orderBy() (ast.SortFields, error) {
	if err := p.match(token.ORDER); err != nil {
		return nil, err
	}
	if err := p.match(token.BY); err != nil {
		return nil, err
	}
	fields := ast.SortFields{}
	field, err := p.sortField()
	if err != nil {
		return fields, err
	}
	fields = append(fields, field)
	for p.token(1).t == token.COMMA {
		if err := p.match(token.COMMA); err != nil {
			return nil, err
		}
		field, err := p.sortField()
		if err != nil {
			return fields, err
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func (p *Parser) sortField() (field *ast.SortField, err error) {
	tok1 := p.token(1)
	field = &ast.SortField{Name: tok1.lit}
	if tok1.t == token.IDENT {
		field.IsQuote = false
		p.read()
	} else if tok1.t == token.QIDENT {
		field.IsQuote = true
		p.read()
	} else {
		return nil, newSyntaxError(tok1)
	}
	field.ASC = true
	tok2 := p.token(1)
	switch tok2.t {
	case token.ASC:
		field.ASC = true
		err = p.match(tok2.t)
	case token.DESC:
		field.ASC = false
		err = p.match(tok2.t)
	}

	if err != nil {
		return nil, err
	}
	return field, nil
}

func (p *Parser) groupBy() (ast.GroupFields, error) {
	if err := p.match(token.GROUP); err != nil {
		return nil, err
	}
	if err := p.match(token.BY); err != nil {
		return nil, err
	}
	fields := ast.GroupFields{}
	field, err := p.groupField()
	if err != nil {
		return fields, err
	}
	fields = append(fields, field)
	for p.token(1).t == token.COMMA {
		if err := p.match(token.COMMA); err != nil {
			return nil, err
		}
		field, err := p.groupField()
		if err != nil {
			return fields, err
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func (p *Parser) groupField() (field *ast.GroupField, err error) {
	tok1 := p.token(1)
	field = &ast.GroupField{Name: tok1.lit}
	if tok1.t == token.IDENT {
		field.IsQuote = false
		p.read()
	} else if tok1.t == token.QIDENT {
		field.IsQuote = true
		p.read()
	} else {
		return nil, newSyntaxError(tok1)
	}
	return field, nil
}

func (p *Parser) limit() (int, error) {
	if err := p.match(token.LIMIT); err != nil {
		return -1, err
	}
	tok := p.token(1)
	if tok.t == token.INTEGER {
		if err := p.match(token.INTEGER); err != nil {
			return -1, err
		}
		v, err := strconv.ParseInt(tok.lit, 10, 32)
		if err != nil {
			return -1, newSyntaxError(tok)
		}
		return int(v), nil
	}
	return -1, newSyntaxError(tok)
}

func (p *Parser) offset() (int, error) {
	if err := p.match(token.OFFSET); err != nil {
		return -1, err
	}
	tok := p.token(1)
	if tok.t == token.INTEGER {
		if err := p.match(token.INTEGER); err != nil {
			return -1, err
		}
		v, err := strconv.ParseInt(tok.lit, 10, 32)
		if err != nil {
			return -1, newSyntaxError(tok)
		}
		return int(v), nil
	}
	return -1, newSyntaxError(tok)
}

func (p *Parser) read() {
	t, lit, pos := p.input.Next()
	p.tokens[p.i] = tok{t, lit, pos}
	p.i = (p.i + 1) % p.k
}

func (p *Parser) token(i int) tok {
	return p.tokens[(p.i+i-1)%p.k]
}

func (p *Parser) match(expect token.Token) error {
	found := p.token(1)
	if found.t == expect {
		p.read()
		return nil
	}
	return newIllegalTokenError(expect, found)
}

func (p *Parser) Parse() (token.Token, error) {
	found := p.token(1)
	switch found.t {
	case token.SELECT, token.DELETE, token.DROP:
		return found.t, nil
	default:
		return token.ILLEGAL, newSyntaxError(found)
	}
}
