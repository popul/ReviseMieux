package services

import (
	"fmt"
	"strings"
	"testing"
)

func TestNouveauServiceOCR_NombreMaxPages(t *testing.T) {
	tests := []struct {
		nom     string
		entree  int
		attendu int
	}{
		{"valeur positive", 50, 50},
		{"valeur zero", 0, 30},
		{"valeur negative", -5, 30},
		{"valeur un", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.nom, func(t *testing.T) {
			svc := NouveauServiceOCR(nil, false, tt.entree)
			if svc.NombreMaxPages() != tt.attendu {
				t.Errorf("NombreMaxPages() = %d, attendu %d", svc.NombreMaxPages(), tt.attendu)
			}
		})
	}
}

func TestErreurTropDePages_MessageDynamique(t *testing.T) {
	svc := NouveauServiceOCR(nil, false, 42)
	err := svc.erreurTropDePages()

	attendu := fmt.Sprintf("Trop de pages (max %d)", 42)
	if err.Message != attendu {
		t.Errorf("message = %q, attendu %q", err.Message, attendu)
	}
	if err.Code != "TROP_DE_PAGES" {
		t.Errorf("code = %q, attendu %q", err.Code, "TROP_DE_PAGES")
	}
	if !strings.Contains(err.Error(), "42") {
		t.Errorf("Error() devrait contenir 42, obtenu %q", err.Error())
	}
}
