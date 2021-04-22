package expr

import (
	"fmt"
	"reflect"
	"testing"
)

// func TestExpression(t *testing.T) {
// 	vars := map[string]interface{}{
// 		"v1": 5,
// 	}
// 	var tests = []struct {
// 		in  string
// 		out interface{}
// 		err string
// 	}{
// 		{in: `v1 > 1 and v1 <=10 and func1(v1) > 2`, out: true},
// 	}

// 	for _, tc := range tests {
// 		if exp, err := New().
// 			Funcs(map[string]interface{}{
// 				"func1": func1,
// 			}).
// 			Parse(tc.in); err != nil {
// 			t.Error(tc.in, ":", err)
// 		} else {

// 			if r, err := exp.Eval(vars); err != nil {
// 				t.Error(tc.in, ":", err)
// 			} else {
// 				if !reflect.DeepEqual(tc.out, r) {
// 					t.Error(tc.in, " expects ", tc.out, "but actual is ", r)
// 				}
// 			}
// 		}
// 	}
// }

func func1(arg int) (int, error) {
	return arg * arg, nil
}

func func2(arg int) (int,error) {
	return arg * arg,nil
}

func func3() (int,error) {
	return -1,nil
}

func func4() (int,error) {
	return -1,fmt.Errorf("error %s","func4")
}
func TestNestedExpression(t *testing.T) {
	vars := map[string]interface{}{
		"v1": 5,
		"line": "hello libnerv, you are so good",
	}
	var tests = []struct {
		in  string
		out interface{}
		err string
	}{
		//{in: `v1 > 1 and v1 <=10 and "func2" == "func2" and func1(func2(v1)) > 2`, out: true},
		{in: `func4() > -2`, out: true},
		// {in: `v1 > 1 and v1 <=10 and func1(func2(v1)) > -2`, out: true},
		// {in: `func1(func2(v1)) = -2`, out: false},
		// {in: `func3() = -1`, out: true},
		// {in: `-1 = func3()`, out: true},

		// regex match
		// {in: `line =~ 'hello' and line !~ 'world'`, out: true},
		// {in: `line !~ 'world' and (line =~ 'hello' or line =~ 'you')`, out: true},
	}

	for _, tc := range tests {
		if exp, err := New().
			Funcs(map[string]interface{}{
				"func1": func1,
				"func2": func2,
				"func3": func3,
				"func4": func4,
			}).
			Parse(tc.in); err != nil {
			t.Error(tc.in, ":", err)
		} else {

			if r, err := exp.Eval(vars); err != nil {
				t.Error(tc.in, ":", err)
			} else {
				if !reflect.DeepEqual(tc.out, r) {
					t.Error(tc.in, " expects ", tc.out, "but actual is ", r)
				}
			}
		}
	}
}
