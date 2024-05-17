package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"keeper/internal/logger"
	"keeper/internal/server/config"
	"keeper/internal/server/service"
	"keeper/internal/server/storage"
	"sync"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

// возможные ошибки пакета
var (
	// ErrCreateConnect описывает ошибку создания подключения к базе данных.
	ErrCreateConnect = errors.New("unable to create connection")

	// ErrCreateUser описывает ошибку, возникающую, когда не удалось создать пользователя.
	ErrCreateUser = errors.New("create user")
	// ErrCreateData описывает ошибку сохранения данных в базе данных.
	ErrCreateData = errors.New("create data")
	// ErrCreateTable описывает ошибку создания таблиц в базе данных.
	ErrCreateTable = errors.New("creating tables")
	// ErrConflict описывает ошибку конфликта при попытке вставки, который уже существует.
	ErrConflict = errors.New("already exist")
	// ErrUserNotFound описывает ошибку получения пользоввателя из базы данных.
	ErrUserNotFound = errors.New("user not found")
	// ErrDataNotFound описывает ошибку получения пользоввателя из базы данных.
	ErrDataNotFound = errors.New("data not found")
)

// Storage реализует интерфейс StorageProvider и предоставляет методы для работы с хранилищем URL.
type Storage struct {
	db    *sql.DB    // Соединение с базой данных.
	mutex sync.Mutex // Мьютекс для синхронизации доступа к базе данных.
}

// New инициализирует новый экземпляр Storage с подключением к базе данных, указанной в конфигурации.
func NewProvider(cfg *config.Config) (storage.StorageProvider, error) {
	db, err := sql.Open("sqlite3", cfg.DSN)
	if err != nil {
		logger.Log.Sugar().Errorf("Не удалось подключиться к БД: %s", err)
		return nil, ErrCreateConnect
	}

	// Проверка соединения с базой данных
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

// Init выполняет инициализацию хранилища, включая создание необходимых таблиц.
func (s *Storage) Init() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Sugar().Errorf("Не удалось создать таблицу: %s", err)
		return ErrCreateTable
	}

	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			logger.Log.Sugar().Errorf("Ошибка при откате транзакции: %v", err)
		}
	}()

	_, err = tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS users (
            username VARCHAR(255) PRIMARY KEY,
            password_hash TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `)
	if err != nil {
		logger.Log.Sugar().Errorf("Ошибка при создании таблицы users: %s", err)
		return ErrCreateTable
	}

	_, err = tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS user_data (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title VARCHAR(50) NOT NULL,
            username VARCHAR(255) REFERENCES users(username) ON DELETE CASCADE,
            data_type INTEGER NOT NULL,
            data TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `)
	if err != nil {
		logger.Log.Sugar().Errorf("Ошибка при создании таблицы user_data: %s", err)
		return ErrCreateTable
	}

	_, err = tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS idx_title_username_unique ON user_data(title, username);`)
	if err != nil {
		logger.Log.Sugar().Errorf("Ошибка при создании индекса: %s", err)
		return ErrCreateTable
	}

	if err := tx.Commit(); err != nil {
		logger.Log.Sugar().Errorf("Ошибка при коммите транзакции: %s", err)
		return ErrCreateTable
	}

	return nil
}

func (s *Storage) CreateUser(ctx context.Context, username string, password string) error {
	// Подготовка SQL-запроса для вставки
	query := `
        INSERT INTO users (username, password_hash)
        VALUES (?, ?)
    `
	// Выполнение SQL-запроса
	_, err := s.db.ExecContext(ctx, query, username, password)
	if err != nil {
		// Проверка, если ошибка связана с существующим username
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			return ErrConflict
		}
		return ErrCreateUser
	}

	return nil
}

// ExistUser проверяет существование пользователя в базе данных.
func (s *Storage) ExistUser(ctx context.Context, username, password string) error {
	// Подготовка SQL-запроса для выборки пароля
	query := `
	SELECT EXISTS (SELECT 1 FROM users WHERE username = ? AND password_hash = ?)
	`

	var isExist int
	err := s.db.QueryRowContext(ctx, query, username, password).Scan(&isExist)
	if err != nil {
		return err
	}

	// Если isExist равен 0, значит пользователь не найден
	if isExist == 0 {
		return ErrUserNotFound
	}

	return nil // Возвращаем nil, если пользователь найден
}

// GetDataByUser возвращает все значения title для заданного username из таблицы user_data
func (s *Storage) GetTitlesByUser(ctx context.Context, username string) ([]string, error) {
	// Подготовка SQL-запроса для выборки title
	query := `
        SELECT title FROM user_data WHERE username = ?
    `

	// Выполнение SQL-запроса с использованием контекста
	rows, err := s.db.QueryContext(ctx, query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var titles []string
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			return nil, err
		}
		titles = append(titles, title)
	}

	// Проверка на наличие ошибок после обработки строк
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return titles, nil
}

// GetData возвращает значение data для заданных username и title из таблицы user_data
func (s *Storage) GetData(ctx context.Context, username string, title string) (string, error) {
	// Подготовка SQL-запроса для выборки data
	query := `
        SELECT data FROM user_data WHERE username = ? AND title = ?
    `

	var data string

	// Выполнение SQL-запроса с использованием контекста
	err := s.db.QueryRowContext(ctx, query, username, title).Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrDataNotFound
		}
		logger.Log.Sugar().Errorf("Error get data: %v", err)
		return "", err
	}

	return data, nil
}

// CreateData добавляет новую запись в таблицу user_data
func (s *Storage) CreateData(ctx context.Context, username string, title string, data_type service.DataType, data string) error {
	// Подготовка SQL-запроса для вставки
	query := `
        INSERT INTO user_data (username, title, data_type, data)
        VALUES (?, ?, ?, ?)
    `

	// Выполнение SQL-запроса с использованием контекста
	_, err := s.db.ExecContext(ctx, query, username, title, data_type, data)
	if err != nil {
		logger.Log.Sugar().Errorf("Error create data: %v", err)
		return ErrCreateData
	}

	return nil
}
