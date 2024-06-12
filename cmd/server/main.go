package main

import (
	"errors"

	"keeper/internal/logger"
	"keeper/internal/server/app"
	"keeper/internal/server/config"
)

func main() {
	// получаем конфигурацию
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	// инициализируем логер
	if err = logger.Initialize(cfg.LogLevel); err != nil {
		panic(err)
	}

	// инициализация приложения
	application, err := app.New(cfg)
	if err != nil {
		panic(err)
	}

	// запускаем приложение
	if err := application.Run(); err != nil {
		if errors.Is(err, app.ErrServerStoped) {
			logger.Log.Sugar().Infoln("stopping application")
			return
		}

		panic(err)
	}

}
