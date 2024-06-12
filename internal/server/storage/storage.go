package storage

import (
	"context"
	"keeper/internal/server/service"
)

type Client struct {
	ClientID string
	Username string
	State    int
}

type Provider interface {
	Init() error
	CreateUser(ctx context.Context, username string, password string) error
	ExistUser(ctx context.Context, username, password string) error
	GetTitlesByUser(ctx context.Context, username string) ([]string, error)
	GetData(ctx context.Context, username string, title string) (string, error)
	CreateData(ctx context.Context, username string, title string, data_type service.DataType, data string) error
	GetAllClients(ctx context.Context) ([]Client, error)
	RemoveClient(ctx context.Context, clientID string) error
	UpdateClientState(ctx context.Context, clientID string, state service.State) error
	AddClient(ctx context.Context, clientID, username string, state service.State) error
}
