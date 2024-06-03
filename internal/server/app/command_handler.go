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

		err := s.provider.RemoveClient(s.ctx, clientID)
		if err != nil {
			logger.Log.Sugar().Errorf("Failed to remove client from DB: %v", err)
		}

		logger.Log.Sugar().Infof("%s disconnected", clientID)
	}
}

func (s *server) addClient(username string, clientID string, client *client) {

	s.clients[clientID] = client
	err := s.provider.AddClient(s.ctx, clientID, username, client.state)
	if err != nil {
		logger.Log.Sugar().Errorf("Failed to add client to DB: %v", err)
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
				s.addClient(username, clientID, client)
				s.mu.Unlock()
				logger.Log.Sugar().Infof("%s connected", username)
			}

			logger.Log.Sugar().Infof("Received command from %s: %s", username, msg.Message)

			// машина состояний
			switch client.state {
			case service.CONNECTED:
				client.ch <- &pb.CommandMessage{Message: "\nВыбирете действие:\n1) GET\n2) CREATE"}
				err := s.updateState(client, clientID, service.SELECT_ACTION)
				if err != nil {
					continue
				}
			case service.SELECT_ACTION:
				switch msg.Message {
				case "1": // GET
					resultMes, err := s.getUserTitles(username, client, dataTitles)
					if err != nil {
						if errors.Is(err, ErrTitlesNotFound) {
							client.ch <- &pb.CommandMessage{Message: "\nУ вас нет сохраненных данных."}
							err := s.updateState(client, clientID, service.CONNECTED)
							if err != nil {
								continue
							}
						}
						continue
					}
					client.ch <- &pb.CommandMessage{Message: resultMes}
					err = s.updateState(client, clientID, service.GET_DATA)
					if err != nil {
						continue
					}
				case "2": // CREATE
					client.ch <- &pb.CommandMessage{Message: "\nЧто хотите создать:\n1) логин/пароль\n2) текстовые данные\n3) банковскую карту\n4) бинарные данные"}
					err := s.updateState(client, clientID, service.CHOSE_CREATE_DATA)
					if err != nil {
						continue
					}
				}
			case service.GET_DATA:
				if title, ok := dataTitles[msg.Message]; ok {
					data, err := s.getData(username, title)
					if err != nil {
						continue
					}
					client.ch <- &pb.CommandMessage{Message: data}
					err = s.updateState(client, clientID, service.CONNECTED)
					if err != nil {
						continue
					}
					dataTitles = make(map[string]string)
				}
			case service.CHOSE_CREATE_DATA:
				switch msg.Message {
				case "1": // пароли
					client.ch <- &pb.CommandMessage{Message: "\nВведите данны по шаблону: [название]::[логин]::[пароль]::[метадата]"}
					err := s.updateState(client, clientID, service.CREATE_DATA)
					if err != nil {
						continue
					}
					createdType = service.PASSWORD
				case "2": // текст
					client.ch <- &pb.CommandMessage{Message: "\nВведите данны по шаблону: [название]::[данные]::[метадата]"}
					err := s.updateState(client, clientID, service.CREATE_DATA)
					if err != nil {
						continue
					}
					createdType = service.TEXT
				case "3": // карта
					client.ch <- &pb.CommandMessage{Message: "\nВведите данны по шаблону: [название]::[номер карты]::[срок действия]::[владелец карты]::[cvv]::[метадата]"}
					err := s.updateState(client, clientID, service.CREATE_DATA)
					if err != nil {
						continue
					}
					createdType = service.CARD
				case "4": // бинарные данные
					client.ch <- &pb.CommandMessage{Message: "\nВведите данны по шаблону: [название]::[данные]::[метадата]"}
					err := s.updateState(client, clientID, service.CREATE_DATA)
					if err != nil {
						continue
					}
					createdType = service.BYTE
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
				err = s.updateState(client, clientID, service.CONNECTED)
				if err != nil {
					continue
				}
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
			msg := &pb.CommandMessage{
				Username: "server",
				Message:  fmt.Sprintf("ОБНОВЛЕНИЕ! Новая запись: %s", title),
			}

			err := client.stream.Send(msg)
			if err != nil {
				logger.Log.Sugar().Errorf("Error sending message to %s: %v", n, err)

				// Делаем две дополнительные попытки переотправки
				maxRetries := 2
				for retries := 0; retries < maxRetries; retries++ {
					err = client.stream.Send(msg)
					if err == nil {
						break
					}
					logger.Log.Sugar().Errorf("Retry %d: Error sending message to %s: %v", retries+1, n, err)
				}
			}
		}
	}
}

func (s *server) updateState(client *client, clientID string, state service.State) error {
	client.state = state
	err := s.provider.UpdateClientState(s.ctx, clientID, service.SELECT_ACTION)
	if err != nil {
		logger.Log.Sugar().Errorf("Failed to update client state: %v", err)
		return err
	}
	return nil
}
