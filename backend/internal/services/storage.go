// Package services contient le service de stockage des fichiers
package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// ServiceStorage gère le stockage des fichiers
type ServiceStorage struct {
	basePath string
}

// NouveauServiceStorage crée une nouvelle instance du service de stockage
func NouveauServiceStorage(basePath string) (*ServiceStorage, error) {
	// Créer le répertoire de base s'il n'existe pas
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("impossible de créer le répertoire de stockage: %w", err)
	}

	return &ServiceStorage{
		basePath: basePath,
	}, nil
}

// SauvegarderImage sauvegarde une image pour un cours et retourne le nom du fichier
func (s *ServiceStorage) SauvegarderImage(coursID string, file *multipart.FileHeader) (string, error) {
	// Créer le répertoire du cours
	coursDir := filepath.Join(s.basePath, "cours", coursID)
	if err := os.MkdirAll(coursDir, 0755); err != nil {
		return "", fmt.Errorf("impossible de créer le répertoire du cours: %w", err)
	}

	// Générer un nom unique pour le fichier
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	nomFichier := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	cheminComplet := filepath.Join(coursDir, nomFichier)

	// Ouvrir le fichier source
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("impossible d'ouvrir le fichier source: %w", err)
	}
	defer src.Close()

	// Créer le fichier destination
	dst, err := os.Create(cheminComplet)
	if err != nil {
		return "", fmt.Errorf("impossible de créer le fichier destination: %w", err)
	}
	defer dst.Close()

	// Copier le contenu
	if _, err = io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("impossible de copier le fichier: %w", err)
	}

	return nomFichier, nil
}

// SauvegarderImageBytes sauvegarde des données d'image pour un cours
func (s *ServiceStorage) SauvegarderImageBytes(coursID string, data []byte, extension string) (string, error) {
	// Créer le répertoire du cours
	coursDir := filepath.Join(s.basePath, "cours", coursID)
	if err := os.MkdirAll(coursDir, 0755); err != nil {
		return "", fmt.Errorf("impossible de créer le répertoire du cours: %w", err)
	}

	// Générer un nom unique pour le fichier
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	nomFichier := fmt.Sprintf("%s%s", uuid.New().String(), extension)
	cheminComplet := filepath.Join(coursDir, nomFichier)

	// Écrire les données
	if err := os.WriteFile(cheminComplet, data, 0644); err != nil {
		return "", fmt.Errorf("impossible d'écrire le fichier: %w", err)
	}

	return nomFichier, nil
}

// ObtenirImage retourne le chemin complet d'une image
func (s *ServiceStorage) ObtenirImage(coursID, nomFichier string) (string, error) {
	// Nettoyer le nom de fichier pour éviter les attaques de traversée de chemin
	nomFichier = filepath.Base(nomFichier)
	cheminComplet := filepath.Join(s.basePath, "cours", coursID, nomFichier)

	// Vérifier que le fichier existe
	if _, err := os.Stat(cheminComplet); os.IsNotExist(err) {
		return "", fmt.Errorf("image non trouvée: %s", nomFichier)
	}

	return cheminComplet, nil
}

// SupprimerImage supprime une image d'un cours
func (s *ServiceStorage) SupprimerImage(coursID, nomFichier string) error {
	nomFichier = filepath.Base(nomFichier)
	cheminComplet := filepath.Join(s.basePath, "cours", coursID, nomFichier)

	if err := os.Remove(cheminComplet); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("impossible de supprimer l'image: %w", err)
	}

	return nil
}

// SupprimerToutesImages supprime toutes les images d'un cours
func (s *ServiceStorage) SupprimerToutesImages(coursID string) error {
	coursDir := filepath.Join(s.basePath, "cours", coursID)
	if err := os.RemoveAll(coursDir); err != nil {
		return fmt.Errorf("impossible de supprimer le répertoire: %w", err)
	}
	return nil
}
