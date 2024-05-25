package app

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"keeper/internal/logger"
	"keeper/internal/server/service"
	pb "keeper/proto"

	"github.com/google/uuid"
)

func newClient(stream pb.KeeperService_CommandServer) *client {
	return &client{stream: stream, ch: make(chan *pb.CommandMessage, 100), done: make(chan struct{}), state: service.CONNECTED}
}

func (s *server) Command(stream pb.KeeperService_CommandServer) error {
	client := newClient(stream)
	recvChan := make(chan *pb.CommandMessage)
	errChan := make(chan error)
	stopRecvChan := make(chan struct{})

	// Горутин для отправки сообщений клиенту
	go func() {
		for {
			select {
			case msg := <-client.ch:
				if err := client.stream.Send(msg); err != nil {
					logger.Log.Sugar().Errorf("Error sending message to %s: %v", msg.Username, err)
				}
			// завершаем горутину если контекст отменен или клиент отключился
			case <-s.ctx.Done():
				return
			case <-client.done:
				return
			}
		}
	}()

	// Горутина для получения сообщений от клиента
	go func() {
		for {
			select {
			case <-stopRecvChan:
				return
			default:
				msg, err := stream.Recv()
				if err != nil {
					errChan <- err
					return
				}
				recvChan <- msg
			}
		}
	}()

	// обрабатываем запросы клиента
	err := s.clientProcessing(client, recvChan, stopRecvChan, errChan, stream)
	if err != nil {
		return err
	}
	return nil
}

func (s *server) removeClient(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if client, exists := s.clients[clientID]; exists {
		close(client.done) // Останавливаем горутину клиента
		close(client.ch)   // Закрываем канал клиента
		delete(s.clients, clientID)
		logger.Log.Sugar().Infof("%s disconnected", clientID)
	}
}

func (s *server) clientProcessing(client *client, recvChan chan *pb.CommandMessage, stopRecvChan chan struct{}, errChan chan error, stream pb.KeeperService_CommandServer) error {
	var username string
	var clientID string
	var createdType service.DataType
	dataTitles := make(map[string]string)

	for {
		select {
		// завершаем горутину если контекст отменен
		case <-s.ctx.Done():
			return nil
		case msg := <-recvChan:

			// авторизация при подключении
			if username == "" {
				s.mu.Lock()
				username = msg.Username
				// Генерация уникального идентификатора
				clientID = username + "::" + uuid.NewString()
				s.clients[clientID] = client
				client.state = service.AUTHORIZATE
				s.mu.Unlock()
				logger.Log.Sugar().Infof("%s connected", username)
			}

			logger.Log.Sugar().Infof("Received command from %s: %s", username, msg.Message)

			// машина состояний
			switch client.state {
			case service.AUTHORIZATE:
				client.ch <- &pb.CommandMessage{Message: "\nВыбирете действие:\n1) GET\n2) CREATE"}
				client.state = service.SELECT_ACTION
			case service.SELECT_ACTION:
				switch msg.Message {
				case "1": // GET
					resultMes, err := s.getUserTitles(username, client, dataTitles)
					if err != nil {
						if errors.Is(err, ErrTitlesNotFound) {
							client.ch <- &pb.CommandMessage{Message: "\nУ вас нет сохраненных данных."}
							client.state = service.AUTHORIZATE
						}
						continue
					}
					client.ch <- &pb.CommandMessage{Message: resultMes}
					client.state = service.GET_DATA
				case "2": // CREATE
					client.ch <- &pb.CommandMessage{Message: "\nЧто хотите создать:\n1) логин/пароль\n2) текстовые данные\n3) банковскую карту"}
					client.state = service.CHOSE_CREATE_DATA
				}
			case service.GET_DATA:
				if title, ok := dataTitles[msg.Message]; ok {
					data, err := s.getData(username, title)
					if err != nil {
						continue
					}
					client.ch <- &pb.CommandMessage{Message: data}
					client.state = service.AUTHORIZATE
					dataTitles = make(map[string]string)
				}
			case service.CHOSE_CREATE_DATA:
				switch msg.Message {
				case "1": // пароли
					client.ch <- &pb.CommandMessage{Message: "\nВведите данны по шаблону: [название]::[логин]::[пароль]"}
					client.state = service.CREATE_DATA
					createdType = service.PASSWORD
				case "2": // текст
					client.ch <- &pb.CommandMessage{Message: "\nВведите данны по шаблону: [название]::[данные]"}
					client.state = service.CREATE_DATA
					createdType = service.TEXT
				case "3": // карта
					client.ch <- &pb.CommandMessage{Message: "\nВведите данны по шаблону: [название]::[номер карты]::[срок действия]::[владелец карты]::[cvv]"}
					client.state = service.CREATE_DATA
					createdType = service.CARD
				default:
					client.ch <- &pb.CommandMessage{Message: "\nВыбрано не cуществующее днйствие!\nЧто хотите создать:\n1) логин/пароль\n2) текстовые данные\n3) банковскую карту"}
				}
			case service.CREATE_DATA:
				title, err := s.createData(msg.Message, username, createdType)
				if err != nil {
					if errors.Is(err, ErrCreateFormat) {
						client.ch <- &pb.CommandMessage{Message: "\nНе верный формат данных."}
					}
					continue
				}

				client.ch <- &pb.CommandMessage{Message: "\nДанные записаны!"}
				client.state = service.AUTHORIZATE
				go s.broadcastMessage(username, clientID, title)
			}

		case err := <-errChan:
			close(stopRecvChan)
			if err == io.EOF {
				s.removeClient(clientID)
				return nil
			}
			s.removeClient(clientID)
			return err
		}
	}
}

func (s *server) broadcastMessage(username string, id string, title string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for clientID, client := range s.clients {
		var n string
		parts := strings.Split(clientID, "::")
		n, _ = parts[0], parts[1]

		if n == username && clientID != id {
			err := client.stream.Send(&pb.CommandMessage{
				Username: "server",
				Message:  fmt.Sprintf("ОБНОВЛЕНИЕ! Новая запись: %s", title),
			})
			if err != nil {
				logger.Log.Sugar().Errorf("Error sending message to %s: %v", n, err)
			}
		}
	}
}
