package config

import (
	"flag"
	"os"
)

var flagRunAddr string
var flagLogLevel string
var flagDSN string
var flagSecret string
var flagCertPath string
var flagCertKeyPath string

const (
	envServerAddress = "SERVER_ADDRESS"
	envLoggerLevel   = "LOG_LEVEL"
	envDSN           = "DATABASE_DSN"
	envSecret        = "SECRET"
	envCertPath      = "CERT_PATH"
	envCertKeyPath   = "CERT_KEY_PATH"
)

// Config определяет конфигурацию приложения, собираемую из аргументов командной строки и переменных окружения.
type Config struct {
	RunAddr     string // Адрес и порт для запуска сервера.
	LogLevel    string // Уровень логирования.
	DSN         string // Data Source Name для подключения к БД.
	Secret      string // Секрет для шифрования данных.
	CertPath    string // путь до файла с сертификатом
	CertKeyPath string // путь до ключа
}

// GetConfig парсит аргументы командной строки и переменные окружения,
// создавая и возвращая конфигурацию приложения. Приоритет имеют значения из переменных окружения.
//
// Возвращает сконфигурированный экземпляр *Config.
func GetConfig() (*Config, error) {

	// парсим аргументы командной строки
	flag.StringVar(&flagRunAddr, "a", "localhost:50051", "address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "info", "log level")
	flag.StringVar(&flagDSN, "d", "DB.db", "DB DSN")
	flag.StringVar(&flagSecret, "j", "thisis32byteencryptionkey1234567", "secret for encryption")
	flag.StringVar(&flagCertPath, "cr", "certs/keeper.crt", "path to cert")
	flag.StringVar(&flagCertKeyPath, "ck", "certs/key.pem", "path to cert key")
	flag.Parse()

	// если есть переменные окружения, используем их значения
	if envRunAddr := os.Getenv(envServerAddress); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}

	if envLogLevel := os.Getenv(envLoggerLevel); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}

	if envDBDSN := os.Getenv(envDSN); envDBDSN != "" {
		flagDSN = envDBDSN
	}
	if envAppSecret := os.Getenv(envSecret); envAppSecret != "" {
		flagSecret = envAppSecret
	}
	if envCert := os.Getenv(envCertPath); envCert != "" {
		flagCertPath = envCert
	}
	if envCertKey := os.Getenv(envCertKeyPath); envCertKey != "" {
		flagCertKeyPath = envCertKey
	}

	return &Config{
		RunAddr:     flagRunAddr,
		LogLevel:    flagLogLevel,
		DSN:         flagDSN,
		Secret:      flagSecret,
		CertPath:    flagCertPath,
		CertKeyPath: flagCertKeyPath,
	}, nil
}
