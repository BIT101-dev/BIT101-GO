package webvpn

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

func EncryptPassword(plaintext string, salt string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return "", err
	}
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", errors.New("invalid AES key length")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	input := []byte(plaintext)
	padding := aes.BlockSize - len(input)%aes.BlockSize
	input = append(input, bytes.Repeat([]byte{byte(padding)}, padding)...)
	output := make([]byte, len(input))
	for offset := 0; offset < len(input); offset += aes.BlockSize {
		block.Encrypt(output[offset:offset+aes.BlockSize], input[offset:offset+aes.BlockSize])
	}
	return base64.StdEncoding.EncodeToString(output), nil
}

func decryptSchoolPassword(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) == 0 || len(ciphertext)%block.BlockSize() != 0 {
		return nil, errors.New("invalid AES ciphertext length")
	}
	plaintext := make([]byte, len(ciphertext))
	for offset := 0; offset < len(ciphertext); offset += block.BlockSize() {
		block.Decrypt(plaintext[offset:offset+block.BlockSize()], ciphertext[offset:offset+block.BlockSize()])
	}
	padding := int(plaintext[len(plaintext)-1])
	if padding == 0 || padding > block.BlockSize() || padding > len(plaintext) {
		return nil, errors.New("invalid AES padding")
	}
	for _, value := range plaintext[len(plaintext)-padding:] {
		if int(value) != padding {
			return nil, errors.New("invalid AES padding")
		}
	}
	return plaintext[:len(plaintext)-padding], nil
}

func SealSessionCookie(plaintext string, secret string) (string, error) {
	if plaintext == "" {
		return "", errors.New("empty session cookie")
	}
	gcm, err := sessionGCM(secret)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	return base64.RawURLEncoding.EncodeToString(append(nonce, sealed...)), nil
}

func OpenSessionCookie(encoded string, secret string) (string, error) {
	gcm, err := sessionGCM(secret)
	if err != nil {
		return "", err
	}
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || len(payload) < gcm.NonceSize()+gcm.Overhead() {
		return "", errors.New("invalid session cookie")
	}
	nonce := payload[:gcm.NonceSize()]
	plaintext, err := gcm.Open(nil, nonce, payload[gcm.NonceSize():], nil)
	if err != nil || len(plaintext) == 0 {
		return "", errors.New("invalid session cookie")
	}
	return string(plaintext), nil
}

func sessionGCM(secret string) (cipher.AEAD, error) {
	if secret == "" {
		return nil, errors.New("proxy token is empty")
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
