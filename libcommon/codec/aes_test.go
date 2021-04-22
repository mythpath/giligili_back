package codec

import (
	"encoding/base64"
	"fmt"
	"testing"
)

//const (
//	key           string = "Jdb@Mysql123"
//	rawdata       string = "Root@100.73"
//	encrypteddata string = "F29aHumguuoMGJQYKEhxGA=="
//)

//const (
//	key           string = "DingCloud@123"
//	rawdata       string = "ha_op"
//	encrypteddata string = "onAvra4pSuRfqNg46nauIQ=="
//)

const (
	key           string = "jdb-20161228-ha2010CJ3612"
	rawdata       string = "root@123"
	encrypteddata string = "COIkmsrmbep9g7lmT5CDkueDTKKUe5SSf6RxP06+DHw="
)

//const (
//	key           string = "DingCloud@HASecret9898"
//	rawdata       string = "jdb-20161228-ha2010CJ3612"
//	encrypteddata string = "8wocXMNiG0HozQHj/Dg7O+qzCf6YduoXgjr8LIBfYI8="
//)

func TestAESEncrypt(t *testing.T) {
	aesCipher := NewAESCipher(AES128, key)
	if err := aesCipher.Init(); err != nil {
		panic(err)
	}
	result := aesCipher.Encrypt([]byte(rawdata))
	fmt.Printf("encrypted data is %s\n", result)
}

func TestAESDecrypt(t *testing.T) {
	aesCipher := NewAESCipher(AES128, key)
	if err := aesCipher.Init(); err != nil {
		t.Logf("init panic. err:%v", err)
		panic(err)
	}

	defer func() {
		if err := recover(); err != nil {
			t.Logf("panic. %v", err)
			a, _ := base64.StdEncoding.DecodeString(encrypteddata)
			b, _ := base64.StdEncoding.DecodeString(string(a))
			t.Logf("recover decrypted data is %s", b)
		}
	}()

	result := aesCipher.Decrypt(encrypteddata)
	t.Logf("decrypted data is %s", result)
}
