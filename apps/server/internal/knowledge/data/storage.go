package data

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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

type s3Storage struct {
	client *minio.Client
	bucket string
	region string
}

func (s s3Storage) Put(ctx context.Context, key string, content []byte) (string, error) {
	object := strings.TrimPrefix(key, "/")
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return "", err
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{Region: s.region}); err != nil {
			return "", err
		}
	}
	reader := bytes.NewReader(content)
	_, err = s.client.PutObject(ctx, s.bucket, object, reader, int64(len(content)), minio.PutObjectOptions{
		ContentType: "text/plain",
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("s3://%s/%s", s.bucket, object), nil
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
	accessKey := strings.TrimSpace(cfg.ObjectStorage.AccessKey)
	secretKey := strings.TrimSpace(cfg.ObjectStorage.SecretKey)
	bucket := strings.TrimSpace(cfg.ObjectStorage.Bucket)
	if accessKey == "" || secretKey == "" || bucket == "" {
		if logger != nil {
			logger.Warn("object storage config missing; fallback to noop")
		}
		return noopStorage{}
	}
	secure := cfg.ObjectStorage.UseSsl
	if parsed, err := url.Parse(endpoint); err == nil && parsed.Host != "" {
		if parsed.Scheme == "https" {
			secure = true
		} else if parsed.Scheme == "http" {
			secure = false
		}
		endpoint = parsed.Host
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
		Region: strings.TrimSpace(cfg.ObjectStorage.Region),
	})
	if err != nil {
		if logger != nil {
			logger.Warnf("object storage init failed: %v", err)
		}
		return noopStorage{}
	}
	return s3Storage{
		client: client,
		bucket: bucket,
		region: strings.TrimSpace(cfg.ObjectStorage.Region),
	}
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
