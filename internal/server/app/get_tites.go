package app

import (
	"fmt"
	"keeper/internal/server/service"
	pb "keeper/proto"
	"strings"
)

func (s *server) getUserTitles(username string, client *client, dataTitles map[string]string) (string, error) {
	userTitles, err := s.provider.GetTitlesByUser(s.ctx, username)
	if err != nil {
		return "", err
	}
	if len(userTitles) == 0 {
		client.ch <- &pb.CommandMessage{Message: "\nУ вас нет сохраненных данных."}
		client.state = service.AUTHORIZATE
		return "", err

	}
	// Перенос значений из titles в dataTitles
	for i, title := range userTitles {
		key := fmt.Sprintf("%d", i+1) // Создание ключа "1", "2", ...
		dataTitles[key] = title       // Присвоение значения из titles
	}
	// Создание строки с перечислением элементов dataTitles
	var builder strings.Builder
	builder.WriteString("\nЧто хотите получить:\n")

	for key, value := range dataTitles {
		builder.WriteString(fmt.Sprintf("%s) %s\n", key, value))
	}

	return builder.String(), nil
}
