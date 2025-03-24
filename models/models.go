package models

import (
	"time"
)

type RepositoryType string

const (
	TypeHosted RepositoryType = "hosted"
	TypeProxy  RepositoryType = "proxy"
	TypeGroup  RepositoryType = "group"
)

type BaseRepository struct {
	ID          int            `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex"`
	Description string         `json:"description"`
	Type        RepositoryType `json:"type"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type DockerRepository struct {
	BaseRepository
	URL          string `json:"url,omitempty" gorm:"default:null"`
	IndexType    string `json:"index_type,omitempty"`
	CacheEnabled bool   `json:"cache_enabled" gorm:"default:true"`
	CacheTTL     int    `json:"cache_ttl" gorm:"default:1440"`
	StoragePath  string `json:"storage_path"`
}

type GitRepository struct {
	BaseRepository
	URL          string `json:"url,omitempty" gorm:"default:null"`
	Branch       string `json:"branch" gorm:"default:'master'"`
	CloneEnabled bool   `json:"clone_enabled" gorm:"default:true"`
	PushEnabled  bool   `json:"push_enabled" gorm:"default:true"`
	StoragePath  string `json:"storage_path"`
}

type HelmRepository struct {
	BaseRepository
	URL          string `json:"url,omitempty" gorm:"default:null"`
	IndexPath    string `json:"index_path" gorm:"default:'index.yaml'"`
	CacheEnabled bool   `json:"cache_enabled" gorm:"default:true"`
	CacheTTL     int    `json:"cache_ttl" gorm:"default:1440"`
	StoragePath  string `json:"storage_path"`
}

type GroupMember struct {
	ID         int    `json:"id" gorm:"primaryKey"`
	GroupID    int    `json:"group_id"`
	MemberID   int    `json:"member_id"`
	MemberName string `json:"member_name"`
	MemberType string `json:"member_type"`
	Priority   int    `json:"priority"`
}

type Artifact struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	RepositoryID  int       `json:"repository_id"`
	RepoType      string    `json:"repo_type"`
	Name          string    `json:"name"`
	Version       string    `json:"version"`
	Path          string    `json:"path"`
	Size          int64     `json:"size"`
	SHA256        string    `json:"sha256"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DownloadCount int       `json:"download_count" gorm:"default:0"`
}

type DockerImage struct {
	Artifact
	Tag      string   `json:"tag"`
	Manifest []byte   `json:"manifest" gorm:"type:jsonb"`
	Layers   []string `json:"layers" gorm:"type:text[]"`
}

type HelmChart struct {
	Artifact
	AppVersion   string   `json:"app_version"`
	Description  string   `json:"description"`
	Keywords     []string `json:"keywords" gorm:"type:text[]"`
	Dependencies []byte   `json:"dependencies" gorm:"type:jsonb"`
}

type StoredFile struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	ArtifactID int       `json:"artifact_id"`
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"file_path"`
	Size       int64     `json:"size"`
	SHA256     string    `json:"sha256"`
	CreatedAt  time.Time `json:"created_at"`
}
