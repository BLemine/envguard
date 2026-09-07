package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptionRoundTripAndAuthentication(t *testing.T) {
	for _, plaintext := range [][]byte{nil, []byte("API_KEY=fixture\nMULTILINE=\"a\nb\"\n"), {0, 255, 128}} {
		encrypted, err := Encrypt(plaintext, "test passphrase")
		if err != nil {
			t.Fatal(err)
		}
		again, err := Encrypt(plaintext, "test passphrase")
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Equal(encrypted, again) {
			t.Fatal("encryption reused salt and nonce")
		}
		decrypted, err := Decrypt(encrypted, "test passphrase")
		if err != nil || !bytes.Equal(decrypted, plaintext) {
			t.Fatalf("roundtrip: %v", err)
		}
		if _, err := Decrypt(encrypted, "wrong"); err == nil {
			t.Fatal("wrong passphrase accepted")
		}
		for _, offset := range []int{0, saltLen, len(encrypted) - 1} {
			corrupted := append([]byte(nil), encrypted...)
			corrupted[offset] ^= 1
			if _, err := Decrypt(corrupted, "test passphrase"); err == nil {
				t.Fatalf("tampering at %d accepted", offset)
			}
		}
	}
	for _, size := range []int{0, 1, 27, 43} {
		if _, err := Decrypt(make([]byte, size), "test"); err == nil {
			t.Fatalf("truncated file of length %d accepted", size)
		}
	}
}
