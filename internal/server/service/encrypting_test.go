package service

import (
	"testing"
)

// TestGenerateRandom проверяет функцию generateRandom
func TestGenerateRandom(t *testing.T) {
	size := 16
	randomBytes, err := generateRandom(size)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(randomBytes) != size {
		t.Errorf("Expected length %d, got %d", size, len(randomBytes))
	}
}

// TestEncryptDecrypt проверяет функции Encrypt и Decrypt
func TestEncryptDecrypt(t *testing.T) {
	plainText := "This is a secret message"
	key := "thisis32byteencryptionkey1234567" // 32 байта

	encryptedText, err := Encrypt(plainText, key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	decryptedText, err := Decrypt(encryptedText, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if decryptedText != plainText {
		t.Errorf("Expected '%s', got '%s'", plainText, decryptedText)
	}
}

// TestEncryptInvalidKey проверяет, что Encrypt возвращает ошибку при использовании некорректного ключа
func TestEncryptInvalidKey(t *testing.T) {
	plainText := "This is a secret message"
	key := "shortkey" // Ключ неправильной длины

	_, err := Encrypt(plainText, key)
	if err == nil {
		t.Fatalf("Expected error for invalid key size, got nil")
	}
}

// TestDecryptInvalidKey проверяет, что Decrypt возвращает ошибку при использовании некорректного ключа
func TestDecryptInvalidKey(t *testing.T) {
	plainText := "This is a secret message"
	key := "thisis32byteencryptionkey1234567" // 32 байта

	encryptedText, err := Encrypt(plainText, key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	invalidKey := "shortkey" // Ключ неправильной длины
	_, err = Decrypt(encryptedText, invalidKey)
	if err == nil {
		t.Fatalf("Expected error for invalid key size, got nil")
	}
}

// TestDecryptInvalidCipherText проверяет, что Decrypt возвращает ошибку при использовании некорректного шифртекста
func TestDecryptInvalidCipherText(t *testing.T) {
	key := "thisis32byteencryptionkey1234567" // 32 байта
	invalidCipherText := "invalidciphertext"

	_, err := Decrypt(invalidCipherText, key)
	if err == nil {
		t.Fatalf("Expected error for invalid ciphertext, got nil")
	}
}
