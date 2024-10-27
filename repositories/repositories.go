package repositories

import (
	"github.com/Viste/larets/config"
	"log"
)

func InitRepositories() {
	if config.Config["ENABLE_DOCKER"] == "true" {
		log.Println("Инициализация Docker репозитория")
		InitDockerRepository()
	}

	if config.Config["ENABLE_GIT"] == "true" {
		log.Println("Инициализация Git репозитория")
		InitGitRepository()
	}

	if config.Config["ENABLE_HELM"] == "true" {
		log.Println("Инициализация Helm репозитория")
		// Пример инициализации Helm репозитория
		// Логика подключения и настройки Helm
	}

	if config.Config["ENABLE_MAVEN"] == "true" {
		log.Println("Инициализация Maven репозитория")
		// Пример инициализации Maven репозитория
		// Логика подключения и настройки Maven
	}

	if config.Config["ENABLE_RPM"] == "true" {
		log.Println("Инициализация RPM репозитория")
		// Пример инициализации RPM репозитория
		// Логика подключения и настройки RPM
	}

	if config.Config["ENABLE_DEB"] == "true" {
		log.Println("Инициализация DEB репозитория")
		// Пример инициализации DEB репозитория
		// Логика подключения и настройки DEB
	}
}
