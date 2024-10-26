package models

// DockerImage модель данных для Docker образов
type DockerImage struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Tag       string `json:"tag"`
	CreatedAt string `json:"created_at"`
}

// GitRepository модель данных для Git репозиториев
type GitRepository struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
}

// HelmChart модель данных для Helm чартов
type HelmChart struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	CreatedAt string `json:"created_at"`
}

// MavenArtifact модель данных для Maven артефактов
type MavenArtifact struct {
	ID         int    `json:"id"`
	GroupID    string `json:"group_id"`
	ArtifactID string `json:"artifact_id"`
	Version    string `json:"version"`
	CreatedAt  string `json:"created_at"`
}

// RpmPackage модель данных для RPM пакетов
type RpmPackage struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	CreatedAt string `json:"created_at"`
}

// DebPackage модель данных для DEB пакетов
type DebPackage struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	CreatedAt string `json:"created_at"`
}
