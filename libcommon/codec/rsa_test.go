package codec

import (
	"testing"
)

const (
	testPubKey = `
-----BEGIN PUBLIC KEY-----
MDwwDQYJKoZIhvcNAQEBBQADKwAwKAIhAImutNGpPSPkjBLsQ79FgV9rcOy0Bi9nDn1L3nB9+JTnAgMBAAE=
-----END PUBLIC KEY-----
`
)

func TestRSADecrypt(t *testing.T) {
	encryptText := "iVYu0mU3GpwMiap/cISnyRt+6k52XuvCfm9egpSqvv8="
	expected := "Hello World!"

	testRsaCipher, _, _ := NewRSASecurity(testPubKey, "")
	result, err := testRsaCipher.String(encryptText, MODE_PUBKEY_DECRYPT)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	if expected != result {
		t.Error("comparation failed")
	}

	t.Logf("decrypted data is %s", result)
}
