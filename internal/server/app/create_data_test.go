package app

import (
	"context"
	"errors"
	"testing"

	"keeper/internal/mocks"
	"keeper/internal/server/config"
	"keeper/internal/server/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateData(t *testing.T) {
	mockProvider := new(mocks.Provider)
	server := &server{
		provider: mockProvider,
		cfg:      &config.Config{Secret: "thisis32byteencryptionkey1234567"},
		ctx:      context.Background(),
	}

	username := "testuser"

	t.Run("successful password creation", func(t *testing.T) {
		msg := "title::login::password"
		dataType := service.PASSWORD
		mockProvider.On("CreateData", mock.Anything, username, "title", mock.Anything, mock.Anything).Return(nil)

		_, err := server.createData(msg, username, dataType)
		assert.NoError(t, err)

		mockProvider.AssertExpectations(t)
		mockProvider.ExpectedCalls = nil
	})

	t.Run("successful text creation", func(t *testing.T) {
		msg := "title::text"
		dataType := service.TEXT
		mockProvider.On("CreateData", mock.Anything, username, "title", mock.Anything, mock.Anything).Return(nil)

		_, err := server.createData(msg, username, dataType)
		assert.NoError(t, err)

		mockProvider.AssertExpectations(t)
		mockProvider.ExpectedCalls = nil
	})

	t.Run("successful card creation", func(t *testing.T) {
		msg := "title::cardnum::expdate::owner::cvv"
		dataType := service.CARD
		mockProvider.On("CreateData", mock.Anything, username, "title", mock.Anything, mock.Anything).Return(nil)

		_, err := server.createData(msg, username, dataType)
		assert.NoError(t, err)

		mockProvider.AssertExpectations(t)
		mockProvider.ExpectedCalls = nil
	})

	t.Run("incorrect format", func(t *testing.T) {
		msg := "title::login"
		dataType := service.PASSWORD

		_, err := server.createData(msg, username, dataType)
		assert.Error(t, err)
		assert.Equal(t, ErrCreateFormat, err)
	})

	t.Run("provider error", func(t *testing.T) {
		msg := "title::login::password"
		dataType := service.PASSWORD
		mockProvider.On("CreateData", mock.Anything, username, "title", dataType, mock.Anything).Return(errors.New("provider error"))

		_, err := server.createData(msg, username, dataType)
		assert.Error(t, err)
		assert.Equal(t, "provider error", err.Error())

		mockProvider.AssertExpectations(t)
		mockProvider.ExpectedCalls = nil
	})
}
