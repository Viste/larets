package api

import (
	"encoding/json"
	"github.com/Viste/larets/config"
	"github.com/Viste/larets/db"
	"github.com/Viste/larets/models"
	"github.com/Viste/larets/repositories"
	"log"
	"net/http"
)

func RunAPIServer() {
	log.Println("Запуск API сервера на порту :8080")

	if config.Config["ENABLE_DOCKER"] == "true" {
		http.HandleFunc("/docker/create", createDockerImage)
		http.HandleFunc("/docker/list", listDockerImages)
	}

	if config.Config["ENABLE_GIT"] == "true" {
		http.HandleFunc("/git/create", createGitRepository)
		http.HandleFunc("/git/list", listGitRepositories)
	}

	if config.Config["ENABLE_HELM"] == "true" {
		http.HandleFunc("/helm/create", createHelmChart)
	}

	if config.Config["ENABLE_MAVEN"] == "true" {
		http.HandleFunc("/maven/create", createMavenArtifact)
	}

	if config.Config["ENABLE_RPM"] == "true" {
		http.HandleFunc("/rpm/create", createRpmPackage)
	}

	if config.Config["ENABLE_DEB"] == "true" {
		http.HandleFunc("/deb/create", createDebPackage)
	}

	http.HandleFunc("/files", handleFiles)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createDockerImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Name string `json:"name"`
		Tag  string `json:"tag"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
		return
	}

	err := repositories.StoreDockerImage(request.Name, request.Tag)
	if err != nil {
		http.Error(w, "Ошибка сохранения Docker образа", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Docker образ успешно создан"))
}

func listDockerImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var images []models.DockerImage
	result := db.DB.Find(&images)
	if result.Error != nil {
		http.Error(w, "Ошибка получения списка Docker образов", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
}

func createGitRepository(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
		return
	}

	err := repositories.StoreGitRepository(request.Name, request.URL)
	if err != nil {
		http.Error(w, "Ошибка сохранения Git репозитория", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Git репозиторий успешно создан"))
}

func listGitRepositories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var repos []models.GitRepository
	result := db.DB.Find(&repos)
	if result.Error != nil {
		http.Error(w, "Ошибка получения списка Git репозиториев", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repos)
}

func createHelmChart(w http.ResponseWriter, r *http.Request) {
	// Логика для создания Helm чарта
	log.Println("Создание Helm чарта")
	w.Write([]byte("Helm чарт успешно создан"))
}

func createMavenArtifact(w http.ResponseWriter, r *http.Request) {
	// Логика для создания Maven артефакта
	log.Println("Создание Maven артефакта")
	w.Write([]byte("Maven артефакт успешно создан"))
}

func createRpmPackage(w http.ResponseWriter, r *http.Request) {
	// Логика для создания RPM пакета
	log.Println("Создание RPM пакета")
	w.Write([]byte("RPM пакет успешно создан"))
}

func createDebPackage(w http.ResponseWriter, r *http.Request) {
	// Логика для создания DEB пакета
	log.Println("Создание DEB пакета")
	w.Write([]byte("DEB пакет успешно создан"))
}

// handleFiles - новый обработчик для работы с файлами
func handleFiles(w http.ResponseWriter, r *http.Request) {
	// Логика для загрузки и хранения файлов
	log.Println("Обработка файлов и архивов")
	w.Write([]byte("Обработка файлов и архивов"))
}
