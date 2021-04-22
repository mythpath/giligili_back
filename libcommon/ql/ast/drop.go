package ast

import (
	"bytes"
)

type Drop struct {
	Tables      Tables
	Where       Expr
	HasPreciseCondition bool
}

func (p Drop) String() string {
	var buf bytes.Buffer
	_, _ = buf.WriteString("DROP ")
	if len(p.Tables) > 0 {
		_, _ = buf.WriteString(" TABLE ")
		_, _ = buf.WriteString(p.Tables.String())
	}
	if p.Where != nil {
		_, _ = buf.WriteString(" WHERE ")
		_, _ = buf.WriteString(p.Where.String())
	}
	return buf.String()
}
