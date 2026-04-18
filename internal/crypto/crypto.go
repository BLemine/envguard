package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	saltLen  = 16
	nonceLen = 12
	keyLen   = 32

	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
)

// Encrypt encrypts plaintext using AES-256-GCM with a key derived from passphrase via Argon2id.
// Output format: [16-byte salt][12-byte nonce][ciphertext+auth-tag]
func Encrypt(plaintext []byte, passphrase string) ([]byte, error) {
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("generating salt: %w", err)
	}

	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generating nonce: %w", err)
	}

	gcm, err := newGCM(passphrase, salt)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	out := make([]byte, saltLen+nonceLen+len(ciphertext))
	copy(out[:saltLen], salt)
	copy(out[saltLen:saltLen+nonceLen], nonce)
	copy(out[saltLen+nonceLen:], ciphertext)
	return out, nil
}

// Decrypt decrypts data produced by Encrypt. Returns a descriptive error on wrong passphrase.
func Decrypt(data []byte, passphrase string) ([]byte, error) {
	// minimum: salt + nonce + 16-byte GCM auth tag
	if len(data) < saltLen+nonceLen+16 {
		return nil, errors.New("file too short to be a valid encrypted .env")
	}

	salt := data[:saltLen]
	nonce := data[saltLen : saltLen+nonceLen]
	ciphertext := data[saltLen+nonceLen:]

	gcm, err := newGCM(passphrase, salt)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption failed — wrong passphrase or corrupted file")
	}
	return plaintext, nil
}

func newGCM(passphrase string, salt []byte) (cipher.AEAD, error) {
	key := argon2.IDKey([]byte(passphrase), salt, argonTime, argonMemory, argonThreads, keyLen)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}
	return gcm, nil
}
