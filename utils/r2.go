package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	r2Client *minio.Client
	r2Bucket string
)

func InitR2() {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	if accountID == "" {
		log.Println("R2_ACCOUNT_ID not set")
		return
	}
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucket := os.Getenv("R2_BUCKET")
	if bucket == "" {
		bucket = "img"
	}

	endpoint := fmt.Sprintf("%s.r2.cloudflarestorage.com", accountID)

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Printf("R2 init error: %v", err)
		return
	}

	r2Client = client
	r2Bucket = bucket
	log.Printf("R2 ready endpoint=%s bucket=%s", endpoint, bucket)
}

func R2Enabled() bool {
	return r2Client != nil
}

func R2PresignedGetURL(name string) (string, error) {
	u, err := r2Client.PresignedGetObject(
		context.Background(),
		r2Bucket,
		name,
		15*time.Minute,
		url.Values{},
	)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func R2GetObject(name string) ([]byte, error) {
	obj, err := r2Client.GetObject(
		context.Background(),
		r2Bucket,
		name,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

func R2PutObject(name string, r io.Reader, size int64, contentType string) error {
	_, err := r2Client.PutObject(
		context.Background(),
		r2Bucket,
		name,
		r,
		size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	return err
}

func R2Remove(name string) error {
	return r2Client.RemoveObject(
		context.Background(),
		r2Bucket,
		name,
		minio.RemoveObjectOptions{},
	)
}

func R2PutAt(key string, r io.Reader, size int64, contentType string) error {
	_, err := r2Client.PutObject(
		context.Background(),
		r2Bucket,
		key,
		r,
		size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	return err
}
