package service

import (
	"crypto/sha256"
)

func GetHashStr(password string) (string, error) {
	// Хэширование пароля
	src := []byte(password)
	// создаём новый hash.Hash, вычисляющий контрольную сумму SHA-256
	h := sha256.New()
	// передаём байты для хеширования
	h.Write(src)
	// вычисляем хеш
	passwordHash := h.Sum(nil)

	return string(passwordHash), nil
}
