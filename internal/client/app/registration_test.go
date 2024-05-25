package app

import (
	"bufio"
	"context"
	"errors"
	"keeper/internal/mocks"
	pb "keeper/proto"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCredentials(t *testing.T) {
	t.Run("successful input", func(t *testing.T) {
		input := "username password\n"
		reader := bufio.NewReader(strings.NewReader(input))
		username, password, err := getCredentials(*reader)
		assert.NoError(t, err)
		assert.Equal(t, "username", username)
		assert.Equal(t, "password", password)
	})

	t.Run("invalid input format", func(t *testing.T) {
		input := "invalid_input_format\n"
		reader := bufio.NewReader(strings.NewReader(input))
		username, password, err := getCredentials(*reader)
		assert.Error(t, err)
		assert.Equal(t, ErrCredentialsFormat, err)
		assert.Equal(t, "", username)
		assert.Equal(t, "", password)
	})

	t.Run("error reading input", func(t *testing.T) {
		reader := bufio.NewReader(strings.NewReader(""))
		_, _, err := getCredentials(*reader)
		assert.Error(t, err)
	})
}

func TestRegistration(t *testing.T) {
	mockClient := new(mocks.KeeperServiceClient)
	mockStream := new(mocks.KeeperService_CommandClient)
	app := &App{
		ctx: context.Background(),
		wg:  &sync.WaitGroup{},
	}

	t.Run("registration failed", func(t *testing.T) {
		input := "username password\n"
		reader := bufio.NewReader(strings.NewReader(input))
		mockClient.On("Register", mock.Anything, &pb.RegisterRequest{Username: "username", Password: "password"}).
			Return(nil, errors.New("registration failed"))

		err := app.registration(*reader, mockClient, mockStream)
		assert.Error(t, err)
		assert.EqualError(t, err, "registration failed")

		mockClient.AssertExpectations(t)
		mockStream.AssertExpectations(t)
	})

	t.Run("invalid input format", func(t *testing.T) {
		input := "invalid_input_format\n"
		reader := bufio.NewReader(strings.NewReader(input))

		err := app.registration(*reader, mockClient, mockStream)
		assert.Error(t, err)
		assert.Equal(t, ErrCredentialsFormat, err)

		mockClient.AssertExpectations(t)
		mockStream.AssertExpectations(t)
	})

	t.Run("error reading input", func(t *testing.T) {
		reader := bufio.NewReader(strings.NewReader(""))

		err := app.registration(*reader, mockClient, mockStream)
		assert.Error(t, err)

		mockClient.AssertExpectations(t)
		mockStream.AssertExpectations(t)
	})
}
