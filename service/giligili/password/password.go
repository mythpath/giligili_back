package password

import (
	"golang.org/x/crypto/bcrypt"
	"selfText/giligili_back/service/giligili/alias"
)

// SetPassword 设置密码
func SetPassword(password string, PassWordCost int) (string, error) {
	if PassWordCost==0 {
		PassWordCost= alias.PassWordCost
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), PassWordCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword 校验密码
func CheckPassword(password, passwordDigest string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordDigest), []byte(password))
	return err == nil
}
