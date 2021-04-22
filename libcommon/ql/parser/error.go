package parser

import (
	"fmt"

	"selfText/giligili_back/libcommon/ql/token"
)

// SyntaxError represents an error that occurred during parsing.
type SyntaxError struct {
	expect token.Token
	found  tok
}

// newSyntaxError returns a new instance of ParseError.
func newSyntaxError(found tok) *SyntaxError {
	return &SyntaxError{expect: token.ILLEGAL, found: found}
}

func newIllegalTokenError(expect token.Token, found tok) *SyntaxError {
	return &SyntaxError{expect: expect, found: found}
}

// Error returns the string representation of the error.
func (e *SyntaxError) Error() string {
	if e.expect == token.ILLEGAL {
		return fmt.Sprintf("syntax error:\tillegal token %s %s at line:%d char:%d", e.found.t, e.found.lit, e.found.pos.Line, e.found.pos.Char)
	}
	return fmt.Sprintf("syntax error:\texpect token %s, but found %s %s at line:%d char:%d", e.expect, e.found.t, e.found.lit, e.found.pos.Line, e.found.pos.Char)
}
