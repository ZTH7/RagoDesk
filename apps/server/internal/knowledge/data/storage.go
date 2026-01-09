package data

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type objectStorage interface {
	Delete(ctx context.Context, uri string) error
	Get(ctx context.Context, uri string) ([]byte, error)
}

type noopStorage struct{}

func (noopStorage) Delete(ctx context.Context, uri string) error {
	return nil
}

func (noopStorage) Get(ctx context.Context, uri string) ([]byte, error) {
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return httpGetBytes(ctx, uri)
	}
	return nil, nil
}

type localStorage struct {
	baseDir string
}

func (l localStorage) Delete(ctx context.Context, uri string) error {
	if uri == "" {
		return nil
	}
	path := strings.TrimPrefix(uri, "file://")
	if path == uri {
		path = strings.TrimPrefix(path, "/")
		if l.baseDir != "" && !filepath.IsAbs(path) {
			path = filepath.Join(l.baseDir, filepath.FromSlash(path))
		}
	}
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (l localStorage) Get(ctx context.Context, uri string) ([]byte, error) {
	if uri == "" {
		return nil, nil
	}
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return httpGetBytes(ctx, uri)
	}
	path := strings.TrimPrefix(uri, "file://")
	if path == uri {
		path = strings.TrimPrefix(path, "/")
		if l.baseDir != "" && !filepath.IsAbs(path) {
			path = filepath.Join(l.baseDir, filepath.FromSlash(path))
		}
	}
	if path == "" {
		return nil, nil
	}
	return os.ReadFile(path)
}

type s3Storage struct {
	client *minio.Client
	bucket string
	region string
}

func (s s3Storage) Delete(ctx context.Context, uri string) error {
	if s.client == nil || uri == "" {
		return nil
	}
	bucket := s.bucket
	object := ""
	if strings.HasPrefix(uri, "s3://") {
		if parsed, err := url.Parse(uri); err == nil {
			if parsed.Host != "" {
				bucket = parsed.Host
			}
			object = strings.TrimPrefix(parsed.Path, "/")
		}
	}
	if object == "" {
		return nil
	}
	return s.client.RemoveObject(ctx, bucket, object, minio.RemoveObjectOptions{})
}

func (s s3Storage) Get(ctx context.Context, uri string) ([]byte, error) {
	if s.client == nil || uri == "" {
		return nil, nil
	}
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return httpGetBytes(ctx, uri)
	}
	bucket := s.bucket
	object := ""
	if strings.HasPrefix(uri, "s3://") {
		if parsed, err := url.Parse(uri); err == nil {
			if parsed.Host != "" {
				bucket = parsed.Host
			}
			object = strings.TrimPrefix(parsed.Path, "/")
		}
	}
	if object == "" {
		return nil, nil
	}
	obj, err := s.client.GetObject(ctx, bucket, object, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

func httpGetBytes(ctx context.Context, uri string) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("object get failed: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
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
