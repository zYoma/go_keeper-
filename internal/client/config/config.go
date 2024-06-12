package config

import (
	"flag"
	"os"
)

var flagServerAddr string
var flagCertPath string

const (
	envServerAddress = "SERVER_ADDRESS"
	envCertPath      = "CERT_PATH"
)

// Config определяет конфигурацию приложения, собираемую из аргументов командной строки и переменных окружения.
type Config struct {
	ServerAddr string // Адрес и порт для подключения к серверу.
	CertPath   string // путь до файла с сертификатом
}

// GetConfig парсит аргументы командной строки и переменные окружения,
// создавая и возвращая конфигурацию приложения. Приоритет имеют значения из переменных окружения.
//
// Возвращает сконфигурированный экземпляр *Config.
func GetConfig() (*Config, error) {

	// парсим аргументы командной строки
	flag.StringVar(&flagServerAddr, "a", "localhost:50051", "address and port to connect server")
	flag.StringVar(&flagCertPath, "cr", "certs/keeper.crt", "path to cert")
	flag.Parse()

	// если есть переменные окружения, используем их значения
	if envRunAddr := os.Getenv(envServerAddress); envRunAddr != "" {
		flagServerAddr = envRunAddr
	}
	if envCert := os.Getenv(envCertPath); envCert != "" {
		flagCertPath = envCert
	}

	return &Config{
		ServerAddr: flagServerAddr,
		CertPath:   flagCertPath,
	}, nil
}
