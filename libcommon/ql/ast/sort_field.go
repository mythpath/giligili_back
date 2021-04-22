package ast

import "strings"

type SortFields []*SortField

func (p SortFields) String() string {
	var str []string
	for _, f := range p {
		str = append(str, f.String())
	}
	return strings.Join(str, ", ")
}

type SortField struct {
	IsQuote bool
	Name    string
	ASC     bool
}

func (p SortField) String() string {
	name := p.Name
	if p.IsQuote {
		name = "`" + name + "`"
	}
	if p.ASC {
		return name + " ASC"
	}
	return name + " DESC"
}
