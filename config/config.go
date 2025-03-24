package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var Config struct {
	EnableDocker bool
	EnableGit    bool
	EnableHelm   bool
	ServerPort   string
	BaseURL      string

	StorageBasePath string
	DockerStorage   string
	GitStorage      string
	HelmStorage     string
	TempStorage     string

	DefaultCacheTTL int

	EnableAuth    bool
	AdminUser     string
	AdminPassword string // переписать с plaintext
}

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("Нет .env файла, используем только переменные окружения")
	}

	Config.EnableDocker = getEnvBool("ENABLE_DOCKER", true)
	Config.EnableGit = getEnvBool("ENABLE_GIT", true)
	Config.EnableHelm = getEnvBool("ENABLE_HELM", true)

	Config.ServerPort = getEnv("SERVER_PORT", "8080")
	Config.BaseURL = getEnv("BASE_URL", "http://localhost:"+Config.ServerPort)

	Config.StorageBasePath = getEnv("STORAGE_PATH", "./storage")
	Config.DockerStorage = filepath.Join(Config.StorageBasePath, "docker")
	Config.GitStorage = filepath.Join(Config.StorageBasePath, "git")
	Config.HelmStorage = filepath.Join(Config.StorageBasePath, "helm")
	Config.TempStorage = filepath.Join(Config.StorageBasePath, "temp")

	Config.DefaultCacheTTL = getEnvInt("DEFAULT_CACHE_TTL", 1440) // 24 часа в минутах

	Config.EnableAuth = getEnvBool("ENABLE_AUTH", false)
	Config.AdminUser = getEnv("ADMIN_USER", "admin")
	Config.AdminPassword = getEnv("ADMIN_PASSWORD", "admin") // Не рекомендуется в production

	log.Println("Конфигурация загружена успешно")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	value = strings.ToLower(value)
	return value == "true" || value == "1" || value == "yes" || value == "y"
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		log.Printf("Ошибка при парсинге значения %s: %v, используем значение по умолчанию %d", key, err, defaultValue)
		return defaultValue
	}

	return result
}
