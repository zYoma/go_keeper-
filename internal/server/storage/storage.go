package storage

import (
	"context"
	"keeper/internal/server/service"
)

type StorageProvider interface {
	Provider
}

type Provider interface {
	Init() error
	CreateUser(ctx context.Context, username string, password string) error
	ExistUser(ctx context.Context, username, password string) error
	GetTitlesByUser(ctx context.Context, username string) ([]string, error)
	GetData(ctx context.Context, username string, title string) (string, error)
	CreateData(ctx context.Context, username string, title string, data_type service.DataType, data string) error
}
