package infrastructure

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

//go:generate mockgen -source=minio.go -destination=_mock/minio.go
type IMinioClient interface {
	GetObject(ctx context.Context, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
	StatObject(ctx context.Context, objectName string) (*minio.ObjectInfo, error)
}

type MinioClient struct {
	Client *minio.Client
	Bucket string
}

func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool, bucket string) (*MinioClient, error) {
	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio.New error: %w", err)
	}

	return &MinioClient{Client: mc, Bucket: bucket}, nil
}

func (m *MinioClient) GetObject(ctx context.Context, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	return m.Client.GetObject(ctx, m.Bucket, objectName, opts)
}

func (m *MinioClient) StatObject(ctx context.Context, objectName string) (*minio.ObjectInfo, error) {
	objInfo, err := m.Client.StatObject(ctx, m.Bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to stat object: %w", err)
	}
	return &objInfo, nil
}
