package pubsub

import (
	"crypto/md5"
	"encoding/hex"
)

func Uniqueidentifier(str ...string) string {
	h := md5.New()
	input := ""
	for _, s := range str {
		input = input + s
	}
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}
