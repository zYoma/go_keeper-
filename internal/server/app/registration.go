package app

import (
	"context"
	"errors"
	"keeper/internal/server/service"
	"keeper/internal/server/storage/sqlite"
	pb "keeper/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	passwordHash, err := service.GetHashStr(req.Password)
	if err != nil {
		return nil, err
	}

	err = s.provider.CreateUser(ctx, req.Username, passwordHash)
	if err != nil {
		if errors.Is(err, sqlite.ErrConflict) {
			return nil, status.Error(codes.AlreadyExists, "username already exists")
		}
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return &pb.RegisterResponse{
		Message: "Регистрация завершена! Переподключитесь и войдите.",
	}, nil
}
