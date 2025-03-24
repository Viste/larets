package db

import (
	"github.com/Viste/larets/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
)

var DB *gorm.DB

func InitDB(connStr string) error {
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return err
	}

	log.Println("Подключение к базе данных успешно установлено")
	DB = db

	err = migrateDB(db)
	if err != nil {
		return err
	}

	return nil
}

func migrateDB(db *gorm.DB) error {
	log.Println("Запуск миграций базы данных...")

	err := db.AutoMigrate(
		&models.DockerRepository{},
		&models.GitRepository{},
		&models.HelmRepository{},
		&models.GroupMember{},
		&models.Artifact{},
		&models.DockerImage{},
		&models.HelmChart{},
		&models.StoredFile{},
	)

	if err != nil {
		log.Printf("Ошибка выполнения миграций: %v", err)
		return err
	}

	log.Println("Миграции успешно выполнены")
	return nil
}

func EnsureStorageDirs(basePath string) error {
	log.Printf("Создание структуры директорий для хранилища в %s", basePath)

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		err = os.MkdirAll(basePath, 0755)
		if err != nil {
			return err
		}
	}

	dirs := []string{
		filepath.Join(basePath, "docker"),
		filepath.Join(basePath, "git"),
		filepath.Join(basePath, "helm"),
		filepath.Join(basePath, "temp"),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
		}
	}

	log.Println("Структура директорий для хранилища успешно создана")
	return nil
}
