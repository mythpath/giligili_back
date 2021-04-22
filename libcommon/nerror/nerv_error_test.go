package nerror_test

import (
	"selfText/giligili_back/libcommon/nerror"
	"log"
	"runtime/debug"
	"testing"
)

func foo1() error {
	nerr := nerror.NewCommonError("hello world")
	return nerr
}

func foo2() error {
	nerr := nerror.NewCommonError("test error")
	return nerr
}

func TestNervError(t *testing.T) {
	err := foo1()
	if err != nil {
		log.Printf("nerv error result: %s", err.Error())
	} else {
		log.Printf("no error")
	}

	err = foo2()
	if err != nil {
		log.Printf("nerv error result: %s", err.Error())
	} else {
		log.Printf("no error")
	}
}

func TestNestedNervError(t *testing.T) {
	err := foo1()
	nerr1 := nerror.NewCommonError("test error %v", err)
	log.Printf("nerr1: %s", nerr1.Error())
	log.Printf("mask: %s", nerr1.Masking())

	nerr2 := nerror.NewCommonError("test error %s", err.Error())
	log.Printf("nerr2: %s", nerr2.Error())
	log.Printf("mask: %s", nerr2.Masking())

	nerr3 := nerror.NewCommonError(err.Error())
	log.Printf("nerr3: %s", nerr3.Error())
	log.Printf("mask: %s", nerr3.Masking())
}

func foo3_1(r interface{}) {
	nerr := nerror.NewPanicError(r)
	log.Printf("catch panic: %s", nerr.Error())
	log.Println("masking", nerr.Masking())
}

func foo3() {
	defer func() {
		if r := recover(); r != nil {
			foo3_1(r)
			debug.PrintStack()
		}
	}()

	errMsg := map[string]string{
		"test1": "111",
		"test2": "222",
	}

	panic(errMsg)
}

func TestNervPanicError(t *testing.T) {
	foo3()
	log.Printf("after panic function")
}

func foo4() error {
	nerr := nerror.NewArgumentError("MyStruct")
	nerr = nerr.FieldError("id", "test error")
	log.Printf("valid error result: %s", nerr.Error())
	nerr.FieldError("name", "2 error")
	log.Printf("valid error result: %s", nerr.Error())
	commErr := nerror.NewCommonError("common error")
	nerr.FieldError("common", commErr.Error())
	return nerr
}

func TestArgumentError(t *testing.T) {
	err := foo4()
	if errMask, ok := err.(nerror.NervError); ok {
		log.Printf("is of type NervError")
		log.Printf("masking error: %s", errMask.Masking())
	} else {
		log.Printf("cannot convert to NervError")
	}
}
