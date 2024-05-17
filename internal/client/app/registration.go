package app

import (
	"bufio"
	"fmt"
	"log"

	pb "keeper/proto"
	"strings"
)

func getCredentials(reader bufio.Reader) (string, string, error) {
	fmt.Println("Введите username и password через пробел. Пример: username password")
	credentials, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("error reading credentials: %v", err)
		return "", "", err
	}
	credentials = strings.TrimSpace(credentials)
	parts := strings.Split(credentials, " ")
	if len(parts) != 2 {
		log.Printf("invalid input format")
		return "", "", ErrCredentialsFormat
	}
	var username, password string
	username, password = parts[0], parts[1]
	return username, password, nil
}

func (s *App) registration(reader bufio.Reader, client pb.KeeperServiceClient) error {
	username, password, err := getCredentials(reader)
	if err != nil {
		return err
	}

	// Отправка запроса на регистрацию
	resp, err := client.Register(s.ctx, &pb.RegisterRequest{Username: username, Password: password})
	if err != nil {
		log.Printf("registration failed: %v", err)
		return err
	}
	fmt.Println("Registration response:", resp.Message)
	return nil
}
