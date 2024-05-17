package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"keeper/internal/client/config"

	pb "keeper/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ErrActionSelected = errors.New("выбранное действие не доступно")
var ErrCredentialsFormat = errors.New("не верный формат для логина/пароля")

type App struct {
	cfg    *config.Config
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func New(cfg *config.Config) (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(2) // Две горутины: одна для получения сообщений, другая для ввода пользователя
	return &App{cfg: cfg, ctx: ctx, cancel: cancel, wg: &wg}, nil
}

func (s *App) Run() error {
	defer s.cancel()

	// Обработка сигналов завершения
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Горутина для обработки сигналов
	go func() {
		sig := <-stopChan
		log.Printf("Received signal: %v. Shutting down...", sig)
		s.cancel()
	}()

	// подключаемся к серверу
	conn, err := grpc.DialContext(s.ctx, s.cfg.ServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v", err)
		return err
	}
	defer conn.Close()

	// стартуем стрим
	client := pb.NewKeeperServiceClient(conn)
	stream, err := client.Command(s.ctx)
	if err != nil {
		log.Printf("could not start command: %v", err)
		return err
	}

	// Запрос действия
	reader := bufio.NewReader(os.Stdin)
	action, err := s.getAction(*reader)
	if err != nil {
		return err
	}

	switch action {
	case "1":
		// Регистрация
		s.registration(*reader, client)
	case "2":
		// Вход
		s.logIn(*reader, client, stream)
	default:
		log.Printf("invalid action selected")
		return ErrActionSelected
	}

	return nil
}

func (s *App) getAction(reader bufio.Reader) (string, error) {

	// Запрос действия у пользователя
	fmt.Println("\nВыбирете действие:")
	fmt.Println("1) Register")
	fmt.Println("2) Login")
	action, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("error reading action: %v", err)
		return "", err
	}
	action = strings.TrimSpace(action)
	return action, nil
}
