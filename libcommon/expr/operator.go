package expr

import (
	"fmt"
	"strconv"
	"strings"

	"selfText/giligili_back/libcommon/ql/ast"
	"selfText/giligili_back/libcommon/ql/token"
	"regexp"
)

func evalExpr(expr ast.Expr, ctx *Context) (v interface{}, err error) {
	switch t := expr.(type) {
	case *ast.BinaryExpr:
		v, err = evalBinaryExpr(t, ctx)
	case *ast.FieldRef:
		v, err = evalVarExpr(t, ctx)
	case *ast.StringLit:
		v, err = t.Val[1:len(t.Val)-1], nil
	case *ast.BoolLit:
		v, err = t.Val, nil
	case *ast.IntegerLit:
		if t.Positive {
			v, err = t.Val, nil
		}else{
			v, err = -t.Val, nil
		}
	case *ast.FloatLit:
		if t.Positive {
			v, err = t.Val, nil
		}else{
			v, err = -t.Val, nil
		}
	case *ast.CallExpr:
		v, err = evalCallExpr(t, ctx)
	case *ast.NotExpr:
		v, err = evalNotExpr(t, ctx)
	default:
		return nil, fmt.Errorf("evalExpr isn't supported expr: %T %s", t, t.String())
	}
	if err != nil {
		err = fmt.Errorf("evalExpr run %s failed. err: %s", expr.String(), err.Error())
	}
	return
}

func evalVarExpr(expr *ast.FieldRef, ctx *Context) (v interface{}, err error) {
	return ctx.value(expr.Field)
}

func evalCallExpr(expr *ast.CallExpr, ctx *Context) (v interface{}, err error) {
	args := expr.Args
	vs := []interface{}{}
	if args != nil {
		for _, arg := range args {
			v, err = evalExpr(arg, ctx)
			if err != nil {
				return
			}
			vs = append(vs, v)
		}
	}
	v, err = ctx.call(expr.Func, vs...)
	return
}

func evalNotExpr(expr *ast.NotExpr, ctx *Context) (v interface{}, err error) {
	var r interface{}
	r, err = evalExpr(expr.Right, ctx)
	if err == nil {
		switch rv := r.(type) {
		case bool:
			v = !rv
		default:
			err = fmt.Errorf("right expr isn't boolean. err: %s ", expr.String())
		}
	}
	return
}

func evalBinaryExpr(expr *ast.BinaryExpr, ctx *Context) (v interface{}, err error) {
	var l, r interface{}
	l, err = evalExpr(expr.Left, ctx)
	if err != nil {
		return nil, err
	}
	r, err = evalExpr(expr.Right, ctx)
	if err != nil {
		return nil, err
	}
	return evalOp(expr.Op, l, r, ctx)
}

type opfunc func(lv, rv interface{}, ctx *Context) (r interface{}, err error)

var ops = [...]opfunc{
	token.ADD: evalAdd,
	token.SUB: evalSub,
	token.MUL: evalMul,
	token.DIV: evalDiv,
	token.MOD: evalMod,
	token.AND: evalAnd,
	token.OR:  evalOr,
	token.EQ:  evalEq,
	token.NEQ: evalNeq,
	token.LT:  evalLt,
	token.LTE: evalLte,
	token.GT:  evalGt,
	token.GTE: evalGte,
	token.EQREGEX: evalEqRegex,
	token.NEQREGEX: evalNeqRegex,
}

func evalAdd(lv, rv interface{}, ctx *Context) (r interface{}, err error) {
	switch lt := lv.(type) {
	case int64:
		switch rt := rv.(type) {
		case int64:
			return lt + rt, nil
		case float64:
			return float64(lt) + rt, nil
		default:
			return nil, fmt.Errorf("add failed. rv's type is %v", rt)
		}
	case float64:
		switch rt := rv.(type) {
		case int64:
			return lt + float64(rt), nil
		case float64:
			return lt + rt, nil
		default:
			return nil, fmt.Errorf("add failed. rv's type is %v", rt)
		}
	default:
		return nil, fmt.Errorf("add failed. lv's type is %v", lt)
	}
}

func evalOp(op token.Token, lv, rv interface{}, ctx *Context) (r interface{}, err error) {
	f := ops[op]
	if f == nil {
		err = fmt.Errorf("unsupport operator %s", op.String())
		return
	}
	return f(lv, rv, ctx)
}

func evalSub(lv, rv interface{}, ctx *Context) (r interface{}, err error) {
	switch lt := lv.(type) {
	case int64:
		switch rt := rv.(type) {
		case int64:
			return lt - rt, nil
		case float64:
			return float64(lt) - rt, nil
		default:
			return nil, fmt.Errorf("sub failed. rv's type is %v", rt)
		}
	case float64:
		switch rt := rv.(type) {
		case int64:
			return lt - float64(rt), nil
		case float64:
			return lt - rt, nil
		default:
			return nil, fmt.Errorf("sub failed. rv's type is %v", rt)
		}
	default:
		return nil, fmt.Errorf("sub failed. lv's type is %v", lt)
	}
}

func evalMul(lv, rv interface{}, ctx *Context) (r interface{}, err error) {
	switch lt := lv.(type) {
	case int64:
		switch rt := rv.(type) {
		case int64:
			return lt * rt, nil
		case float64:
			return float64(lt) * rt, nil
		default:
			return nil, fmt.Errorf("mul failed. rv's type is %v", rt)
		}
	case float64:
		switch rt := rv.(type) {
		case int64:
			return lt * float64(rt), nil
		case float64:
			return lt * rt, nil
		default:
			return nil, fmt.Errorf("mul failed. rv's type is %v", rt)
		}
	default:
		return nil, fmt.Errorf("mul failed. lv's type is %v", lt)
	}
}

func evalDiv(lv, rv interface{}, ctx *Context) (r interface{}, err error) {
	switch lt := lv.(type) {
	case int64:
		switch rt := rv.(type) {
		case int64:
			return lt / rt, nil
		case float64:
			return float64(lt) / rt, nil
		default:
			return nil, fmt.Errorf("div failed. rv's type is %v", rt)
		}
	case float64:
		switch rt := rv.(type) {
		case int64:
			return lt / float64(rt), nil
		case float64:
			return lt / rt, nil
		default:
			return nil, fmt.Errorf("div failed. rv's type is %v", rt)
		}
	default:
		return nil, fmt.Errorf("div failed. lv's type is %v", lt)
	}
}

func evalMod(lv, rv interface{}, ctx *Context) (r interface{}, err error) {
	switch lt := lv.(type) {
	case int64:
		switch rt := rv.(type) {
		case int64:
			if rt == 0 {
				return nil, fmt.Errorf("mod failed. rv is zero. rv's type is %T", rt)
			}
			return lt % rt, nil
		default:
			return nil, fmt.Errorf("mod failed. rv's type is %v", rt)
		}
	default:
		return nil, fmt.Errorf("mod failed. lv's type is %v", lt)
	}
}

func evalAnd(lv, rv interface{}, ctx *Context) (r interface{}, err error) {
	switch lt := lv.(type) {
	case bool:
		switch rt := rv.(type) {
		case bool:

			return lt && rt, nil
		default:
			return nil, fmt.Errorf("and failed. rv's type is %v", rt)
		}
	default:
		return nil, fmt.Errorf("and failed. lv's type is %v", lt)
	}
}

func evalOr(lv, rv interface{}, ctx *Context) (r interface{}, err error) {
	switch lt := lv.(type) {
	case bool:
		switch rt := rv.(type) {
		case bool:
			return lt || rt, nil
		default:
			return nil, fmt.Errorf("or failed. rv's type is %v", rt)
		}
	default:
		return nil, fmt.Errorf("or failed. lv's type is %v", lt)
	}
}

func evalEq(lv, rv interface{}, ctx *Context) (r interface{}, err error) {
	var v int64
	v, err = evalCompare(lv, rv, ctx)
	r = v == 0
	return
}

func evalNeq(lv, rv interface{}, ctx *Context) (r interface{}, err error) {
	var v int64
	v, err = evalCompare(lv, rv, ctx)
	r = v != 0
	return
}

func evalCompare(lv, rv interface{}, ctx *Context) (r int64, err error) {
	//fmt.Printf("l:%v r:%v\n", lv, rv)
	switch lt := lv.(type) {
	case int:
		switch rt := rv.(type) {
		case int:
			return int64(lt - rt), nil
		case int64:
			return int64(lt) - rt, nil
		case float64:
			return int64(float64(lt) - rt), nil
		case string:
			if v, err := strconv.ParseInt(rt, 10, 64); err != nil {
				return 0, err
			} else {
				return int64(lt) - v, nil
			}
		default:
			return 0, fmt.Errorf("compare failed. rv's type is %T", rt)
		}
	case int64:
		switch rt := rv.(type) {
		case int:
			return lt - int64(rt), nil
		case int64:
			return lt - rt, nil
		case float64:
			return int64(float64(lt) - rt), nil
		case string:
			if v, err := strconv.ParseInt(rt, 10, 64); err != nil {
				return 0, err
			} else {
				return lt - v, nil
			}
		default:
			return 0, fmt.Errorf("compare failed. rv's type is %T", rt)
		}
	case float64:
		switch rt := rv.(type) {
		case int:
			return int64(lt - float64(rt)), nil
		case int64:
			return int64(lt - float64(rt)), nil
		case float64:
			return int64(lt - rt), nil
		case string:
			if v, err := strconv.ParseFloat(rt, 64); err != nil {
				return 0, err
			} else {
				return int64(lt - v), nil
			}
		default:
			return 0, fmt.Errorf("compare failed. rv's type is %T", rt)
		}
	case string:
		switch rt := rv.(type) {
		case int:
			if v, err := strconv.ParseInt(lt, 10, 64); err != nil {
				return 0, err
			} else {
				return v - int64(rt), nil
			}
		case int64:
			if v, err := strconv.ParseInt(lt, 10, 64); err != nil {
				return 0, err
			} else {
				return v - rt, nil
			}
		case float64:
			if v, err := strconv.ParseFloat(lt, 64); err != nil {
				return 0, err
			} else {
				return int64(v - rt), nil
			}
		case string:
			return int64(strings.Compare(lt, rt)), nil
		default:
			return 0, fmt.Errorf("compare failed. rv's type is %T", rt)
		}
	case bool:
		switch rt := rv.(type) {
		case bool:
			if lt == rt {
				return 0, nil
			} else if lt == true {
				return 1, nil
			} else {
				return -1, nil
			}
		default:
			return 0, fmt.Errorf("compare failed. rv's type is %T", rv)
		}
	default:
		return 0, fmt.Errorf("compare failed. lv's type is %T", lv)
	}
}

func evalLt(lv, rv interface{}, ctx *Context) (r interface{}, err error) {
	var v int64
	v, err = evalCompare(lv, rv, ctx)
	r = v < 0
	return
}

func evalLte(lv, rv interface{}, ctx *Context) (r interface{}, err error) {
	var v int64
	v, err = evalCompare(lv, rv, ctx)
	r = v <= 0
	return
}

func evalGt(lv, rv interface{}, ctx *Context) (r interface{}, err error) {
	var v int64
	v, err = evalCompare(lv, rv, ctx)
	r = v > 0
	return
}

func evalGte(lv, rv interface{}, ctx *Context) (r interface{}, err error) {
	var v int64
	v, err = evalCompare(lv, rv, ctx)
	r = v >= 0
	return
}

func evalEqRegex(lv, rv interface{}, ctx *Context) (r interface{}, err error) {
	return evalMatch(lv, rv, token.EQREGEX)
}

func evalNeqRegex(lv, rv interface{}, ctx *Context) (r interface{}, err error) {
	return evalMatch(lv, rv, token.NEQREGEX)
}

// rv is the default keyword
func evalMatch(lv, rv interface{}, op token.Token) (r interface{}, err error) {
	//fmt.Printf("l:%v, r:%v\n", lv, rv)
	var re *regexp.Regexp

	switch rt := rv.(type) {
	case string:
		re = regexp.MustCompile(rt)
		switch lt := lv.(type) {
		case string:
			// start match
			switch op {
			case token.EQREGEX:
				return re.MatchString(lt), nil
			case token.NEQREGEX:
				return !re.MatchString(lt), nil
			default:
				return false, fmt.Errorf("match failed, unsupport operator: %v", op.String())
			}
		default:
			return false, fmt.Errorf("match failed, lv's type is %T", lv)
		}
	default:
		return false, fmt.Errorf("match failed, rv's type is %T", rv)
	}

	return nil, nil
}
