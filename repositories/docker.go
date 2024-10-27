package repositories

import (
	"github.com/Viste/larets/db"
	"github.com/Viste/larets/models"
	"log"
)

func StoreDockerImage(imageName string, tag string) error {
	dockerImage := models.DockerImage{Name: imageName, Tag: tag}
	result := db.DB.Create(&dockerImage)
	if result.Error != nil {
		return result.Error
	}
	log.Printf("Сохраняем образ: %s с тэгом: %s", imageName, tag)
	return nil
}

func ProxyDockerRepository(repoURL string) {
	// Логика проксирования запросов к удаленным Docker репозиториям
	log.Printf("Проксируем Docker репозиторий: %s", repoURL)
}

func InitDockerRepository() {
	log.Println("Инициализация Docker репозитория")
}
