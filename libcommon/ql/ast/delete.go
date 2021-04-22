package ast

import (
	"bytes"
)

type Delete struct {
	Tables      Tables
	Where       Expr
	HasPreciseCondition bool
}

func (p Delete) String() string {
	var buf bytes.Buffer
	_, _ = buf.WriteString("DELETE ")
	if len(p.Tables) > 0 {
		_, _ = buf.WriteString(" FROM ")
		_, _ = buf.WriteString(p.Tables.String())
	}
	if p.Where != nil {
		_, _ = buf.WriteString(" WHERE ")
		_, _ = buf.WriteString(p.Where.String())
	}
	return buf.String()
}

