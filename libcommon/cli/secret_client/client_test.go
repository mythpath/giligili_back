package secret_client

import (
	"reflect"
	"testing"
)

func TestSecretClient(t *testing.T) {
	sc := New("", []string{"http://localhost:5000"})
	oriData := map[string]interface{}{"foo": "bar", "aa": "2"}
	ret, err := sc.Encrypt(oriData)
	if err != nil {
		t.Error("Encrypt err: ", err)
		return
	}
	expData, err := sc.Decrypt(ret)
	if err != nil {
		t.Error("Decrypt err: ", err)
		return
	}
	if !reflect.DeepEqual(expData, oriData) {
		t.Error("Decrypt result not as expected")
	}
}
