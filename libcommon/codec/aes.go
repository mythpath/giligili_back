package codec

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
)

type AESLevel int8

const (
	AES128 AESLevel = 0
	AES192 AESLevel = 1
	AES256 AESLevel = 2
)

type AESCipher struct {
	level     AESLevel
	rawKey    string
	block     cipher.Block
	blockSize int
	realKey   []byte
}

func NewAESCipher(level AESLevel, rawKey string) *AESCipher {
	return &AESCipher{
		level:  level,
		rawKey: rawKey,
	}
}

func (ac *AESCipher) Init() (err error) {
	arg := sha256.Sum256([]byte(ac.rawKey))
	switch ac.level {
	case AES128:
		ac.realKey = arg[:16]
	case AES192:
		ac.realKey = arg[:24]
	case AES256:
		ac.realKey = arg[:32]
	}
	ac.block, err = aes.NewCipher(ac.realKey)
	ac.blockSize = ac.block.BlockSize()
	return
}

func (ac *AESCipher) Encrypt(plain_body []byte) string {
	EncryptMode := cipher.NewCBCEncrypter(ac.block, ac.realKey[:ac.blockSize])
	plain_body = PKCS5Padding(plain_body, ac.blockSize)
	EncryptMode.CryptBlocks(plain_body, plain_body)
	return base64.StdEncoding.EncodeToString(plain_body)
}

func (ac *AESCipher) Decrypt(encrypt_body string) []byte {
	DecryptMode := cipher.NewCBCDecrypter(ac.block, ac.realKey[:ac.blockSize])
	encrypt_body_binary, _ := base64.StdEncoding.DecodeString(encrypt_body)
	DecryptMode.CryptBlocks(encrypt_body_binary, encrypt_body_binary)
	return PKCS5UnPadding(encrypt_body_binary)
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	if length == 0 {
		//避免crash
		return origData[:0]
	}
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
