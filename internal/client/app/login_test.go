package app

import (
	"bufio"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"keeper/internal/client/config"
	"keeper/internal/mocks"
	pb "keeper/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogIn(t *testing.T) {
	mockClient := new(mocks.KeeperServiceClient)
	mockStream := new(mocks.KeeperService_CommandClient)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app := &App{
		ctx: ctx,
		cfg: &config.Config{
			ServerAddr: "localhost:50051",
		},
		wg: &sync.WaitGroup{},
	}

	t.Run("invalid input format", func(t *testing.T) {
		input := "invalid_input_format\n"
		reader := bufio.NewReader(strings.NewReader(input))

		err := app.logIn(*reader, mockClient, mockStream)
		assert.Error(t, err)
		assert.Equal(t, ErrCredentialsFormat, err)

		mockClient.AssertExpectations(t)
		mockStream.AssertExpectations(t)
	})

	t.Run("login failed", func(t *testing.T) {
		input := "username password\n"
		reader := bufio.NewReader(strings.NewReader(input))

		mockClient.On("Login", mock.Anything, &pb.LoginRequest{Username: "username", Password: "password"}).
			Return(nil, errors.New("login failed"))

		err := app.logIn(*reader, mockClient, mockStream)
		assert.Error(t, err)
		assert.EqualError(t, err, "login failed")

		mockClient.AssertExpectations(t)
		mockStream.AssertExpectations(t)
	})
}
