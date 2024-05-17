package config

import (
	"flag"
	"os"
)

var flagServerAddr string

const (
	envServerAddress = "SERVER_ADDRESS"
)

// Config определяет конфигурацию приложения, собираемую из аргументов командной строки и переменных окружения.
type Config struct {
	ServerAddr string // Адрес и порт для подключения к серверу.
}

// GetConfig парсит аргументы командной строки и переменные окружения,
// создавая и возвращая конфигурацию приложения. Приоритет имеют значения из переменных окружения.
//
// Возвращает сконфигурированный экземпляр *Config.
func GetConfig() (*Config, error) {

	// парсим аргументы командной строки
	flag.StringVar(&flagServerAddr, "a", "localhost:50051", "address and port to connect server")
	flag.Parse()

	// если есть переменные окружения, используем их значения
	if envRunAddr := os.Getenv(envServerAddress); envRunAddr != "" {
		flagServerAddr = envRunAddr
	}

	return &Config{
		ServerAddr: flagServerAddr,
	}, nil
}
