package repositories

import (
	"larets/config"
	"log"
)

func InitRepositories() {
	if config.Config["ENABLE_DOCKER"] == "true" {
		log.Println("Инициализация Docker репозитория")
		// Здесь добавить инициализацию для работы с Docker
	}

	if config.Config["ENABLE_GIT"] == "true" {
		log.Println("Инициализация Git репозитория")
		// Здесь добавить инициализацию для работы с Git
	}

	if config.Config["ENABLE_HELM"] == "true" {
		log.Println("Инициализация Helm репозитория")
		// Здесь добавить инициализацию для работы с Helm
	}

	if config.Config["ENABLE_MAVEN"] == "true" {
		log.Println("Инициализация Maven репозитория")
		// Здесь добавить инициализацию для работы с Maven
	}

	if config.Config["ENABLE_RPM"] == "true" {
		log.Println("Инициализация RPM репозитория")
		// Здесь добавить инициализацию для работы с RPM
	}

	if config.Config["ENABLE_DEB"] == "true" {
		log.Println("Инициализация DEB репозитория")
		// Здесь добавить инициализацию для работы с DEB
	}
}
