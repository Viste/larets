package models

type DockerImage struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Tag       string `json:"tag"`
	CreatedAt string `json:"created_at"`
}

type GitRepository struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
}

type StoredFile struct {
	ID        int    `json:"id" gorm:"primaryKey"`
	FileName  string `json:"file_name"`
	FilePath  string `json:"file_path"`
	CreatedAt string `json:"created_at"`
}

type HelmChart struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	CreatedAt string `json:"created_at"`
}

type MavenArtifact struct {
	ID         int    `json:"id"`
	GroupID    string `json:"group_id"`
	ArtifactID string `json:"artifact_id"`
	Version    string `json:"version"`
	CreatedAt  string `json:"created_at"`
}

type RpmPackage struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	CreatedAt string `json:"created_at"`
}

type DebPackage struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	CreatedAt string `json:"created_at"`
}
