package ast

import "strings"

type GroupFields []*GroupField

func (p GroupFields) String() string {
	var str []string
	for _, f := range p {
		str = append(str, f.String())
	}
	return strings.Join(str, ", ")
}

type GroupField struct {
	IsQuote bool
	Name    string
}

func (p GroupField) String() string {
	if p.IsQuote {
		return "`" + p.Name + "`"
	}
	return p.Name
}
