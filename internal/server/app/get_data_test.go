package app

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"keeper/internal/mocks"
	"keeper/internal/server/config"
	"keeper/internal/server/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetData(t *testing.T) {
	mockProvider := new(mocks.Provider)
	server := &server{
		provider: mockProvider,
		cfg:      &config.Config{Secret: "thisis32byteencryptionkey1234567"},
		ctx:      context.Background(),
	}

	username := "testuser"
	title := "testtitle"

	t.Run("successful data retrieval", func(t *testing.T) {
		// Mocking GetData
		dataMap := map[string]string{
			"login":    "testlogin",
			"password": "testpassword",
		}
		dataMapJSON, _ := json.Marshal(dataMap)
		encryptedData, _ := service.Encrypt(string(dataMapJSON), server.cfg.Secret)
		mockProvider.On("GetData", mock.Anything, username, title).Return(encryptedData, nil)

		message, err := server.getData(username, title)
		assert.NoError(t, err)
		assert.NotNil(t, message)
		assert.Contains(t, message, "Ваши данные:\n")
		assert.Contains(t, message, "login: testlogin\n")
		assert.Contains(t, message, "password: testpassword\n")

		mockProvider.AssertExpectations(t)
		mockProvider.ExpectedCalls = nil
	})

	t.Run("data decryption error", func(t *testing.T) {
		// Mocking GetData
		mockProvider.On("GetData", mock.Anything, username, title).Return("invalid encrypted data", nil)

		message, err := server.getData(username, title)
		assert.Error(t, err)
		assert.Equal(t, "", message)

		mockProvider.AssertExpectations(t)
		mockProvider.ExpectedCalls = nil
	})

	t.Run("data unmarshalling error", func(t *testing.T) {
		// Mocking GetData
		encryptedData, _ := service.Encrypt("invalid json", server.cfg.Secret)
		mockProvider.On("GetData", mock.Anything, username, title).Return(encryptedData, nil)

		message, err := server.getData(username, title)
		assert.Error(t, err)
		assert.Equal(t, "", message)

		mockProvider.AssertExpectations(t)
		mockProvider.ExpectedCalls = nil
	})

	t.Run("provider error", func(t *testing.T) {
		// Mocking GetData
		mockProvider.On("GetData", mock.Anything, username, title).Return("", fmt.Errorf("provider error"))

		message, err := server.getData(username, title)
		assert.Error(t, err)
		assert.Equal(t, "", message)

		mockProvider.AssertExpectations(t)
		mockProvider.ExpectedCalls = nil
	})
}
