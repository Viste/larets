package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Viste/larets/config"
	"github.com/Viste/larets/db"
	"github.com/Viste/larets/models"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DockerService struct{}

func (s *DockerService) CreateRepository(name, description string, repoType models.RepositoryType, url string) error {
	var count int64
	db.DB.Model(&models.DockerRepository{}).Where("name = ?", name).Count(&count)
	if count > 0 {
		return errors.New("репозиторий с таким именем уже существует")
	}

	storagePath := filepath.Join(config.Config.DockerStorage, name)
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории хранилища: %w", err)
	}

	repo := models.DockerRepository{
		BaseRepository: models.BaseRepository{
			Name:        name,
			Description: description,
			Type:        repoType,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		URL:          url,
		IndexType:    "v2",
		CacheEnabled: true,
		CacheTTL:     config.Config.DefaultCacheTTL,
		StoragePath:  storagePath,
	}

	if err := db.DB.Create(&repo).Error; err != nil {
		os.RemoveAll(storagePath)
		return fmt.Errorf("ошибка сохранения репозитория: %w", err)
	}

	log.Printf("Создан Docker репозиторий: %s, тип: %s", name, repoType)
	return nil
}

func (s *DockerService) ListRepositories() ([]models.DockerRepository, error) {
	var repos []models.DockerRepository
	err := db.DB.Find(&repos).Error
	return repos, err
}

func (s *DockerService) GetRepository(name string) (*models.DockerRepository, error) {
	var repo models.DockerRepository
	err := db.DB.Where("name = ?", name).First(&repo).Error
	if err != nil {
		return nil, fmt.Errorf("репозиторий не найден: %w", err)
	}
	return &repo, nil
}

func (s *DockerService) StoreImage(repoName, imageName, tag string, imageData io.Reader) error {
	repo, err := s.GetRepository(repoName)
	if err != nil {
		return err
	}

	if repo.Type != models.TypeHosted {
		return errors.New("нельзя сохранять образы в репозиторий, который не является хостовым")
	}

	imagePath := filepath.Join(repo.StoragePath, imageName, tag)
	if err := os.MkdirAll(imagePath, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории образа: %w", err)
	}

	imageTarPath := filepath.Join(imagePath, "image.tar")
	imageFile, err := os.Create(imageTarPath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла образа: %w", err)
	}
	defer imageFile.Close()

	size, err := io.Copy(imageFile, imageData)
	if err != nil {
		return fmt.Errorf("ошибка записи данных образа: %w", err)
	}

	// TODO: посчитать SHA256 для чека целостности

	imageRecord := models.DockerImage{
		Artifact: models.Artifact{
			RepositoryID:  repo.ID,
			RepoType:      "docker",
			Name:          imageName,
			Version:       tag,
			Path:          imageTarPath,
			Size:          size,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			DownloadCount: 0,
		},
		Tag: tag,
	}

	if err := db.DB.Create(&imageRecord).Error; err != nil {
		os.Remove(imageTarPath)
		return fmt.Errorf("ошибка сохранения записи образа: %w", err)
	}

	log.Printf("Сохранен Docker образ: %s:%s в репозиторий %s", imageName, tag, repoName)
	return nil
}

func (s *DockerService) FetchImageFromProxy(repoName, imageName, tag string) (string, error) {
	repo, err := s.GetRepository(repoName)
	if err != nil {
		return "", err
	}

	if repo.Type != models.TypeProxy {
		return "", errors.New("репозиторий не является прокси")
	}

	imagePath := filepath.Join(repo.StoragePath, imageName, tag)
	imageTarPath := filepath.Join(imagePath, "image.tar")

	if repo.CacheEnabled {
		info, err := os.Stat(imageTarPath)
		if err == nil {
			modTime := info.ModTime()
			cacheDuration := time.Duration(repo.CacheTTL) * time.Minute
			if time.Since(modTime) < cacheDuration {
				log.Printf("Используем кешированный образ: %s:%s", imageName, tag)
				return imageTarPath, nil
			}
		}
	}

	log.Printf("Получение образа %s:%s из удаленного репозитория %s", imageName, tag, repo.URL)

	remoteURL := fmt.Sprintf("%s/v2/%s/manifests/%s", repo.URL, imageName, tag)

	resp, err := http.Get(remoteURL)
	if err != nil {
		return "", fmt.Errorf("ошибка получения манифеста: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка получения манифеста, код ответа: %d", resp.StatusCode)
	}

	var manifest map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return "", fmt.Errorf("ошибка декодирования манифеста: %w", err)
	}

	if err := os.MkdirAll(imagePath, 0755); err != nil {
		return "", fmt.Errorf("ошибка создания директории образа: %w", err)
	}

	// TODO: Скачивание слоев образа и сборка tar-архива
	// 1. Получить список слоев из манифеста
	// 2. Скачать каждый слой
	// 3. Собрать tar-архив образа

	// Заглушка: записываем манифест в файл для примера
	manifestPath := filepath.Join(imagePath, "manifest.json")
	manifestFile, err := os.Create(manifestPath)
	if err != nil {
		return "", fmt.Errorf("ошибка создания файла манифеста: %w", err)
	}
	defer manifestFile.Close()

	if err := json.NewEncoder(manifestFile).Encode(manifest); err != nil {
		return "", fmt.Errorf("ошибка записи манифеста: %w", err)
	}

	imageFile, err := os.Create(imageTarPath)
	if err != nil {
		return "", fmt.Errorf("ошибка создания файла образа: %w", err)
	}
	defer imageFile.Close()

	imageRecord := models.DockerImage{
		Artifact: models.Artifact{
			RepositoryID:  repo.ID,
			RepoType:      "docker",
			Name:          imageName,
			Version:       tag,
			Path:          imageTarPath,
			Size:          0, // TODO: посчитать реальный размер
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			DownloadCount: 0,
		},
		Tag: tag,
	}

	var count int64
	db.DB.Model(&models.DockerImage{}).
		Joins("JOIN artifacts ON docker_images.artifact_id = artifacts.id").
		Where("artifacts.repository_id = ? AND artifacts.name = ? AND docker_images.tag = ?",
			repo.ID, imageName, tag).
		Count(&count)

	if count > 0 {
		db.DB.Model(&models.DockerImage{}).
			Joins("JOIN artifacts ON docker_images.artifact_id = artifacts.id").
			Where("artifacts.repository_id = ? AND artifacts.name = ? AND docker_images.tag = ?",
				repo.ID, imageName, tag).
			Updates(map[string]interface{}{
				"updated_at": time.Now(),
			})
	} else {
		if err := db.DB.Create(&imageRecord).Error; err != nil {
			return "", fmt.Errorf("ошибка сохранения записи образа: %w", err)
		}
	}

	log.Printf("Образ %s:%s получен из удаленного репозитория и сохранен в кеше", imageName, tag)
	return imageTarPath, nil
}

func (s *DockerService) ListImages(repoName string) ([]models.DockerImage, error) {
	repo, err := s.GetRepository(repoName)
	if err != nil {
		return nil, err
	}

	var images []models.DockerImage
	err = db.DB.Joins("JOIN artifacts ON docker_images.artifact_id = artifacts.id").
		Where("artifacts.repository_id = ?", repo.ID).
		Find(&images).Error

	return images, err
}

func (s *DockerService) SearchImages(query string) ([]models.DockerImage, error) {
	var images []models.DockerImage

	if strings.Contains(query, ":") {
		parts := strings.Split(query, ":")
		name, tag := parts[0], parts[1]

		err := db.DB.Joins("JOIN artifacts ON docker_images.artifact_id = artifacts.id").
			Where("artifacts.name LIKE ? AND docker_images.tag LIKE ?",
				"%"+name+"%", "%"+tag+"%").
			Find(&images).Error
		return images, err
	}

	err := db.DB.Joins("JOIN artifacts ON docker_images.artifact_id = artifacts.id").
		Where("artifacts.name LIKE ?", "%"+query+"%").
		Find(&images).Error
	return images, err
}
