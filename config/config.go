package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

var Config map[string]string

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("Нет .env файла, пропускаем")
	}

	Config = make(map[string]string)
	Config["ENABLE_DOCKER"] = os.Getenv("ENABLE_DOCKER")
	Config["ENABLE_GIT"] = os.Getenv("ENABLE_GIT")
	Config["ENABLE_HELM"] = os.Getenv("ENABLE_HELM")
	Config["ENABLE_MAVEN"] = os.Getenv("ENABLE_MAVEN")
	Config["ENABLE_RPM"] = os.Getenv("ENABLE_RPM")
	Config["ENABLE_DEB"] = os.Getenv("ENABLE_DEB")
}
