package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChargerDefauts(t *testing.T) {
	// Sans YAML ni env var → default 30
	cfg := ChargerAvecFichier("inexistant.yaml")
	if cfg.NombreMaxPages != 30 {
		t.Errorf("attendu 30, obtenu %d", cfg.NombreMaxPages)
	}
}

func TestChargerDepuisYAML(t *testing.T) {
	dir := t.TempDir()
	chemin := filepath.Join(dir, "config.yaml")
	os.WriteFile(chemin, []byte("nombre_max_pages: 50\n"), 0644)

	cfg := ChargerAvecFichier(chemin)
	if cfg.NombreMaxPages != 50 {
		t.Errorf("attendu 50, obtenu %d", cfg.NombreMaxPages)
	}
}

func TestEnvVarOverrideYAML(t *testing.T) {
	dir := t.TempDir()
	chemin := filepath.Join(dir, "config.yaml")
	os.WriteFile(chemin, []byte("nombre_max_pages: 50\n"), 0644)

	t.Setenv("NOMBRE_MAX_PAGES", "100")

	cfg := ChargerAvecFichier(chemin)
	if cfg.NombreMaxPages != 100 {
		t.Errorf("attendu 100, obtenu %d", cfg.NombreMaxPages)
	}
}

func TestEnvVarSansYAML(t *testing.T) {
	t.Setenv("NOMBRE_MAX_PAGES", "15")

	cfg := ChargerAvecFichier("inexistant.yaml")
	if cfg.NombreMaxPages != 15 {
		t.Errorf("attendu 15, obtenu %d", cfg.NombreMaxPages)
	}
}

func TestYAMLInvalide(t *testing.T) {
	dir := t.TempDir()
	chemin := filepath.Join(dir, "config.yaml")
	os.WriteFile(chemin, []byte("{{invalid yaml content"), 0644)

	cfg := ChargerAvecFichier(chemin)
	if cfg.NombreMaxPages != 30 {
		t.Errorf("attendu 30 (fallback), obtenu %d", cfg.NombreMaxPages)
	}
}

func TestYAMLAvecValeurZero(t *testing.T) {
	dir := t.TempDir()
	chemin := filepath.Join(dir, "config.yaml")
	os.WriteFile(chemin, []byte("nombre_max_pages: 0\n"), 0644)

	cfg := ChargerAvecFichier(chemin)
	if cfg.NombreMaxPages != 30 {
		t.Errorf("attendu 30 (fallback pour 0), obtenu %d", cfg.NombreMaxPages)
	}
}
