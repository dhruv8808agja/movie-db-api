package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var MinioClient *minio.Client
var MinioBucket string

// InitMinIO initializes the MinIO client
func InitMinIO() {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}

	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	if accessKey == "" {
		accessKey = "minioadmin"
	}

	secretKey := os.Getenv("MINIO_SECRET_KEY")
	if secretKey == "" {
		secretKey = "minioadmin123"
	}

	useSSL := false
	if os.Getenv("MINIO_USE_SSL") == "true" {
		useSSL = true
	}

	MinioBucket = os.Getenv("MINIO_BUCKET")
	if MinioBucket == "" {
		MinioBucket = "videos"
	}

	var err error
	MinioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatal("Failed to initialize MinIO client:", err)
	}

	// Check if bucket exists, create if not
	ctx := context.Background()
	exists, err := MinioClient.BucketExists(ctx, MinioBucket)
	if err != nil {
		log.Fatal("Failed to check bucket existence:", err)
	}

	if !exists {
		err = MinioClient.MakeBucket(ctx, MinioBucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatal("Failed to create bucket:", err)
		}
		log.Printf("Bucket '%s' created successfully", MinioBucket)
	}

	log.Printf("MinIO client initialized (endpoint: %s, bucket: %s)", endpoint, MinioBucket)
}

// UploadFile uploads a file to MinIO
func UploadFile(ctx context.Context, objectName string, filePath string, contentType string) error {
	if MinioClient == nil {
		return fmt.Errorf("MinIO client not initialized")
	}

	_, err := MinioClient.FPutObject(ctx, MinioBucket, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})

	return err
}

// UploadStream uploads from an io.Reader to MinIO
func UploadStream(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	if MinioClient == nil {
		return fmt.Errorf("MinIO client not initialized")
	}

	_, err := MinioClient.PutObject(ctx, MinioBucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})

	return err
}

// GetPresignedURL generates a presigned URL for downloading
func GetPresignedURL(ctx context.Context, objectName string, expiry int) (string, error) {
	if MinioClient == nil {
		return "", fmt.Errorf("MinIO client not initialized")
	}

	// Default expiry: 7 days
	if expiry == 0 {
		expiry = 604800
	}

	url, err := MinioClient.PresignedGetObject(ctx, MinioBucket, objectName,
		time.Duration(expiry)*time.Second, nil)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

// DeleteFile deletes a file from MinIO
func DeleteFile(ctx context.Context, objectName string) error {
	if MinioClient == nil {
		return fmt.Errorf("MinIO client not initialized")
	}

	return MinioClient.RemoveObject(ctx, MinioBucket, objectName, minio.RemoveObjectOptions{})
}

// GetFileInfo gets information about an object
func GetFileInfo(ctx context.Context, objectName string) (*minio.ObjectInfo, error) {
	if MinioClient == nil {
		return nil, fmt.Errorf("MinIO client not initialized")
	}

	info, err := MinioClient.StatObject(ctx, MinioBucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// GetUploadChunkSize returns the configured chunk size for uploads
func GetUploadChunkSize() int64 {
	chunkSizeStr := os.Getenv("UPLOAD_CHUNK_SIZE")
	if chunkSizeStr == "" {
		return 5 * 1024 * 1024 // Default: 5MB
	}

	chunkSize, err := strconv.ParseInt(chunkSizeStr, 10, 64)
	if err != nil {
		return 5 * 1024 * 1024
	}

	return chunkSize
}

// GetMaxVideoSize returns the maximum allowed video size
func GetMaxVideoSize() int64 {
	maxSizeStr := os.Getenv("MAX_VIDEO_SIZE")
	if maxSizeStr == "" {
		return 10 * 1024 * 1024 * 1024 // Default: 10GB
	}

	maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64)
	if err != nil {
		return 10 * 1024 * 1024 * 1024
	}

	return maxSize
}
