// Package storage abstracts S3-compatible object storage (MinIO in dev, AWS S3
// in prod). Binary files (images, cropped panels, evidence, generated reports,
// dataset artifacts) live here; PostgreSQL stores only the storage KEY, never
// the bytes. Business logic depends on the ObjectStorage interface, not the
// MinIO SDK directly.
package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ObjectStorage is the storage boundary business logic depends on. Swapping
// providers (MinIO <-> AWS S3 <-> another S3-compatible store) means
// implementing this interface only; callers never touch the SDK directly.
type ObjectStorage interface {
	Put(ctx context.Context, key string, body []byte, contentType string) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	SignedDownloadURL(ctx context.Context, key string, ttl time.Duration) (string, error)
	SignedUploadURL(ctx context.Context, key string, contentType string, ttl time.Duration) (string, error)
	EnsureBucket(ctx context.Context) error
	// Ping verifies the storage backend is reachable and the configured
	// bucket exists (used by the /ready health check).
	Ping(ctx context.Context) error
}

// Config holds the S3/MinIO connection settings (see .env.example).
type Config struct {
	Endpoint       string
	Region         string
	AccessKey      string
	SecretKey      string
	Bucket         string
	UseSSL         bool
	ForcePathStyle bool
}

type minioStorage struct {
	client *minio.Client
	bucket string
}

// New builds an ObjectStorage backed by a MinIO/S3-compatible client.
func New(cfg Config) (ObjectStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}
	return &minioStorage{client: client, bucket: cfg.Bucket}, nil
}

func (s *minioStorage) Put(ctx context.Context, key string, body []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(body), int64(len(body)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("put object %s: %w", key, err)
	}
	return nil
}

func (s *minioStorage) Get(ctx context.Context, key string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object %s: %w", key, err)
	}
	defer obj.Close()
	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read object %s: %w", key, err)
	}
	return data, nil
}

func (s *minioStorage) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete object %s: %w", key, err)
	}
	return nil
}

func (s *minioStorage) SignedDownloadURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, ttl, nil)
	if err != nil {
		return "", fmt.Errorf("presign download %s: %w", key, err)
	}
	return u.String(), nil
}

func (s *minioStorage) SignedUploadURL(ctx context.Context, key string, contentType string, ttl time.Duration) (string, error) {
	u, err := s.client.PresignedPutObject(ctx, s.bucket, key, ttl)
	if err != nil {
		return "", fmt.Errorf("presign upload %s: %w", key, err)
	}
	return u.String(), nil
}

func (s *minioStorage) Ping(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket %s: %w", s.bucket, err)
	}
	if !exists {
		return fmt.Errorf("bucket %s does not exist", s.bucket)
	}
	return nil
}

func (s *minioStorage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket %s: %w", s.bucket, err)
	}
	if exists {
		return nil
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{Region: ""}); err != nil {
		return fmt.Errorf("create bucket %s: %w", s.bucket, err)
	}
	return nil
}

var unsafeKeyChars = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

// BuildObjectKey constructs a collision-resistant, date-partitioned object key,
// e.g. "inspections/<id>/original/2026-09-01/<uuid>-label.jpg".
func BuildObjectKey(prefix, filename string) string {
	safe := unsafeKeyChars.ReplaceAllString(filename, "_")
	date := time.Now().UTC().Format("2006-01-02")
	return fmt.Sprintf("%s/%s/%s-%s", prefix, date, uuid.NewString(), safe)
}

// SHA256Hex returns the hex-encoded SHA-256 digest of data (used for checksums).
func SHA256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
