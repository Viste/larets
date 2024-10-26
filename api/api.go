package api

import (
	"log"
	"net/http"
)

func RunAPIServer() {
	log.Println("Запуск API сервера на порту :8080")
	http.HandleFunc("/images", handleImages)
	http.HandleFunc("/git", handleGit)
	http.HandleFunc("/helm", handleHelm)
	http.HandleFunc("/maven", handleMaven)
	http.HandleFunc("/rpm", handleRpm)
	http.HandleFunc("/deb", handleDeb)
	// дополнительные маршруты для работы с CRUD

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleImages(w http.ResponseWriter, r *http.Request) {
	// Обработка запросов по Docker образам
	log.Println("Обработка запроса Docker образов")
	w.Write([]byte("Обработка Docker образов"))
}

func handleGit(w http.ResponseWriter, r *http.Request) {
	// Обработка запросов по Git репозиториям
	log.Println("Обработка запроса Git репозиториев")
	w.Write([]byte("Обработка Git репозиториев"))
}

func handleHelm(w http.ResponseWriter, r *http.Request) {
	// Обработка запросов по Helm репозиториям
	log.Println("Обработка запроса Helm репозиториев")
	w.Write([]byte("Обработка Helm репозиториев"))
}

func handleMaven(w http.ResponseWriter, r *http.Request) {
	// Обработка запросов по Maven репозиториям
	log.Println("Обработка запроса Maven репозиториев")
	w.Write([]byte("Обработка Maven репозиториев"))
}

func handleRpm(w http.ResponseWriter, r *http.Request) {
	// Обработка запросов по RPM репозиториям
	log.Println("Обработка запроса RPM репозиториев")
	w.Write([]byte("Обработка RPM репозиториев"))
}

func handleDeb(w http.ResponseWriter, r *http.Request) {
	// Обработка запросов по DEB репозиториям
	log.Println("Обработка запроса DEB репозиториев")
	w.Write([]byte("Обработка DEB репозиториев"))
}
