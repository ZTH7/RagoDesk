package data

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	biz "github.com/ZTH7/RAGDesk/apps/server/internal/apimgmt/biz"
	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const exportURLTTL = 15 * time.Minute

type usageExporter struct {
	client *minio.Client
	bucket string
	log    *log.Helper
}

// NewUsageExporter creates an exporter backed by object storage.
func NewUsageExporter(cfg *conf.Data, logger log.Logger) biz.UsageExporter {
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
		log.NewHelper(logger).Warn("object storage config missing; usage export disabled")
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
		log.NewHelper(logger).Warnf("object storage init failed for usage export: %v", err)
		return nil
	}
	return &usageExporter{
		client: client,
		bucket: bucket,
		log:    log.NewHelper(logger),
	}
}

func (e *usageExporter) ExportUsageCSV(ctx context.Context, tenantID string, filename string, reader io.Reader, contentType string) (string, string, error) {
	if e == nil || e.client == nil || e.bucket == "" {
		return "", "", fmt.Errorf("object storage not configured")
	}
	if reader == nil {
		return "", "", fmt.Errorf("export reader missing")
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "text/csv"
	}
	if strings.TrimSpace(tenantID) == "" {
		tenantID = "unknown"
	}
	timestamp := time.Now().UTC().Format("20060102_150405")
	baseName := strings.TrimSpace(filename)
	if baseName == "" {
		baseName = fmt.Sprintf("api_usage_%s.csv", timestamp)
	}
	baseName = strings.ReplaceAll(baseName, " ", "_")
	objectName := fmt.Sprintf("api-usage/%s/%s_%s_%s", tenantID, timestamp, uuid.NewString(), baseName)
	if !strings.HasSuffix(strings.ToLower(objectName), ".csv") {
		objectName += ".csv"
	}
	_, err := e.client.PutObject(ctx, e.bucket, objectName, reader, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", "", err
	}
	objectURI := fmt.Sprintf("s3://%s/%s", e.bucket, objectName)
	downloadURL := ""
	if presigned, err := e.client.PresignedGetObject(ctx, e.bucket, objectName, exportURLTTL, nil); err == nil {
		downloadURL = presigned.String()
	} else if e.log != nil {
		e.log.Warnf("usage export presign failed: %v", err)
	}
	return objectURI, downloadURL, nil
}
