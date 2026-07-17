package file

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	pkgminio "chatapp-backend/pkg/minio"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

var (
	ErrInvalidFileType = errors.New("不支持的文件类型")
	ErrFileTooLarge    = errors.New("文件过大")
)

type Service struct {
	db          *gorm.DB
	minioClient *minio.Client
}

func NewService(db *gorm.DB, minioClient *minio.Client) *Service {
	return &Service{
		db:          db,
		minioClient: minioClient,
	}
}

func (s *Service) Upload(ctx context.Context, uploaderID uint64, fileType string, header *multipart.FileHeader) (*UploadFileResponse, error) {
	if err := validateFile(fileType, header); err != nil {
		return nil, err
	}

	src, err := header.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	objectName := buildObjectName(uploaderID, fileType, header.Filename)
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = s.minioClient.PutObject(
		ctx,
		pkgminio.DefaultBucket,
		objectName,
		src,
		header.Size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return nil, err
	}

	record := FileRecord{
		UploaderID:   uploaderID,
		BucketName:   pkgminio.DefaultBucket,
		ObjectName:   objectName,
		OriginalName: header.Filename,
		FileType:     fileType,
		MimeType:     contentType,
		Size:         header.Size,
	}

	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, err
	}

	fileURL, err := s.GenerateURL(ctx, objectName)
	if err != nil {
		return nil, err
	}

	return &UploadFileResponse{
		ID:         record.ID,
		FileType:   record.FileType,
		FileName:   record.OriginalName,
		ObjectName: record.ObjectName,
		MimeType:   record.MimeType,
		Size:       record.Size,
		URL:        fileURL,
		CreateTime: record.CreateTime,
	}, nil
}

func (s *Service) GenerateURL(ctx context.Context, objectName string) (string, error) {
	reqParams := make(url.Values)

	presignedURL, err := s.minioClient.PresignedGetObject(
		ctx,
		pkgminio.DefaultBucket,
		objectName,
		24*time.Hour,
		reqParams,
	)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}

func validateFile(fileType string, header *multipart.FileHeader) error {
	switch fileType {
	case FileTypeImage:
		if header.Size > 10*1024*1024 {
			return ErrFileTooLarge
		}

		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
			return ErrInvalidFileType
		}

	case FileTypeVoice:
		if header.Size > 20*1024*1024 {
			return ErrFileTooLarge
		}

		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".mp3" && ext != ".wav" && ext != ".ogg" && ext != ".m4a" && ext != ".webm" {
			return ErrInvalidFileType
		}

	case FileTypeFile:
		if header.Size > 100*1024*1024 {
			return ErrFileTooLarge
		}

	default:
		return ErrInvalidFileType
	}

	return nil
}

func buildObjectName(userID uint64, fileType string, filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	date := time.Now().Format("20060102")
	id := uuid.NewString()

	return fmt.Sprintf("uploads/%s/%d/%s/%s%s", fileType, userID, date, id, ext)
}
