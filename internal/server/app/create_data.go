package app

import (
	"encoding/json"
	"errors"
	"keeper/internal/logger"
	"keeper/internal/server/service"
	"strings"
)

var ErrCreateFormat = errors.New("incorrect data format")

func (s *server) createData(msg string, username string, createdType service.DataType) error {
	var partsCount int
	switch createdType {
	case service.PASSWORD:
		partsCount = 3
	case service.TEXT:
		partsCount = 2
	case service.CARD:
		partsCount = 5
	}

	// разбиваем полученные данные по разделителю
	parts := strings.Split(msg, "::")
	if len(parts) != partsCount {
		return ErrCreateFormat
	}

	createDataMap := make(map[string]string)
	var title string

	// собираем мапу с данными в зависимости от типа данных
	switch createdType {
	case service.PASSWORD:
		var login, pass string
		title, login, pass = parts[0], parts[1], parts[2]
		createDataMap["login"] = login
		createDataMap["password"] = pass
	case service.TEXT:
		var text string
		title, text = parts[0], parts[1]
		createDataMap["text"] = text
	case service.CARD:
		var cardNum, expirationDate, owner, cvv string
		title, cardNum, expirationDate, owner, cvv = parts[0], parts[1], parts[2], parts[3], parts[4]
		createDataMap["card_num"] = cardNum
		createDataMap["expiration_date"] = expirationDate
		createDataMap["owner"] = owner
		createDataMap["cvv"] = cvv
	}

	// сериализуем мапу
	createDataJson, err := json.Marshal(createDataMap)
	if err != nil {
		logger.Log.Sugar().Errorf("Error marshalling map to JSON: %v", err)
		return err
	}

	// шифруем данные
	cipherText, err := service.Encrypt(string(createDataJson), s.cfg.Secret)
	if err != nil {
		logger.Log.Sugar().Errorf("Encryption error: %v\n", err)
		return err
	}

	// сохраняем данные
	err = s.provider.CreateData(s.ctx, username, title, service.PASSWORD, cipherText)
	if err != nil {
		return err
	}
	return nil
}
