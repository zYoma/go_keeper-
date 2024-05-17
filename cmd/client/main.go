package main

import (
	"keeper/internal/client/app"
	"keeper/internal/client/config"
)

func main() {
	// получаем конфигурацию
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	// инициализация приложения
	application, err := app.New(cfg)
	if err != nil {
		panic(err)
	}

	// запускаем приложение
	if err := application.Run(); err != nil {
		panic(err)
	}

}
