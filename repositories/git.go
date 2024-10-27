package repositories

import (
	"github.com/Viste/larets/db"
	"github.com/Viste/larets/models"
	"log"
)

func StoreGitRepository(name string, url string) error {
	gitRepo := models.GitRepository{Name: name, URL: url}
	result := db.DB.Create(&gitRepo)
	if result.Error != nil {
		return result.Error
	}
	log.Printf("Сохраняем Git репозиторий: %s с URL: %s", name, url)
	return nil
}

func ProxyGitRepository(repoURL string) {
	// Логика проксирования запросов к удаленным Git репозиториям
	log.Printf("Проксируем Git репозиторий: %s", repoURL)
}

func InitGitRepository() {
	log.Println("Инициализация Git репозитория")
}
