package codec

import "fmt"

func Decrypt(encrypt, aesKey string) (decrypt string, err error) {
	aesCipher := NewAESCipher(AES128, aesKey)
	if err = aesCipher.Init(); err != nil {
		err = fmt.Errorf("failed to init aes cipher: %v", err)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("decrypt panic: %v", r)
		}
	}()

	decrypt = string(aesCipher.Decrypt(encrypt))
	return
}

func Encrypt(encrypt, aesKey string) (decrypt string, err error) {
	aesCipher := NewAESCipher(AES128, aesKey)
	if err = aesCipher.Init(); err != nil {
		err = fmt.Errorf("failed to init aes cipher")
	}

	decrypt = aesCipher.Encrypt([]byte(encrypt))

	return
}
