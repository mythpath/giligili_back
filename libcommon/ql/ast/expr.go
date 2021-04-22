package ast

import (
	"fmt"
	"strconv"
	"strings"

	"time"

	"selfText/giligili_back/libcommon/ql/token"
)

// Expr
type Expr interface {
	String() string
}

type InExpr struct {
	Left  Expr
	Right Expr
}

func (p InExpr) String() string {
	return fmt.Sprintf("%s IN (%s)", p.Left.String(), p.Right.String())
}

type ArgsExpr []Expr

func (p ArgsExpr) String() string {
	var str []string
	for _, arg := range p {
		str = append(str, arg.String())
	}
	return strings.Join(str, ",")
}


type BinaryExpr struct {
	Op    token.Token
	Left  Expr
	Right Expr
}

func (p BinaryExpr) String() string {
	left, right := "", ""
	if p.Left != nil {
		left = p.Left.String()
	}
	if p.Right != nil {
		right = p.Right.String()
	}
	return fmt.Sprintf("(%s %s %s)", left, p.Op.String(), right)
}

type NotExpr struct {
	Right Expr
}

func (p NotExpr) String() string {
	return fmt.Sprintf("(NOT %s)", p.Right.String())
}

type AllField struct {
}

func (p AllField) String() string {
	return "*"
}

type FieldRef struct {
	IsQuote bool
	Field   string
}

func (p FieldRef) String() string {
	if p.IsQuote {
		return "`" + p.Field + "`"
	}
	return p.Field
}

type CallExpr struct {
	Func string
	Args CallArgs
	AS   string
}

func (p CallExpr) String() string {
	if p.Args == nil {
		return fmt.Sprintf("%s()", p.Func)
	}
	return fmt.Sprintf("%s(%s)", p.Func, p.Args.String())
}

func (p CallExpr) AsName() string {
	if len(p.AS) > 0 {
		return p.AS
	}
	return p.String()
}

type CallArgs []Expr

func (p CallArgs) String() string {
	var str []string
	for _, arg := range p {
		str = append(str, arg.String())
	}
	return strings.Join(str, ", ")
}

type StringLit struct {
	Val string
}

func (p StringLit) String() string {
	return p.Val
}

type IntegerLit struct {
	Positive bool
	Val      int64
}

func (p IntegerLit) String() string {
	if p.Positive {
		return fmt.Sprintf("%d", p.Val)
	} else {
		return fmt.Sprintf("-%d", p.Val)
	}
}

type FloatLit struct {
	Positive bool
	Val      float64
}

func (p FloatLit) String() string {
	s := strconv.FormatFloat(p.Val, 'f', -1, 64)
	if p.Positive == false {
		s = "-" + s
	}
	return s
}

type BoolLit struct {
	Val bool
}

func (p BoolLit) String() string {
	if p.Val {
		return "true"
	}
	return "false"
}

// Parameter is a ?
type Parameter struct {
}

func (p Parameter) String() string {
	return "?"
}

type BinaryLit struct {
	Val Expr
}

func (b BinaryLit) String() string {
	return fmt.Sprintf("BINARY %s", b.Val.String())
}

type TimeLit struct {
	Val time.Time
}

func (t *TimeLit) String() string {
	return fmt.Sprintf("%v", t.Val)
}

type NilLit struct {
}

func (t *NilLit) String() string {
	return `nil`
}

type DurationLit struct {
	Val time.Duration
}

func (t *DurationLit) String() string {
	return FormatDuration(t.Val)
}

// FormatDuration formats a duration to a string.
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	} else if d%(7*24*time.Hour) == 0 {
		return fmt.Sprintf("%dw", d/(7*24*time.Hour))
	} else if d%(24*time.Hour) == 0 {
		return fmt.Sprintf("%dd", d/(24*time.Hour))
	} else if d%time.Hour == 0 {
		return fmt.Sprintf("%dh", d/time.Hour)
	} else if d%time.Minute == 0 {
		return fmt.Sprintf("%dm", d/time.Minute)
	} else if d%time.Second == 0 {
		return fmt.Sprintf("%ds", d/time.Second)
	} else if d%time.Millisecond == 0 {
		return fmt.Sprintf("%dms", d/time.Millisecond)
	}

	return fmt.Sprintf("%du", d/time.Microsecond)
}
