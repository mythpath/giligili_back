package token

import (
	"fmt"
)

// Pos specifies the line and character position of a token.
// The Char and Line are both zero-based indexes.
type Pos struct {
	Line int
	Char int
}

func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Char)
}
