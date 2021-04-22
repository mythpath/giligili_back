package util

import "reflect"

func ShallowCopyStruct(src, dst interface{}) interface{} {
	lValue := reflect.ValueOf(src).Elem()
	rValue := reflect.ValueOf(dst).Elem()

	for i := 0; i < rValue.NumField(); i ++ {
		rf := rValue.Type().Field(i)
		lv := lValue.FieldByName(rf.Name)
		if ! lv.IsValid() {
			continue
		}

		if lv.Type() != rf.Type {
			continue
		}

		rValue.Field(i).Set(lv)
	}

	return dst
}
