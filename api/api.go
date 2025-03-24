package api

import (
	"encoding/json"
	"fmt"
	"github.com/Viste/larets/config"
	"github.com/Viste/larets/models"
	"github.com/Viste/larets/services"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	dockerService = &services.DockerService{}
	gitService    = &services.GitService{}
	helmService   = &services.HelmService{}
)

func RunAPIServer() {
	log.Printf("Запуск API сервера на порту :%s", config.Config.ServerPort)
	http.HandleFunc("/api/health", handleHealth)

	if config.Config.EnableDocker {
		http.HandleFunc("/api/docker/repositories", handleDockerRepositories)
		http.HandleFunc("/api/docker/repositories/", handleDockerRepositoryByName)
		http.HandleFunc("/api/docker/images", handleDockerImages)
		http.HandleFunc("/v2/", handleDockerRegistryAPI)
	}

	if config.Config.EnableGit {
		http.HandleFunc("/api/git/repositories", handleGitRepositories)
		http.HandleFunc("/api/git/repositories/", handleGitRepositoryByName)
		http.HandleFunc("/api/git/sync/", handleGitSync)
		http.HandleFunc("/git/", handleGitProtocol)
	}

	if config.Config.EnableHelm {
		http.HandleFunc("/api/helm/repositories", handleHelmRepositories)
		http.HandleFunc("/api/helm/repositories/", handleHelmRepositoryByName)
		http.HandleFunc("/api/helm/charts", handleHelmCharts)
		http.HandleFunc("/api/helm/sync/", handleHelmSync)
		http.HandleFunc("/helm/", handleHelmAccess)
	}

	listenAddr := fmt.Sprintf(":%s", config.Config.ServerPort)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":  "ok",
		"version": "0.1.0",
		"features": map[string]bool{
			"docker": config.Config.EnableDocker,
			"git":    config.Config.EnableGit,
			"helm":   config.Config.EnableHelm,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Docker API Handlers
func handleDockerRepositories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		repos, err := dockerService.ListRepositories()
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка получения списка репозиториев: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repos)

	case http.MethodPost:
		var request struct {
			Name        string                `json:"name"`
			Description string                `json:"description"`
			Type        models.RepositoryType `json:"type"`
			URL         string                `json:"url,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
			return
		}

		err := dockerService.CreateRepository(request.Name, request.Description, request.Type, request.URL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка создания репозитория: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		response := map[string]string{"message": "Репозиторий успешно создан"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func handleDockerRepositoryByName(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	repoName := pathParts[3]

	switch r.Method {
	case http.MethodGet:
		repo, err := dockerService.GetRepository(repoName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка получения информации о репозитории: %v", err), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repo)

	case http.MethodDelete:
		// TODO: удаление репозитория
		http.Error(w, "Метод пока не реализован", http.StatusNotImplemented)

	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func handleDockerImages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		query := r.URL.Query()
		repoName := query.Get("repository")
		searchQuery := query.Get("q")

		if searchQuery != "" {
			images, err := dockerService.SearchImages(searchQuery)
			if err != nil {
				http.Error(w, fmt.Sprintf("Ошибка поиска образов: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(images)
			return
		}

		if repoName != "" {
			images, err := dockerService.ListImages(repoName)
			if err != nil {
				http.Error(w, fmt.Sprintf("Ошибка получения списка образов: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(images)
			return
		}

		http.Error(w, "Необходимо указать параметр repository или q", http.StatusBadRequest)

	case http.MethodPost:
		repoName := r.URL.Query().Get("repository")
		imageName := r.URL.Query().Get("name")
		tag := r.URL.Query().Get("tag")

		if repoName == "" || imageName == "" || tag == "" {
			http.Error(w, "Необходимо указать параметры repository, name и tag", http.StatusBadRequest)
			return
		}

		err := dockerService.StoreImage(repoName, imageName, tag, r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка сохранения образа: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		response := map[string]string{"message": "Образ успешно сохранен"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func handleDockerRegistryAPI(w http.ResponseWriter, r *http.Request) {
	// TODO: Docker Registry API v2
	// аутентификация, манифесты, слои и.т.д
	http.Error(w, "Docker Registry API v2 пока не реализован", http.StatusNotImplemented)
}

// Git API Handlers
func handleGitRepositories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		repos, err := gitService.ListRepositories()
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка получения списка репозиториев: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repos)

	case http.MethodPost:
		var request struct {
			Name        string                `json:"name"`
			Description string                `json:"description"`
			Type        models.RepositoryType `json:"type"`
			URL         string                `json:"url,omitempty"`
			Branch      string                `json:"branch,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
			return
		}

		if request.Branch == "" {
			request.Branch = "master"
		}

		err := gitService.CreateRepository(request.Name, request.Description, request.Type, request.URL, request.Branch)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка создания репозитория: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		response := map[string]string{"message": "Репозиторий успешно создан"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func handleGitRepositoryByName(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	repoName := pathParts[3]

	switch r.Method {
	case http.MethodGet:
		repoInfo, err := gitService.GetRepoInfo(repoName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка получения информации о репозитории: %v", err), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repoInfo)

	case http.MethodDelete:
		// TODO: удаление репозитория
		http.Error(w, "Метод пока не реализован", http.StatusNotImplemented)

	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func handleGitSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	repoName := pathParts[3]

	err := gitService.SyncRepository(repoName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка синхронизации репозитория: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "Репозиторий успешно синхронизирован"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGitProtocol(w http.ResponseWriter, r *http.Request) {
	// TODO: Git HTTP protocol
	// обработка запросов git-upload-pack, git-receive-pack и.т.д
	http.Error(w, "Git HTTP protocol пока не реализован", http.StatusNotImplemented)
}

// Helm API Handlers
func handleHelmRepositories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		repos, err := helmService.ListRepositories()
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка получения списка репозиториев: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repos)

	case http.MethodPost:
		var request struct {
			Name        string                `json:"name"`
			Description string                `json:"description"`
			Type        models.RepositoryType `json:"type"`
			URL         string                `json:"url,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
			return
		}

		err := helmService.CreateRepository(request.Name, request.Description, request.Type, request.URL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка создания репозитория: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		response := map[string]string{"message": "Репозиторий успешно создан"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func handleHelmRepositoryByName(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	repoName := pathParts[3]

	switch r.Method {
	case http.MethodGet:
		repo, err := helmService.GetRepository(repoName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка получения информации о репозитории: %v", err), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repo)

	case http.MethodDelete:
		// TODO: удаление репозитория
		http.Error(w, "Метод пока не реализован", http.StatusNotImplemented)

	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func handleHelmCharts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Получаем параметры запроса
		repoName := r.URL.Query().Get("repository")

		if repoName == "" {
			http.Error(w, "Необходимо указать параметр repository", http.StatusBadRequest)
			return
		}

		charts, err := helmService.ListCharts(repoName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка получения списка чартов: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(charts)

	case http.MethodPost:
		repoName := r.URL.Query().Get("repository")
		filename := r.URL.Query().Get("filename")

		if repoName == "" || filename == "" {
			http.Error(w, "Необходимо указать параметры repository и filename", http.StatusBadRequest)
			return
		}

		if !strings.HasSuffix(filename, ".tgz") {
			http.Error(w, "Файл должен иметь расширение .tgz", http.StatusBadRequest)
			return
		}

		err := helmService.UploadChart(repoName, r.Body, filename)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка загрузки чарта: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		response := map[string]string{"message": "Чарт успешно загружен"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func handleHelmSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	repoName := pathParts[3]

	err := helmService.SyncRepository(repoName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка синхронизации репозитория: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "Репозиторий успешно синхронизирован"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleHelmAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Неверный URL", http.StatusBadRequest)
		return
	}

	repoName := pathParts[2]
	repo, err := helmService.GetRepository(repoName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Репозиторий не найден: %v", err), http.StatusNotFound)
		return
	}

	if len(pathParts) == 4 && pathParts[3] == "index.yaml" {
		indexPath := filepath.Join(repo.StoragePath, repo.IndexPath)
		http.ServeFile(w, r, indexPath)
		return
	}

	if len(pathParts) >= 5 && pathParts[3] == "charts" {
		chartFileName := pathParts[4]

		chartPath := filepath.Join(repo.StoragePath, "charts", chartFileName)
		if _, err := os.Stat(chartPath); os.IsNotExist(err) {
			if repo.Type == models.TypeProxy {
				chartNameVersion := strings.TrimSuffix(chartFileName, ".tgz")
				parts := strings.Split(chartNameVersion, "-")
				if len(parts) < 2 {
					http.Error(w, "Неверное имя файла чарта", http.StatusBadRequest)
					return
				}

				version := parts[len(parts)-1]
				name := strings.Join(parts[:len(parts)-1], "-")

				chartPath, err = helmService.FetchChartFromProxy(repoName, name, version)
				if err != nil {
					http.Error(w, fmt.Sprintf("Ошибка получения чарта: %v", err), http.StatusInternalServerError)
					return
				}
			} else {
				http.Error(w, "Чарт не найден", http.StatusNotFound)
				return
			}
		}

		http.ServeFile(w, r, chartPath)
		return
	}

	http.Error(w, "Неверный путь", http.StatusBadRequest)
}
