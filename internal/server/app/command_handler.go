package app

import (
	"errors"
	"io"

	"keeper/internal/logger"
	"keeper/internal/server/service"
	pb "keeper/proto"
)

func (s *server) Command(stream pb.KeeperService_CommandServer) error {
	client := &client{stream: stream, ch: make(chan *pb.CommandMessage, 100), done: make(chan struct{}), state: service.CONNECTED}
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
	err := s.clientProcessing(client, recvChan, stopRecvChan, errChan)
	if err != nil {
		return err
	}
	return nil
}

func (s *server) removeClient(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if client, exists := s.clients[username]; exists {
		close(client.done) // Останавливаем горутину клиента
		close(client.ch)   // Закрываем канал клиента
		delete(s.clients, username)
		logger.Log.Sugar().Infof("%s disconnected", username)
	}
}

func (s *server) clientProcessing(client *client, recvChan chan *pb.CommandMessage, stopRecvChan chan struct{}, errChan chan error) error {
	var username string
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
				s.clients[username] = client
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
				}
			case service.CREATE_DATA:
				err := s.createData(msg.Message, username, createdType)
				if err != nil {
					if errors.Is(err, ErrCreateFormat) {
						client.ch <- &pb.CommandMessage{Message: "\nНе верный формат данных."}
					}
					continue
				}

				client.ch <- &pb.CommandMessage{Message: "\nДанные записаны!"}
				client.state = service.AUTHORIZATE
			}

		case err := <-errChan:
			close(stopRecvChan)
			if err == io.EOF {
				s.removeClient(username)
				return nil
			}
			s.removeClient(username)
			return err
		}
	}
}
