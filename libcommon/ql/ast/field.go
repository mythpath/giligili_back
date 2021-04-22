package ast

import (
	"fmt"
	"strings"
)

// Fields
type Fields []*Field

func (p Fields) String() string {
	var str []string
	for _, f := range p {
		str = append(str, f.String())
	}
	return strings.Join(str, ", ")
}

// Field
type Field struct {
	// IsQuote bool
	Expr    Expr
	AS      string
}

func (p Field) String() string {
	str := p.Expr.String()

	if p.AS == "" {
		return str
	}
	// if p.IsQuote {
	// 	return fmt.Sprintf("`%s` AS %s", str, p.AS)
	// }
	return fmt.Sprintf("%s AS %s", str, p.AS)
}

func (p Field) AsName() string {
	if len(p.AS) > 0 {
		return p.AS
	}
	return p.Expr.String()
}

// Distinct
type Distinct struct {
	Exprs 	[]Expr
}

func (p *Distinct) String() string {
	var args []string
	for _, expr := range p.Exprs {
		args = append(args, expr.String())
	}

	return fmt.Sprintf("DISTINCT %s", strings.Join(args, ", "))
}
