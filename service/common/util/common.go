package util

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"time"
)

// RandStringRunes 返回随机字符串
func RandStringRunes(n int) string {
	var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// ConvertUint2String Uint转字符串
func ConvertUint2String(target []uint) []string {
	if len(target) <= 0 {
		return []string{}
	}
	out := make([]string, len(target))
	for _, it := range target {
		out = append(out, strconv.Itoa(int(it)))
	}
	return out
}

// ConvertInt2String int转字符串
func ConvertInt2String(target []int) []string {
	if len(target) <= 0 {
		return []string{}
	}
	out := make([]string, len(target))
	for _, it := range target {
		out = append(out, strconv.Itoa(it))
	}
	return out
}

// DeepCopy 结构体深度拷贝
func DeepCopy(src, dst interface{}) error {
	if src == nil {
		return fmt.Errorf("src struct should not be nil")
	}

	if !isStruct(src) {
		return fmt.Errorf("src should be struct")
	}
	if !isPointer(dst) {
		return fmt.Errorf("dst should be ptr")
	}

	if data, err := json.Marshal(src); err == nil {
		if revErr := json.Unmarshal(data, dst); revErr != nil {
			return revErr
		}
	} else {
		return err
	}

	return nil
}

func isStruct(i interface{}) bool {
	return reflect.ValueOf(i).Type().Kind() == reflect.Struct
}

func isPointer(i interface{}) bool {
	return reflect.ValueOf(i).Type().Kind() == reflect.Ptr
}
