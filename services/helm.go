package services

import (
	"errors"
	"fmt"
	"github.com/Viste/larets/config"
	"github.com/Viste/larets/db"
	"github.com/Viste/larets/models"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type HelmService struct{}

func (s *HelmService) CreateRepository(name, description string, repoType models.RepositoryType, url string) error {
	var count int64
	db.DB.Model(&models.HelmRepository{}).Where("name = ?", name).Count(&count)
	if count > 0 {
		return errors.New("репозиторий с таким именем уже существует")
	}

	storagePath := filepath.Join(config.Config.HelmStorage, name)
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории хранилища: %w", err)
	}

	repo := models.HelmRepository{
		BaseRepository: models.BaseRepository{
			Name:        name,
			Description: description,
			Type:        repoType,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		URL:          url,
		IndexPath:    "index.yaml",
		CacheEnabled: true,
		CacheTTL:     config.Config.DefaultCacheTTL,
		StoragePath:  storagePath,
	}

	if err := db.DB.Create(&repo).Error; err != nil {
		os.RemoveAll(storagePath)
		return fmt.Errorf("ошибка сохранения репозитория: %w", err)
	}

	if repoType == models.TypeHosted {
		indexPath := filepath.Join(storagePath, "index.yaml")
		// пустой index.yaml с базовой структурой
		indexContent := `apiVersion: v1
entries: {}
generated: "%s"
`
		// RFC3339
		indexContent = fmt.Sprintf(indexContent, time.Now().Format(time.RFC3339))

		if err := ioutil.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
			db.DB.Delete(&repo)
			os.RemoveAll(storagePath)
			return fmt.Errorf("ошибка создания индексного файла: %w", err)
		}
	} else if repoType == models.TypeProxy && url != "" {
		indexURL := fmt.Sprintf("%s/index.yaml", url)
		resp, err := http.Get(indexURL)
		if err != nil {
			db.DB.Delete(&repo)
			os.RemoveAll(storagePath)
			return fmt.Errorf("ошибка получения индексного файла: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			db.DB.Delete(&repo)
			os.RemoveAll(storagePath)
			return fmt.Errorf("ошибка получения индексного файла, код ответа: %d", resp.StatusCode)
		}

		indexPath := filepath.Join(storagePath, "index.yaml")
		indexFile, err := os.Create(indexPath)
		if err != nil {
			db.DB.Delete(&repo)
			os.RemoveAll(storagePath)
			return fmt.Errorf("ошибка создания индексного файла: %w", err)
		}
		defer indexFile.Close()

		if _, err := io.Copy(indexFile, resp.Body); err != nil {
			db.DB.Delete(&repo)
			os.RemoveAll(storagePath)
			return fmt.Errorf("ошибка записи индексного файла: %w", err)
		}
	}

	log.Printf("Создан Helm репозиторий: %s, тип: %s", name, repoType)
	return nil
}

func (s *HelmService) ListRepositories() ([]models.HelmRepository, error) {
	var repos []models.HelmRepository
	err := db.DB.Find(&repos).Error
	return repos, err
}

func (s *HelmService) GetRepository(name string) (*models.HelmRepository, error) {
	var repo models.HelmRepository
	err := db.DB.Where("name = ?", name).First(&repo).Error
	if err != nil {
		return nil, fmt.Errorf("репозиторий не найден: %w", err)
	}
	return &repo, nil
}

func (s *HelmService) UploadChart(repoName string, chartData io.Reader, filename string) error {
	repo, err := s.GetRepository(repoName)
	if err != nil {
		return err
	}

	if repo.Type != models.TypeHosted {
		return errors.New("нельзя загружать чарты в репозиторий, который не является хостовым")
	}

	tempDir, err := ioutil.TempDir(config.Config.TempStorage, "helm-chart-")
	if err != nil {
		return fmt.Errorf("ошибка создания временной директории: %w", err)
	}
	defer os.RemoveAll(tempDir)

	tempChartPath := filepath.Join(tempDir, filename)
	tempChartFile, err := os.Create(tempChartPath)
	if err != nil {
		return fmt.Errorf("ошибка создания временного файла: %w", err)
	}
	defer tempChartFile.Close()

	size, err := io.Copy(tempChartFile, chartData)
	if err != nil {
		return fmt.Errorf("ошибка записи данных чарта: %w", err)
	}

	chartsDir := filepath.Join(repo.StoragePath, "charts")
	if err := os.MkdirAll(chartsDir, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории чартов: %w", err)
	}

	chartPath := filepath.Join(chartsDir, filename)
	if err := copyFile(tempChartPath, chartPath); err != nil {
		return fmt.Errorf("ошибка копирования чарта в хранилище: %w", err)
	}

	indexPath := filepath.Join(repo.StoragePath, repo.IndexPath)
	cmd := exec.Command("helm", "repo", "index", chartsDir, "--url", config.Config.BaseURL+"/helm/"+repoName+"/charts", "--merge", indexPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка обновления индекса репозитория: %w", err)
	}

	// example получаем информацию о чарте
	// для этого нужно распаковать архив и прочитать Chart.yaml
	chartName := filepath.Base(filename)
	chartName = chartName[:len(chartName)-len(filepath.Ext(chartName))]
	chartVersion := "0.1.0" // todo извлечь из Chart.yaml

	chartRecord := models.HelmChart{
		Artifact: models.Artifact{
			RepositoryID:  repo.ID,
			RepoType:      "helm",
			Name:          chartName,
			Version:       chartVersion,
			Path:          chartPath,
			Size:          size,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			DownloadCount: 0,
		},
		AppVersion:  "1.0.0",          // todo извлечь из Chart.yaml
		Description: "Uploaded chart", // todo извлечь из Chart.yaml
	}

	if err := db.DB.Create(&chartRecord).Error; err != nil {
		return fmt.Errorf("ошибка сохранения записи чарта: %w", err)
	}

	log.Printf("Загружен Helm чарт: %s в репозиторий %s", filename, repoName)
	return nil
}

func (s *HelmService) SyncRepository(name string) error {
	repo, err := s.GetRepository(name)
	if err != nil {
		return err
	}

	if repo.Type != models.TypeProxy {
		return errors.New("репозиторий не является прокси")
	}

	if repo.URL == "" {
		return errors.New("URL удаленного репозитория не указан")
	}

	log.Printf("Синхронизация Helm репозитория %s с удаленным источником %s", name, repo.URL)

	indexURL := fmt.Sprintf("%s/index.yaml", repo.URL)
	resp, err := http.Get(indexURL)
	if err != nil {
		return fmt.Errorf("ошибка получения индексного файла: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка получения индексного файла, код ответа: %d", resp.StatusCode)
	}

	indexPath := filepath.Join(repo.StoragePath, repo.IndexPath)
	indexFile, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("ошибка создания индексного файла: %w", err)
	}
	defer indexFile.Close()

	if _, err := io.Copy(indexFile, resp.Body); err != nil {
		return fmt.Errorf("ошибка записи индексного файла: %w", err)
	}

	repo.UpdatedAt = time.Now()
	if err := db.DB.Save(repo).Error; err != nil {
		return fmt.Errorf("ошибка обновления записи репозитория: %w", err)
	}

	log.Printf("Helm репозиторий %s успешно синхронизирован", name)
	return nil
}

func (s *HelmService) FetchChartFromProxy(repoName, chartName, version string) (string, error) {
	repo, err := s.GetRepository(repoName)
	if err != nil {
		return "", err
	}

	if repo.Type != models.TypeProxy {
		return "", errors.New("репозиторий не является прокси")
	}

	chartsDir := filepath.Join(repo.StoragePath, "charts")
	if err := os.MkdirAll(chartsDir, 0755); err != nil {
		return "", fmt.Errorf("ошибка создания директории чартов: %w", err)
	}

	chartFileName := fmt.Sprintf("%s-%s.tgz", chartName, version)
	chartPath := filepath.Join(chartsDir, chartFileName)

	if repo.CacheEnabled {
		info, err := os.Stat(chartPath)
		if err == nil {
			modTime := info.ModTime()
			cacheDuration := time.Duration(repo.CacheTTL) * time.Minute
			if time.Since(modTime) < cacheDuration {
				log.Printf("Используем кешированный чарт: %s-%s", chartName, version)
				return chartPath, nil
			}
		}
	}

	log.Printf("Получение чарта %s-%s из удаленного репозитория %s", chartName, version, repo.URL)

	chartURL := fmt.Sprintf("%s/charts/%s", repo.URL, chartFileName)
	resp, err := http.Get(chartURL)
	if err != nil {
		return "", fmt.Errorf("ошибка получения чарта: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка получения чарта, код ответа: %d", resp.StatusCode)
	}

	chartFile, err := os.Create(chartPath)
	if err != nil {
		return "", fmt.Errorf("ошибка создания файла чарта: %w", err)
	}
	defer chartFile.Close()

	size, err := io.Copy(chartFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка записи данных чарта: %w", err)
	}

	chartRecord := models.HelmChart{
		Artifact: models.Artifact{
			RepositoryID:  repo.ID,
			RepoType:      "helm",
			Name:          chartName,
			Version:       version,
			Path:          chartPath,
			Size:          size,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			DownloadCount: 0,
		},
		AppVersion:  "unknown",     // todo извлечь из Chart.yaml
		Description: "Proxy chart", // todo Chart.yaml
	}

	var count int64
	db.DB.Model(&models.HelmChart{}).
		Joins("JOIN artifacts ON helm_charts.artifact_id = artifacts.id").
		Where("artifacts.repository_id = ? AND artifacts.name = ? AND artifacts.version = ?",
			repo.ID, chartName, version).
		Count(&count)

	if count > 0 {
		db.DB.Model(&models.HelmChart{}).
			Joins("JOIN artifacts ON helm_charts.artifact_id = artifacts.id").
			Where("artifacts.repository_id = ? AND artifacts.name = ? AND artifacts.version = ?",
				repo.ID, chartName, version).
			Updates(map[string]interface{}{
				"updated_at": time.Now(),
			})
	} else {
		if err := db.DB.Create(&chartRecord).Error; err != nil {
			return "", fmt.Errorf("ошибка сохранения записи чарта: %w", err)
		}
	}

	log.Printf("Чарт %s-%s получен из удаленного репозитория и сохранен в кеше", chartName, version)
	return chartPath, nil
}

func (s *HelmService) ListCharts(repoName string) ([]models.HelmChart, error) {
	repo, err := s.GetRepository(repoName)
	if err != nil {
		return nil, err
	}

	var charts []models.HelmChart
	err = db.DB.Joins("JOIN artifacts ON helm_charts.artifact_id = artifacts.id").
		Where("artifacts.repository_id = ?", repo.ID).
		Find(&charts).Error

	return charts, err
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
