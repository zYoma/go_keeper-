package app

import (
	"encoding/json"
	"fmt"
	"keeper/internal/logger"
	"keeper/internal/server/service"
	"strings"
)

func (s *server) getData(username string, title string) (string, error) {

	jsonString, err := s.provider.GetData(s.ctx, username, title)
	if err != nil {
		return "", err
	}

	decryptedJson, err := service.Decrypt(jsonString, s.cfg.Secret)
	if err != nil {
		logger.Log.Sugar().Errorf("Decryption error: %v\n", err)
		return "", err
	}

	dataMap := make(map[string]string)

	// Преобразование JSON-строки в карту
	err = json.Unmarshal([]byte(decryptedJson), &dataMap)
	if err != nil {
		logger.Log.Sugar().Errorf("Error unmarshalling JSON: %v", err)
		return "", err
	}

	var builder strings.Builder
	builder.WriteString("Ваши данные:\n")

	for key, value := range dataMap {
		builder.WriteString(fmt.Sprintf("%s: %s\n", key, value))
	}

	return builder.String(), nil
}
