package data

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
)

type objectStorage interface {
	Put(ctx context.Context, key string, content []byte) (string, error)
}

type noopStorage struct{}

func (noopStorage) Put(ctx context.Context, key string, content []byte) (string, error) {
	return "", nil
}

type localStorage struct {
	baseDir string
}

func (l localStorage) Put(ctx context.Context, key string, content []byte) (string, error) {
	if l.baseDir == "" {
		return "", nil
	}
	safeKey := filepath.FromSlash(strings.TrimPrefix(key, "/"))
	path := filepath.Join(l.baseDir, safeKey)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", err
	}
	return "file://" + path, nil
}

func newObjectStorage(cfg *conf.Data, logger *log.Helper) objectStorage {
	if cfg == nil || cfg.ObjectStorage == nil {
		return noopStorage{}
	}
	endpoint := strings.TrimSpace(cfg.ObjectStorage.Endpoint)
	if endpoint == "" {
		return noopStorage{}
	}
	if strings.HasPrefix(endpoint, "file://") {
		baseDir := strings.TrimPrefix(endpoint, "file://")
		if baseDir == "" {
			return noopStorage{}
		}
		return localStorage{baseDir: baseDir}
	}
	if looksLikePath(endpoint) {
		return localStorage{baseDir: endpoint}
	}
	if logger != nil {
		logger.Warnf("object storage endpoint %s not supported in MVP; fallback to noop", endpoint)
	}
	return noopStorage{}
}

func looksLikePath(value string) bool {
	if value == "" {
		return false
	}
	if strings.HasPrefix(value, "/") || strings.HasPrefix(value, ".") {
		return true
	}
	if len(value) >= 2 && value[1] == ':' {
		return true
	}
	return false
}
