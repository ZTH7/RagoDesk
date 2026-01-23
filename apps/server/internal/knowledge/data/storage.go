package data

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/ZTH7/RagoDesk/apps/server/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type objectStorage interface {
	Delete(ctx context.Context, uri string) error
	Get(ctx context.Context, uri string) ([]byte, error)
}

type s3Storage struct {
	client *minio.Client
	bucket string
}

func (s s3Storage) Delete(ctx context.Context, uri string) error {
	if s.client == nil || uri == "" {
		return fmt.Errorf("object storage not initialized")
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
		return fmt.Errorf("unsupported object uri: %s", uri)
	}
	return s.client.RemoveObject(ctx, bucket, object, minio.RemoveObjectOptions{})
}

func (s s3Storage) Get(ctx context.Context, uri string) ([]byte, error) {
	if s.client == nil || uri == "" {
		return nil, fmt.Errorf("object storage not initialized")
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
		return nil, fmt.Errorf("unsupported object uri: %s", uri)
	}
	obj, err := s.client.GetObject(ctx, bucket, object, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

func newObjectStorage(cfg *conf.Data, logger *log.Helper) objectStorage {
	if cfg == nil || cfg.ObjectStorage == nil {
		return nil
	}
	endpoint := strings.TrimSpace(cfg.ObjectStorage.Endpoint)
	if endpoint == "" {
		return nil
	}
	accessKey := strings.TrimSpace(cfg.ObjectStorage.AccessKey)
	secretKey := strings.TrimSpace(cfg.ObjectStorage.SecretKey)
	bucket := strings.TrimSpace(cfg.ObjectStorage.Bucket)
	if accessKey == "" || secretKey == "" || bucket == "" {
		if logger != nil {
			logger.Warn("object storage config missing; disabled")
		}
		return nil
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
		return nil
	}
	return s3Storage{
		client: client,
		bucket: bucket,
	}
}
