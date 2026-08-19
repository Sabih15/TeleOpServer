package playback

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sabih15/TeleOpServer/internal/platform/config"
)

// Storage persists raw frame bytes and returns where they were written.
type Storage interface {
	Save(robotID uint, t time.Time, data []byte) (string, error)
}

// fileStorage writes frames under <baseDir>/<robotID>/<unixNano>.bin.
type fileStorage struct {
	baseDir string
}

func NewStorage(cfg *config.Config) Storage {
	return &fileStorage{baseDir: cfg.Playback.StoragePath}
}

func (s *fileStorage) Save(robotID uint, t time.Time, data []byte) (string, error) {
	dir := filepath.Join(s.baseDir, strconv.FormatUint(uint64(robotID), 10))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create playback storage dir: %w", err)
	}

	path := filepath.Join(dir, fmt.Sprintf("%d.bin", t.UnixNano()))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write playback frame: %w", err)
	}
	return path, nil
}
