package app

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var ErrTitlesNotFound = errors.New("titles not found")

func (s *server) getUserTitles(username string, client *client, dataTitles map[string]string) (string, error) {
	userTitles, err := s.provider.GetTitlesByUser(s.ctx, username)
	if err != nil {
		return "", err
	}
	if len(userTitles) == 0 {
		return "", ErrTitlesNotFound

	}
	// Перенос значений из titles в dataTitles
	for i, title := range userTitles {
		key := fmt.Sprintf("%d", i+1) // Создание ключа "1", "2", ...
		dataTitles[key] = title       // Присвоение значения из titles
	}

	// Сортировка ключей
	var keys []int
	for key := range dataTitles {
		numKey, _ := strconv.Atoi(key)
		keys = append(keys, numKey)
	}
	sort.Ints(keys)

	// Создание строки с перечислением элементов dataTitles
	var builder strings.Builder
	builder.WriteString("\nЧто хотите получить:\n")

	for _, numKey := range keys {
		key := fmt.Sprintf("%d", numKey)
		builder.WriteString(fmt.Sprintf("%s) %s\n", key, dataTitles[key]))
	}

	return builder.String(), nil
}
