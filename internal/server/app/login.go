package app

import (
	"context"
	"keeper/internal/server/service"
	pb "keeper/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	passwordHash, err := service.GetHashStr(req.Password)
	if err != nil {
		return nil, err
	}

	err = s.provider.ExistUser(s.ctx, req.Username, passwordHash)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "wrong credentials")
	}
	return &pb.LoginResponse{
		Message: "Вы успешно вошли!",
	}, nil
}
