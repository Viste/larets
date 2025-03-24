package main

import (
	"github.com/Viste/larets/api"
	"github.com/Viste/larets/config"
	"github.com/Viste/larets/db"
	"log"
	"os"
)

func main() {
	log.Println("Larets - менеджер-репозиториев")

	config.LoadConfig()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("Не задана переменная окружения DATABASE_URL")
	}

	err := db.InitDB(connStr)
	if err != nil {
		log.Fatal("Ошибка при инициализации базы данных: ", err)
	}
	log.Println("База данных успешно инициализирована")

	err = db.EnsureStorageDirs(config.Config.StorageBasePath)
	if err != nil {
		log.Fatal("Ошибка при создании директорий хранилища: ", err)
	}

	api.RunAPIServer()
}
