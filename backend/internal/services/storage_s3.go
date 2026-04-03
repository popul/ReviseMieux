// Package services contient le service de stockage S3/MinIO
package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ServiceStorageS3 gere le stockage des fichiers sur S3/MinIO
type ServiceStorageS3 struct {
	client     *minio.Client
	bucketName string
}

// ConfigS3 contient la configuration du service S3
type ConfigS3 struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// NouveauServiceStorageS3 cree une nouvelle instance du service de stockage S3/MinIO
func NouveauServiceStorageS3(cfg ConfigS3) (*ServiceStorageS3, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("services.StorageS3: connexion MinIO: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Creer le bucket s'il n'existe pas
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("services.StorageS3: verification bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("services.StorageS3: creation bucket %q: %w", cfg.Bucket, err)
		}
	}

	return &ServiceStorageS3{
		client:     client,
		bucketName: cfg.Bucket,
	}, nil
}

// SauvegarderImage sauvegarde une image pour un cours et retourne le nom du fichier
func (s *ServiceStorageS3) SauvegarderImage(coursID string, file *multipart.FileHeader) (string, error) {
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	nomFichier := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	objectKey := fmt.Sprintf("cours/%s/%s", coursID, nomFichier)

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("services.StorageS3: ouverture fichier: %w", err)
	}
	defer src.Close()

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err = s.client.PutObject(ctx, s.bucketName, objectKey, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("services.StorageS3: upload %q: %w", objectKey, err)
	}

	return nomFichier, nil
}

// SauvegarderImageBytes sauvegarde des donnees d'image pour un cours
func (s *ServiceStorageS3) SauvegarderImageBytes(coursID string, data []byte, extension string) (string, error) {
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	nomFichier := fmt.Sprintf("%s%s", uuid.New().String(), extension)
	objectKey := fmt.Sprintf("cours/%s/%s", coursID, nomFichier)

	contentType := detecterContentType(data, extension)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err := s.client.PutObject(ctx, s.bucketName, objectKey, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("services.StorageS3: upload bytes %q: %w", objectKey, err)
	}

	return nomFichier, nil
}

// ObtenirImage retourne le chemin (cle S3) d'une image.
// Pour la compatibilite avec ServiceStorage, on retourne la cle de l'objet.
// Le handler HTTP devra utiliser ObtenirImageData pour servir le fichier.
func (s *ServiceStorageS3) ObtenirImage(coursID, nomFichier string) (string, error) {
	nomFichier = filepath.Base(nomFichier)
	objectKey := fmt.Sprintf("cours/%s/%s", coursID, nomFichier)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.client.StatObject(ctx, s.bucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("services.StorageS3: image non trouvee: %s", nomFichier)
	}

	return objectKey, nil
}

// ObtenirImageData retourne les donnees binaires d'une image
func (s *ServiceStorageS3) ObtenirImageData(coursID, nomFichier string) ([]byte, string, error) {
	nomFichier = filepath.Base(nomFichier)
	objectKey := fmt.Sprintf("cours/%s/%s", coursID, nomFichier)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	obj, err := s.client.GetObject(ctx, s.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("services.StorageS3: recuperation %q: %w", objectKey, err)
	}
	defer obj.Close()

	info, err := obj.Stat()
	if err != nil {
		return nil, "", fmt.Errorf("services.StorageS3: stat %q: %w", objectKey, err)
	}

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, "", fmt.Errorf("services.StorageS3: lecture %q: %w", objectKey, err)
	}

	return data, info.ContentType, nil
}

// EcraserImage remplace le contenu d'une image existante
func (s *ServiceStorageS3) EcraserImage(coursID, nomFichier string, data []byte) error {
	nomFichier = filepath.Base(nomFichier)
	objectKey := fmt.Sprintf("cours/%s/%s", coursID, nomFichier)

	ext := filepath.Ext(nomFichier)
	contentType := detecterContentType(data, ext)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err := s.client.PutObject(ctx, s.bucketName, objectKey, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("services.StorageS3: ecrasement %q: %w", objectKey, err)
	}

	return nil
}

// SupprimerImage supprime une image d'un cours
func (s *ServiceStorageS3) SupprimerImage(coursID, nomFichier string) error {
	nomFichier = filepath.Base(nomFichier)
	objectKey := fmt.Sprintf("cours/%s/%s", coursID, nomFichier)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.client.RemoveObject(ctx, s.bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("services.StorageS3: suppression %q: %w", objectKey, err)
	}

	return nil
}

// SupprimerToutesImages supprime toutes les images d'un cours
func (s *ServiceStorageS3) SupprimerToutesImages(coursID string) error {
	prefix := fmt.Sprintf("cours/%s/", coursID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectsCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for obj := range objectsCh {
		if obj.Err != nil {
			return fmt.Errorf("services.StorageS3: listage %q: %w", prefix, obj.Err)
		}
		err := s.client.RemoveObject(ctx, s.bucketName, obj.Key, minio.RemoveObjectOptions{})
		if err != nil {
			return fmt.Errorf("services.StorageS3: suppression %q: %w", obj.Key, err)
		}
	}

	return nil
}

// detecterContentType detecte le content type a partir des donnees et de l'extension
func detecterContentType(data []byte, ext string) string {
	// Essayer la detection par magic bytes
	ct := http.DetectContentType(data)
	if ct != "application/octet-stream" {
		return ct
	}

	// Fallback sur l'extension
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}
