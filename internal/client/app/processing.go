package app

import (
	"bufio"
	"io"

	pb "keeper/proto"
	"log"
	"os"
)

func (s *App) getData(stream pb.KeeperService_CommandClient) {
	defer s.wg.Done()
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			msgChan := make(chan *pb.CommandMessage)
			errChan := make(chan error)

			// Запуск отдельной горутины чтобы не блокироваться на `stream.Recv()`
			go func() {
				msg, err := stream.Recv()
				if err != nil {
					errChan <- err
				} else {
					msgChan <- msg
				}
			}()

			select {
			case msg := <-msgChan:
				log.Println(msg.Message)
			case err := <-errChan:
				if err == io.EOF {
					log.Printf("Stream closed by server")
					s.cancel()
					return
				}
				log.Printf("error receiving message: %v", err)
				s.cancel()
				return
			case <-s.ctx.Done():
				return
			}
		}
	}

}

func (s *App) sendData(stream pb.KeeperService_CommandClient, username string) {
	defer s.wg.Done()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			textChan := make(chan string)
			errChan := make(chan error)

			// Запуск отдельной горутины для чтобы не блокироваться на `scanner.Scan()`
			go func() {
				if scanner.Scan() {
					textChan <- scanner.Text()
				} else {
					if err := scanner.Err(); err != nil {
						errChan <- err
					} else {
						errChan <- io.EOF
					}
				}
			}()

			select {
			case msg := <-textChan:
				err := s.send(stream, username, msg)
				if err != nil {
					s.cancel()
				}
			case err := <-errChan:
				if err != io.EOF {
					log.Printf("error reading from input: %v", err)
				}
				s.cancel()
				return
			case <-s.ctx.Done():
				return
			}
		}
	}

}

func (s *App) send(stream pb.KeeperService_CommandClient, username string, msg string) error {
	if err := stream.Send(&pb.CommandMessage{Username: username, Message: msg}); err != nil {
		log.Printf("error sending message: %v", err)
		return err
	}
	return nil
}
