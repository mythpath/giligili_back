package ast

import (
	"strings"
)

type Tables []*Table

func (p Tables) String() string {
	var str []string
	for _, t := range p {
		str = append(str, t.String())
	}
	return strings.Join(str, ", ")
}

//Table
type Table struct {
	IsQuote bool
	Name    string
}

func (p Table) String() string {
	if p.IsQuote {
		return "`" + p.Name + "`"
	}
	return p.Name
}
