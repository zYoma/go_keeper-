package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// generateRandom генерирует криптостойкие случайные байты заданного размера
func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Encrypt шифрует данные с использованием ключа шифрования
func Encrypt(plainText string, key string) (string, error) {
	// Преобразование строки в байтовый срез
	plainTextBytes := []byte(plainText)
	encryptionKey := []byte(key)

	aesblock, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	nonce, err := generateRandom(aesgcm.NonceSize())
	if err != nil {
		return "", err
	}

	cipherText := aesgcm.Seal(nonce, nonce, plainTextBytes, nil)
	return hex.EncodeToString(cipherText), nil
}

// Decrypt расшифровывает данные с использованием ключа шифрования
func Decrypt(cipherTextHex string, key string) (string, error) {
	encryptionKey := []byte(key)

	cipherText, err := hex.DecodeString(cipherTextHex)
	if err != nil {
		return "", err
	}

	aesblock, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	nonceSize := aesgcm.NonceSize()
	if len(cipherText) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plainText, err := aesgcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
