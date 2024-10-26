package main

import (
	"github.com/Viste/larets/api"
	"github.com/Viste/larets/config"
	"github.com/Viste/larets/db"
	"log"
	"os"
)

func main() {
	// Загрузка конфигурации
	config.LoadConfig()

	// Чтение конфигурации из переменных окружения
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("Не задана переменная окружения DATABASE_URL")
	}

	// Инициализация базы данных
	database, err := db.InitDB(connStr)
	if err != nil {
		log.Fatal("Ошибка при инициализации базы данных: ", err)
	}
	defer database.Close()

	log.Println("База данных успешно инициализирована")

	// Запуск API сервера
	api.RunAPIServer()
}
