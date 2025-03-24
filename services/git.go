package services

import (
	"errors"
	"fmt"
	"github.com/Viste/larets/config"
	"github.com/Viste/larets/db"
	"github.com/Viste/larets/models"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type GitService struct{}

func (s *GitService) CreateRepository(name, description string, repoType models.RepositoryType, url, branch string) error {
	var count int64
	db.DB.Model(&models.GitRepository{}).Where("name = ?", name).Count(&count)
	if count > 0 {
		return errors.New("репозиторий с таким именем уже существует")
	}

	storagePath := filepath.Join(config.Config.GitStorage, name)
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории хранилища: %w", err)
	}

	repo := models.GitRepository{
		BaseRepository: models.BaseRepository{
			Name:        name,
			Description: description,
			Type:        repoType,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		URL:          url,
		Branch:       branch,
		CloneEnabled: true,
		PushEnabled:  repoType == models.TypeHosted,
		StoragePath:  storagePath,
	}

	if err := db.DB.Create(&repo).Error; err != nil {
		// Очищаем созданную директорию в случае ошибки
		os.RemoveAll(storagePath)
		return fmt.Errorf("ошибка сохранения репозитория: %w", err)
	}

	if repoType == models.TypeHosted {
		cmd := exec.Command("git", "init", "--bare")
		cmd.Dir = storagePath
		if err := cmd.Run(); err != nil {
			db.DB.Delete(&repo)
			os.RemoveAll(storagePath)
			return fmt.Errorf("ошибка инициализации Git репозитория: %w", err)
		}
	} else if repoType == models.TypeProxy && url != "" {
		cmd := exec.Command("git", "clone", "--mirror", url, ".")
		cmd.Dir = storagePath
		if err := cmd.Run(); err != nil {
			db.DB.Delete(&repo)
			os.RemoveAll(storagePath)
			return fmt.Errorf("ошибка клонирования удаленного репозитория: %w", err)
		}
	}

	log.Printf("Создан Git репозиторий: %s, тип: %s", name, repoType)
	return nil
}

func (s *GitService) ListRepositories() ([]models.GitRepository, error) {
	var repos []models.GitRepository
	err := db.DB.Find(&repos).Error
	return repos, err
}

func (s *GitService) GetRepository(name string) (*models.GitRepository, error) {
	var repo models.GitRepository
	err := db.DB.Where("name = ?", name).First(&repo).Error
	if err != nil {
		return nil, fmt.Errorf("репозиторий не найден: %w", err)
	}
	return &repo, nil
}

func (s *GitService) SyncRepository(name string) error {
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

	log.Printf("Синхронизация Git репозитория %s с удаленным источником %s", name, repo.URL)

	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = repo.StoragePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка выполнения git fetch: %w", err)
	}

	repo.UpdatedAt = time.Now()
	if err := db.DB.Save(repo).Error; err != nil {
		return fmt.Errorf("ошибка обновления записи репозитория: %w", err)
	}

	log.Printf("Git репозиторий %s успешно синхронизирован", name)
	return nil
}

func (s *GitService) GetRepoInfo(name string) (map[string]interface{}, error) {
	repo, err := s.GetRepository(name)
	if err != nil {
		return nil, err
	}

	branchesCmd := exec.Command("git", "branch", "--list")
	branchesCmd.Dir = repo.StoragePath
	branchesOutput, err := branchesCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка веток: %w", err)
	}

	logsCmd := exec.Command("git", "log", "--oneline", "-n", "10")
	logsCmd.Dir = repo.StoragePath
	logsOutput, err := logsCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения истории коммитов: %w", err)
	}

	// Собираем информацию
	info := map[string]interface{}{
		"name":        repo.Name,
		"description": repo.Description,
		"type":        repo.Type,
		"url":         repo.URL,
		"branch":      repo.Branch,
		"branches":    string(branchesOutput),
		"recent_logs": string(logsOutput),
		"created_at":  repo.CreatedAt,
		"updated_at":  repo.UpdatedAt,
	}

	return info, nil
}

func (s *GitService) CreateBranch(repoName, branchName, baseBranch string) error {
	repo, err := s.GetRepository(repoName)
	if err != nil {
		return err
	}

	if repo.Type != models.TypeHosted {
		return errors.New("нельзя создавать ветки в репозитории, который не является хостовым")
	}

	cmd := exec.Command("git", "branch", branchName, baseBranch)
	cmd.Dir = repo.StoragePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка создания ветки: %w", err)
	}

	log.Printf("Создана ветка %s в репозитории %s", branchName, repoName)
	return nil
}

func (s *GitService) DeleteBranch(repoName, branchName string) error {
	repo, err := s.GetRepository(repoName)
	if err != nil {
		return err
	}

	if repo.Type != models.TypeHosted {
		return errors.New("нельзя удалять ветки в репозитории, который не является хостовым")
	}

	cmd := exec.Command("git", "branch", "-D", branchName)
	cmd.Dir = repo.StoragePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка удаления ветки: %w", err)
	}

	log.Printf("Удалена ветка %s в репозитории %s", branchName, repoName)
	return nil
}
