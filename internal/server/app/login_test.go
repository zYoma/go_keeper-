package app

import (
	"context"
	"errors"
	"testing"

	"keeper/internal/mocks"
	pb "keeper/proto"

	_ "github.com/mattn/go-sqlite3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestLogin тестирует метод Login
func TestLogin(t *testing.T) {
	mockProvider := new(mocks.Provider)
	server := &server{
		provider: mockProvider,
	}

	ctx := context.Background()
	req := &pb.LoginRequest{
		Username: "testuser",
		Password: "password",
	}

	t.Run("successful login", func(t *testing.T) {
		mockProvider.On("ExistUser", mock.Anything, req.Username, mock.Anything).Return(nil)

		resp, err := server.Login(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Вы успешно вошли!", resp.Message)

		mockProvider.AssertExpectations(t)
		mockProvider.ExpectedCalls = nil
	})

	t.Run("wrong credentials", func(t *testing.T) {
		mockProvider.On("ExistUser", mock.Anything, req.Username, mock.Anything).Return(errors.New("wrong credentials"))

		resp, err := server.Login(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, st.Code())

		mockProvider.AssertExpectations(t)
		mockProvider.ExpectedCalls = nil
	})
}
