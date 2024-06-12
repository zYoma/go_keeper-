package app

import (
	"fmt"
	"testing"

	"keeper/internal/mocks"
	"keeper/internal/server/service"
	pb "keeper/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUserTitles(t *testing.T) {
	mockProvider := new(mocks.Provider)
	server := &server{
		provider: mockProvider,
	}

	username := "testuser"
	dataTitles := make(map[string]string)
	client := &client{
		ch:    make(chan *pb.CommandMessage, 1),
		state: service.CONNECTED,
	}

	t.Run("no saved data", func(t *testing.T) {
		mockProvider.On("GetTitlesByUser", mock.Anything, username).Return([]string{}, nil)

		message, err := server.getUserTitles(username, client, dataTitles)
		assert.Error(t, err)
		assert.Equal(t, ErrTitlesNotFound, err)
		assert.Equal(t, "", message)

		mockProvider.AssertExpectations(t)
		mockProvider.ExpectedCalls = nil
	})

	t.Run("titles available", func(t *testing.T) {
		titles := []string{"Title 1", "Title 2", "Title 3"}
		mockProvider.On("GetTitlesByUser", mock.Anything, username).Return(titles, nil)

		message, err := server.getUserTitles(username, client, dataTitles)
		assert.NoError(t, err)
		assert.NotEqual(t, "", message)
		assert.Equal(t, service.CONNECTED, client.state) // state should not change in this case

		expectedMessage := "\nЧто хотите получить:\n1) Title 1\n2) Title 2\n3) Title 3\n"
		assert.Equal(t, expectedMessage, message)

		for i, title := range titles {
			key := fmt.Sprintf("%d", i+1)
			assert.Equal(t, title, dataTitles[key])
		}

		mockProvider.AssertExpectations(t)
		mockProvider.ExpectedCalls = nil
	})

	t.Run("provider error", func(t *testing.T) {
		mockProvider.On("GetTitlesByUser", mock.Anything, username).Return(nil, fmt.Errorf("provider error"))

		message, err := server.getUserTitles(username, client, dataTitles)
		assert.Error(t, err)
		assert.Equal(t, "", message)
		assert.Equal(t, service.CONNECTED, client.state) // state should not change in this case

		mockProvider.AssertExpectations(t)
		mockProvider.ExpectedCalls = nil
	})
}
