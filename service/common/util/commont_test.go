package util

import (
	"reflect"
	"testing"
)

func TestDeepCopy(t *testing.T) {
	testcases := []struct {
		name string
		in   func()
		out  func()
	}{
		{
			name: "test value deep copy",
			in: func() {
				type test struct {
					A int
					B string
					C bool
				}

				src := test{
					A: 233,
					B: "hello",
					C: false,
				}
				dst := test{}
				if err := DeepCopy(src, dst); err != nil {
					t.Error(err)
				}
				if reflect.DeepEqual(src, dst) {
					t.Log("相同")
				} else {
					t.Error("不相同")
				}
			},
		},
		{
			name: "test reference deep copy",
			in: func() {
				type test struct {
					A int
					B string
					C bool
				}

				src := test{
					A: 233,
					B: "hello",
					C: false,
				}
				dst := test{}
				if err := DeepCopy(src, &dst); err != nil {
					t.Error(err)
				}
				if reflect.DeepEqual(src, dst) {
					t.Log("相同")
				} else {
					t.Error("不相同")
				}
			},
		},
	}

	for _, testcase := range testcases {
		testcase.in()
	}
}
