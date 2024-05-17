package app

import (
	"context"
	"errors"
	"testing"

	"keeper/internal/mocks"
	"keeper/internal/server/storage/sqlite"
	pb "keeper/proto"

	_ "github.com/mattn/go-sqlite3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestRegister тестирует метод Register
func TestRegister(t *testing.T) {
	mockProvider := new(mocks.Provider)
	server := &server{
		provider: mockProvider,
	}

	ctx := context.Background()
	req := &pb.RegisterRequest{
		Username: "testuser",
		Password: "password",
	}

	t.Run("successful registration", func(t *testing.T) {
		mockProvider.On("CreateUser", ctx, req.Username, mock.Anything).Return(nil)

		resp, err := server.Register(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Регистрация завершена! Переподключитесь и войдите.", resp.Message)

		mockProvider.AssertExpectations(t)
		// Очищаем ожидаемые вызовы после завершения теста
		mockProvider.ExpectedCalls = nil
	})

	t.Run("user already exists", func(t *testing.T) {
		mockProvider.On("CreateUser", ctx, req.Username, mock.Anything).Return(sqlite.ErrConflict)

		resp, err := server.Register(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.AlreadyExists, st.Code())

		mockProvider.AssertExpectations(t)
		// Очищаем ожидаемые вызовы после завершения теста
		mockProvider.ExpectedCalls = nil
	})

	t.Run("internal error", func(t *testing.T) {
		mockProvider.On("CreateUser", ctx, req.Username, mock.Anything).Return(errors.New("internal error"))

		resp, err := server.Register(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())

		mockProvider.AssertExpectations(t)
		// Очищаем ожидаемые вызовы после завершения теста
		mockProvider.ExpectedCalls = nil
	})
}
