package minio

import (
	"context"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const DefaultBucket = "lightchat-files"

func InitMinIO() (*minio.Client, error) {
	endpoint := os.Getenv("LIGHTCHAT_MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "127.0.0.1:9000"
	}

	accessKey := os.Getenv("LIGHTCHAT_MINIO_ACCESS_KEY")
	if accessKey == "" {
		accessKey = "admin"
	}

	secretKey := os.Getenv("LIGHTCHAT_MINIO_SECRET_KEY")
	if secretKey == "" {
		secretKey = "12345678"
	}

	useSSL := false

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	if err := ensureBucket(client, DefaultBucket); err != nil {
		return nil, err
	}

	return client, nil
}

func ensureBucket(client *minio.Client, bucketName string) error {
	ctx := context.Background()

	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
}
