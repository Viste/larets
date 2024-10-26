package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

func InitDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Проверка подключения
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Подключение к базе данных успешно установлено")
	return db, nil
}
