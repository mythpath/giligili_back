package expr

import (
	"strings"

	"selfText/giligili_back/libcommon/ql/ast"
	"selfText/giligili_back/libcommon/ql/lexer"
	"selfText/giligili_back/libcommon/ql/parser"
)

// Expression xxxx
type Expression struct {
	qlExpr ast.Expr
	funcs  funcs
}

// New create an expression
func New() *Expression {
	return &Expression{}
}

// Parse an expression
func (p *Expression) Parse(s string) (*Expression, error) {
	qlExpr, err := parser.NewParser(lexer.NewLexer(strings.NewReader(s))).Expression()
	if err != nil {
		return nil, err
	}
	return &Expression{qlExpr: qlExpr, funcs: p.funcs}, nil
}

// Funcs register functions
func (p *Expression) Funcs(funcs map[string]interface{}) *Expression {
	p.funcs = newFuncs(funcs)
	return p
}

// Eval evaluate the value of the expression
func (p *Expression) Eval(data interface{}) (interface{}, error) {
	ctx := &Context{data: data, funcs: p.funcs}
	return evalExpr(p.qlExpr, ctx)
}
