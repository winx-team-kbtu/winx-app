package photo

import (
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	dto "winx-profile/internal/app/domain/core/dto/services/photo"
	repository "winx-profile/internal/app/domain/repositories/photo"
	"winx-profile/internal/app/models/models"
	"winx-profile/pkg/s3storage"

	"gorm.io/gorm"
)

var ErrInvalidPhoto = errors.New("photo must be an image")
var ErrNotFound = errors.New("profile photo not found")

type Service interface {
	GetByUserID(ctx context.Context, userID int64) (models.ProfilePhoto, error)
	Store(ctx context.Context, input dto.StoreDTO) (models.ProfilePhoto, bool, error)
}

type service struct {
	repository repository.Repository
	storage    s3storage.Storage
}

const signedPhotoURLTTL = 24 * time.Hour

func NewService(db *gorm.DB, storage s3storage.Storage) Service {
	return &service{
		repository: repository.NewRepository(db),
		storage:    storage,
	}
}

func (s *service) GetByUserID(ctx context.Context, userID int64) (models.ProfilePhoto, error) {
	photo, err := s.repository.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.ProfilePhoto{}, ErrNotFound
		}

		return models.ProfilePhoto{}, fmt.Errorf("get profile photo: %w", err)
	}

	accessURL, err := s.storage.SignedGetURL(ctx, photo.ObjectKey, signedPhotoURLTTL)
	if err != nil {
		return models.ProfilePhoto{}, fmt.Errorf("presign photo url: %w", err)
	}

	photo.URL = accessURL

	return photo, nil
}

func (s *service) Store(ctx context.Context, input dto.StoreDTO) (models.ProfilePhoto, bool, error) {
	contentType, err := detectContentType(input)
	if err != nil {
		return models.ProfilePhoto{}, false, err
	}

	width, height, err := readImageDimensions(input.File)
	if err != nil {
		return models.ProfilePhoto{}, false, fmt.Errorf("read image dimensions: %w", err)
	}

	if _, err := input.File.Seek(0, io.SeekStart); err != nil {
		return models.ProfilePhoto{}, false, fmt.Errorf("rewind photo file: %w", err)
	}

	objectKey := buildObjectKey(input.UserID, input.FileName)
	url, err := s.storage.Upload(ctx, objectKey, input.File, contentType)
	if err != nil {
		return models.ProfilePhoto{}, false, fmt.Errorf("upload photo to s3: %w", err)
	}

	accessURL, err := s.storage.SignedGetURL(ctx, objectKey, signedPhotoURLTTL)
	if err != nil {
		return models.ProfilePhoto{}, false, fmt.Errorf("presign photo url: %w", err)
	}

	current, err := s.repository.GetByUserID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			photo := newPhoto(input, s.storage.Bucket(), objectKey, url, contentType, width, height)
			created, createErr := s.repository.Create(ctx, photo)
			if createErr != nil {
				return models.ProfilePhoto{}, false, fmt.Errorf("create profile photo: %w", createErr)
			}

			created.URL = accessURL

			return created, true, nil
		}

		return models.ProfilePhoto{}, false, fmt.Errorf("get profile photo: %w", err)
	}

	current.Email = input.Email
	current.Provider = "s3"
	current.Bucket = s.storage.Bucket()
	current.ObjectKey = objectKey
	current.URL = url
	current.MimeType = contentType
	current.SizeBytes = input.SizeBytes
	current.Width = width
	current.Height = height
	current.UpdatedAt = time.Now().UTC()

	updated, err := s.repository.Update(ctx, current)
	if err != nil {
		return models.ProfilePhoto{}, false, fmt.Errorf("update profile photo: %w", err)
	}

	updated.URL = accessURL

	return updated, false, nil
}

func newPhoto(
	input dto.StoreDTO,
	bucket string,
	objectKey string,
	url string,
	contentType string,
	width *int,
	height *int,
) models.ProfilePhoto {
	now := time.Now().UTC()

	return models.ProfilePhoto{
		UserID:    input.UserID,
		Email:     input.Email,
		Provider:  "s3",
		Bucket:    bucket,
		ObjectKey: objectKey,
		URL:       url,
		MimeType:  contentType,
		SizeBytes: input.SizeBytes,
		Width:     width,
		Height:    height,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func detectContentType(input dto.StoreDTO) (string, error) {
	if input.File == nil {
		return "", ErrInvalidPhoto
	}

	contentType := strings.TrimSpace(input.ContentType)
	if contentType == "" || contentType == "application/octet-stream" {
		head := make([]byte, 512)
		n, err := input.File.Read(head)
		if err != nil && !errors.Is(err, io.EOF) {
			return "", fmt.Errorf("read photo header: %w", err)
		}

		contentType = http.DetectContentType(head[:n])

		if _, err := input.File.Seek(0, io.SeekStart); err != nil {
			return "", fmt.Errorf("rewind photo header: %w", err)
		}
	}

	if !strings.HasPrefix(contentType, "image/") {
		return "", ErrInvalidPhoto
	}

	return contentType, nil
}

func readImageDimensions(file io.ReadSeeker) (*int, *int, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, nil, err
	}

	cfg, _, err := image.DecodeConfig(file)
	if err != nil {
		if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
			return nil, nil, seekErr
		}

		return nil, nil, nil
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, nil, err
	}

	width := cfg.Width
	height := cfg.Height

	return &width, &height, nil
}

var invalidFileNameChars = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

func buildObjectKey(userID int64, fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	base := strings.TrimSuffix(fileName, ext)
	base = invalidFileNameChars.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" {
		base = "photo"
	}
	if ext == "" {
		ext = ".jpg"
	}

	return fmt.Sprintf("profiles/%d/%d-%s%s", userID, time.Now().UTC().UnixNano(), base, ext)
}
