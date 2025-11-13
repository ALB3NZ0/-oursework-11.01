package models

import "time"

// BackupResponse представляет ответ при создании бэкапа
type BackupResponse struct {
	Message  string `json:"message"`
	Success  bool   `json:"success"`
	FilePath string `json:"file_path"`
}

// BackupInfo представляет информацию о файле бэкапа
type BackupInfo struct {
	Filename   string    `json:"filename"`
	Path       string    `json:"path"`
	SizeBytes  int64     `json:"size_bytes"`
	SizeMB     float64   `json:"size_mb"`
	Created    time.Time `json:"created"`
}

// BackupListResponse представляет ответ со списком бэкапов
type BackupListResponse struct {
	BackupFiles []BackupInfo `json:"backup_files"`
	TotalFiles  int          `json:"total_files"`
	DesktopPath string       `json:"desktop_path"`
}

// BackupDeleteResponse представляет ответ при удалении бэкапа
type BackupDeleteResponse struct {
	Message string `json:"message"`
}































