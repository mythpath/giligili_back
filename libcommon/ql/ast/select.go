package ast

import (
	"bytes"
	"fmt"
)

// Select statement
type Select struct {
	Fields      Fields
	Tables      Tables
	Where       Expr
	GroupFields GroupFields
	SortFields  SortFields
	Limit       int
	Offset      int
}

func (p Select) String() string {
	var buf bytes.Buffer
	_, _ = buf.WriteString("SELECT ")
	_, _ = buf.WriteString(p.Fields.String())
	if len(p.Tables) > 0 {
		_, _ = buf.WriteString(" FROM ")
		_, _ = buf.WriteString(p.Tables.String())
	}
	if p.Where != nil {
		_, _ = buf.WriteString(" WHERE ")
		_, _ = buf.WriteString(p.Where.String())
	}
	if p.GroupFields != nil {
		_, _ = buf.WriteString(" GROUP BY ")
		_, _ = buf.WriteString(p.GroupFields.String())
	}
	if p.SortFields != nil {
		_, _ = buf.WriteString(" ORDER BY ")
		_, _ = buf.WriteString(p.SortFields.String())
	}
	if p.Limit > 0 {
		_, _ = fmt.Fprintf(&buf, " LIMIT %d", p.Limit)
	}
	if p.Offset > 0 {
		_, _ = fmt.Fprintf(&buf, " OFFSET %d", p.Offset)
	}
	return buf.String()
}
