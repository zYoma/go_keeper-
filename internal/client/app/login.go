package app

import (
	"bufio"
	"fmt"
	"log"

	pb "keeper/proto"
)

func (s *App) logIn(reader bufio.Reader, client pb.KeeperServiceClient, stream pb.KeeperService_CommandClient) error {
	username, password, err := getCredentials(reader)
	if err != nil {
		return err
	}

	// Отправка запроса на авторизацию
	resp, err := client.Login(s.ctx, &pb.LoginRequest{Username: username, Password: password})
	if err != nil {
		log.Printf("login failed: %v", err)
		return err
	}
	fmt.Println(resp.Message)

	s.startSession(username, stream)
	return nil
}

func (s *App) startSession(username string, stream pb.KeeperService_CommandClient) error {

	// Отправка начального сообщения для инициализации
	err := s.send(stream, username, "/init")
	if err != nil {
		return err
	}

	// Горутина для получения сообщений от сервера
	go s.getData(stream)

	// Горутина для чтения пользовательского ввода и отправки сообщений
	go s.sendData(stream, username)

	s.wg.Wait() // Ожидание завершения горутины
	stream.CloseSend()
	return nil
}
