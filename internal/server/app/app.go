package app

import (
	"context"
	"errors"
	"keeper/internal/logger"
	"keeper/internal/server/config"
	"keeper/internal/server/service"
	"keeper/internal/server/storage"
	"keeper/internal/server/storage/sqlite"
	pb "keeper/proto"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type client struct {
	stream pb.KeeperService_CommandServer
	ch     chan *pb.CommandMessage
	done   chan struct{}
	state  service.State
}

type server struct {
	pb.UnimplementedKeeperServiceServer
	clients  map[string]*client
	mu       sync.Mutex
	cfg      *config.Config
	provider storage.Provider
	ctx      context.Context
	cancel   context.CancelFunc
}

// ErrServerStoped описывает ошибку, возникающую при остановке сервера.
var ErrServerStoped = errors.New("server stoped")

// ErrServerStart описывает ошибку, возникающую при старте сервера.
var ErrServerStart = errors.New("server start")

func New(cfg *config.Config) (*server, error) {
	// создаем провайдер для storage
	provider, err := sqlite.NewProvider(cfg)
	if err != nil {
		return nil, err
	}

	// инициализируем провайдера
	if err := provider.Init(); err != nil {
		return nil, err
	}

	// контекст необходим для остановки всех горутин
	ctx, cancel := context.WithCancel(context.Background())
	return &server{clients: make(map[string]*client), cfg: cfg, provider: provider, ctx: ctx, cancel: cancel}, nil
}

func (s *server) Run() error {
	// Загрузка сертификата сервера и закрытого ключа
	creds, err := credentials.NewServerTLSFromFile(s.cfg.CertPath, s.cfg.CertKeyPath)
	if err != nil {
		logger.Log.Sugar().Errorf("Failed to generate credentials: %v", err)
		return ErrServerStart
	}

	lis, err := net.Listen("tcp", s.cfg.RunAddr)
	if err != nil {
		logger.Log.Sugar().Errorf("failed to listen: %v", err)
		return ErrServerStart
	}

	gs := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterKeeperServiceServer(gs, s)

	// Создание канала для ошибок
	errChan := make(chan error)

	// запустить сервис
	logger.Log.Sugar().Infof("start application at %v", lis.Addr())
	go func() {
		if err := gs.Serve(lis); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// канал для сигналов
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Ждем либо сигнала об остановке, либо ошибки
	select {
	case <-stopChan:
		logger.Log.Sugar().Info("Shutting down server...")
	case err := <-errChan:
		logger.Log.Sugar().Errorf("Server error: %v", err)
	}

	// останавливаем сервер
	s.Shutdown(gs)
	return nil
}

func (s *server) Shutdown(gs *grpc.Server) {
	// отменяем контекст, чтобы остановить все горутины
	s.cancel()

	// Закрытие всех клиентских соединений и остановка их горутин
	s.mu.Lock()
	for username, client := range s.clients {
		close(client.done)
		close(client.ch)
		delete(s.clients, username)
		logger.Log.Sugar().Infof("%s left the chat", username)
	}
	s.mu.Unlock()

	// контекст для ожидания остановки сервера
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// останавливаем сервер и сигнализируем об этом в канал
	stopped := make(chan struct{})
	go func() {
		gs.GracefulStop()
		close(stopped)
	}()

	// либо сервер и все горутины завершены, либо вышло время ожидания и произошла принудительная остановка
	select {
	case <-stopped:
		logger.Log.Sugar().Info("Server stopped gracefully")
	case <-ctx.Done():
		logger.Log.Sugar().Info("Graceful stop timed out, forcing shutdown")
		gs.Stop()
	}
}

func (s *server) loadClientsFromDB() error {
	clients, err := s.provider.GetAllClients(s.ctx)
	if err != nil {
		return err
	}

	for _, clientData := range clients {
		client := &client{
			ch:    make(chan *pb.CommandMessage, 100),
			done:  make(chan struct{}),
			state: service.State(clientData.State),
		}
		s.addClient(clientData.Username, clientData.ClientID, client)
	}
	return nil
}
