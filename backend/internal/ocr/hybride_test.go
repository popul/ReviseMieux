package ocr

import (
	"math"
	"testing"

	"github.com/revisemieux/backend/internal/llm"
)

func TestNormaliserTexte(t *testing.T) {
	tests := []struct {
		entree string
		attend string
	}{
		{"Hello World", "hello world"},
		{"  Espaces   multiples  ", "espaces multiples"},
		{"Éléphant café résumé", "elephant cafe resume"},
		{"", ""},
		{"déjà vu", "deja vu"},
	}

	for _, tt := range tests {
		resultat := normaliserTexte(tt.entree)
		if resultat != tt.attend {
			t.Errorf("normaliserTexte(%q) = %q, attendu %q", tt.entree, resultat, tt.attend)
		}
	}
}

func TestSimilariteTexte(t *testing.T) {
	tests := []struct {
		a, b     string
		minScore float64
		maxScore float64
	}{
		{"hello world", "hello world", 1.0, 1.0},
		{"hello world", "hello worlds", 0.7, 1.0},
		{"bonjour le monde", "bonjour le monde!", 0.8, 1.0},
		{"abc", "xyz", 0.0, 0.1},
		{"", "hello", 0.0, 0.0},
		{"", "", 0.0, 0.0},
		{"Le théorème de Pythagore", "Le theoreme de Pythagore", 0.8, 1.0},
	}

	for _, tt := range tests {
		score := similariteTexte(tt.a, tt.b)
		if score < tt.minScore || score > tt.maxScore {
			t.Errorf("similariteTexte(%q, %q) = %.3f, attendu entre %.3f et %.3f",
				tt.a, tt.b, score, tt.minScore, tt.maxScore)
		}
	}
}

func TestCalculerIoU(t *testing.T) {
	tests := []struct {
		r1, r2 Rect
		attend float64
	}{
		// Rectangles identiques → IoU = 1.0
		{Rect{0, 0, 100, 100}, Rect{0, 0, 100, 100}, 1.0},
		// Aucun chevauchement
		{Rect{0, 0, 50, 50}, Rect{60, 60, 50, 50}, 0.0},
		// Chevauchement partiel : 25x25=625, union=2*2500-625=4375, IoU≈0.143
		{Rect{0, 0, 50, 50}, Rect{25, 25, 50, 50}, 625.0 / 4375.0},
		// Un rectangle contenu dans l'autre
		{Rect{0, 0, 100, 100}, Rect{25, 25, 50, 50}, 2500.0 / 10000.0},
	}

	for _, tt := range tests {
		iou := calculerIoU(tt.r1, tt.r2)
		if math.Abs(iou-tt.attend) > 0.001 {
			t.Errorf("calculerIoU(%v, %v) = %.4f, attendu %.4f", tt.r1, tt.r2, iou, tt.attend)
		}
	}
}

func TestFusionnerBlocsHybride_PositionsTesseractUniquement(t *testing.T) {
	llmBlocs := []llm.BlocTexte{
		{
			Texte:     "Le théorème de Pythagore",
			Confiance: 0.9,
		},
	}

	tessBlocs := []BlocTesseract{
		{
			Texte:     "Le theoreme de Pythagore",
			X:         50,
			Y:         100,
			Largeur:   900,
			Hauteur:   80,
			Confiance: 85,
		},
	}

	resultat := FusionnerBlocsHybride(llmBlocs, tessBlocs, 1000, 1000)

	if len(resultat) != 1 {
		t.Fatalf("attendu 1 bloc, obtenu %d", len(resultat))
	}

	// La position doit venir de Tesseract (avec padding de 4px)
	pos := resultat[0].Position
	attenduX := float64(50-paddingPx) / 1000 * 100   // (50-4)/1000*100 = 4.6
	attenduY := float64(100-paddingPx) / 1000 * 100   // (100-4)/1000*100 = 9.6
	if math.Abs(pos.X-attenduX) > 0.1 || math.Abs(pos.Y-attenduY) > 0.1 {
		t.Errorf("position incorrecte: got (%.1f, %.1f), attendu (%.1f, %.1f)",
			pos.X, pos.Y, attenduX, attenduY)
	}

	// Le texte doit être celui du LLM (match par similarité)
	if resultat[0].Texte != "Le théorème de Pythagore" {
		t.Errorf("le texte devrait venir du LLM: got %q", resultat[0].Texte)
	}
}

func TestFusionnerBlocsHybride_SansTesseract_FallbackLLM(t *testing.T) {
	llmBlocs := []llm.BlocTexte{
		{
			Texte:     "Texte du LLM",
			Position:  llm.PositionBloc{X: 10, Y: 20, Largeur: 80, Hauteur: 5},
			Confiance: 0.9,
		},
	}

	// Sans Tesseract → fallback sur les blocs LLM
	resultat := FusionnerBlocsHybride(llmBlocs, nil, 1000, 1000)
	if len(resultat) != 1 {
		t.Fatalf("fallback LLM: attendu 1 bloc, obtenu %d", len(resultat))
	}
	if resultat[0].Texte != "Texte du LLM" {
		t.Errorf("fallback LLM: mauvais texte: got %q", resultat[0].Texte)
	}
	if resultat[0].Position.X != 10 {
		t.Errorf("fallback LLM: position devrait venir du LLM: X got %.1f", resultat[0].Position.X)
	}
}

func TestFusionnerBlocsHybride_SansTesseractNiLLM_RetourneNil(t *testing.T) {
	resultat := FusionnerBlocsHybride(nil, nil, 1000, 1000)
	if resultat != nil {
		t.Errorf("sans Tesseract ni LLM, devrait retourner nil")
	}
}

func TestFusionnerBlocsHybride_FaibleConfiance_FallbackLLM(t *testing.T) {
	llmBlocs := []llm.BlocTexte{
		{
			Texte:     "Texte manuscrit",
			Position:  llm.PositionBloc{X: 5, Y: 15, Largeur: 90, Hauteur: 8},
			Confiance: 0.9,
		},
	}

	tessBlocs := []BlocTesseract{
		{
			Texte:     "Txet mcnscrit",
			X:         50,
			Y:         100,
			Largeur:   900,
			Hauteur:   80,
			Confiance: 15, // < seuilConfianceMoyenne (50)
		},
	}

	// Tesseract trop faible → fallback sur les blocs LLM
	resultat := FusionnerBlocsHybride(llmBlocs, tessBlocs, 1000, 1000)
	if len(resultat) != 1 {
		t.Fatalf("fallback LLM: attendu 1 bloc, obtenu %d", len(resultat))
	}
	if resultat[0].Texte != "Texte manuscrit" {
		t.Errorf("fallback LLM: devrait utiliser le texte LLM: got %q", resultat[0].Texte)
	}
	if resultat[0].Position.X != 5 {
		t.Errorf("fallback LLM: position devrait venir du LLM: X got %.1f", resultat[0].Position.X)
	}
}

func TestFusionnerBlocsHybride_BlocFaibleConfianceFiltre(t *testing.T) {
	// Un bloc à haute confiance + un bloc à faible confiance
	// La moyenne est assez haute pour passer le seuil global,
	// mais le bloc individuel faible doit être filtré
	tessBlocs := []BlocTesseract{
		{Texte: "Bloc correct", X: 10, Y: 10, Largeur: 200, Hauteur: 50, Confiance: 80},
		{Texte: "Bloc faible", X: 10, Y: 100, Largeur: 200, Hauteur: 50, Confiance: 20},
	}

	resultat := FusionnerBlocsHybride(nil, tessBlocs, 1000, 1000)

	if len(resultat) != 1 {
		t.Fatalf("attendu 1 bloc (le faible filtré), obtenu %d", len(resultat))
	}
	if resultat[0].Texte != "Bloc correct" {
		t.Errorf("mauvais texte conservé: got %q", resultat[0].Texte)
	}
}

func TestFusionnerBlocsHybride_TexteCourtFiltre(t *testing.T) {
	tessBlocs := []BlocTesseract{
		{Texte: "ab", X: 10, Y: 10, Largeur: 50, Hauteur: 20, Confiance: 90},  // trop court
		{Texte: "  x ", X: 10, Y: 50, Largeur: 50, Hauteur: 20, Confiance: 90}, // trop court après trim
		{Texte: "Texte suffisant", X: 10, Y: 100, Largeur: 200, Hauteur: 40, Confiance: 85},
	}

	resultat := FusionnerBlocsHybride(nil, tessBlocs, 1000, 1000)

	if len(resultat) != 1 {
		t.Fatalf("attendu 1 bloc (les courts filtrés), obtenu %d", len(resultat))
	}
	if resultat[0].Texte != "Texte suffisant" {
		t.Errorf("mauvais texte: got %q", resultat[0].Texte)
	}
}

func TestFusionnerBlocsHybride_TesseractGarbage_FallbackLLM(t *testing.T) {
	// Quand Tesseract a une confiance OK mais du texte garbage (aucun match LLM),
	// on doit faire fallback sur les blocs LLM
	llmBlocs := []llm.BlocTexte{
		{
			Texte:     "Texte completement different",
			Position:  llm.PositionBloc{X: 10, Y: 20, Largeur: 80, Hauteur: 5},
			Confiance: 0.9,
		},
	}

	tessBlocs := []BlocTesseract{
		{
			Texte:     "xkjf asd qwe",
			X:         100,
			Y:         200,
			Largeur:   400,
			Hauteur:   50,
			Confiance: 80,
		},
	}

	resultat := FusionnerBlocsHybride(llmBlocs, tessBlocs, 1000, 1000)

	if len(resultat) != 1 {
		t.Fatalf("fallback LLM: attendu 1 bloc, obtenu %d", len(resultat))
	}

	// Tesseract garbage → fallback sur le texte et la position LLM
	if resultat[0].Texte != "Texte completement different" {
		t.Errorf("devrait faire fallback LLM: got %q", resultat[0].Texte)
	}
	if resultat[0].Position.X != 10 {
		t.Errorf("position devrait venir du LLM: X got %.1f", resultat[0].Position.X)
	}
}

func TestFusionnerBlocsHybride_SansLLM_GardeTexteTesseract(t *testing.T) {
	// Sans blocs LLM, on garde Tesseract même sans match
	tessBlocs := []BlocTesseract{
		{
			Texte:     "Texte Tesseract seul",
			X:         100,
			Y:         200,
			Largeur:   400,
			Hauteur:   50,
			Confiance: 80,
		},
	}

	resultat := FusionnerBlocsHybride(nil, tessBlocs, 1000, 1000)

	if len(resultat) != 1 {
		t.Fatalf("attendu 1 bloc, obtenu %d", len(resultat))
	}
	if resultat[0].Texte != "Texte Tesseract seul" {
		t.Errorf("sans LLM, texte Tesseract conservé: got %q", resultat[0].Texte)
	}
}

func TestFusionnerBlocsHybride_Padding(t *testing.T) {
	tessBlocs := []BlocTesseract{
		{
			Texte:     "Test padding",
			X:         2, // proche du bord, padding clampé à 0
			Y:         2,
			Largeur:   100,
			Hauteur:   30,
			Confiance: 90,
		},
	}

	resultat := FusionnerBlocsHybride(nil, tessBlocs, 1000, 1000)

	if len(resultat) != 1 {
		t.Fatalf("attendu 1 bloc, obtenu %d", len(resultat))
	}

	// X et Y doivent être clampés à 0 (2 - 4 = -2 → 0)
	pos := resultat[0].Position
	if pos.X != 0 {
		t.Errorf("X devrait être clampé à 0, got %.1f", pos.X)
	}
	if pos.Y != 0 {
		t.Errorf("Y devrait être clampé à 0, got %.1f", pos.Y)
	}
}

func TestFusionnerBlocsHybride_BlocVideIgnore(t *testing.T) {
	tessBlocs := []BlocTesseract{
		{Texte: "", X: 10, Y: 10, Largeur: 100, Hauteur: 50, Confiance: 90},
		{Texte: "Texte valide", X: 10, Y: 100, Largeur: 200, Hauteur: 40, Confiance: 85},
	}

	resultat := FusionnerBlocsHybride(nil, tessBlocs, 1000, 1000)

	if len(resultat) != 1 {
		t.Fatalf("les blocs vides devraient être ignorés: attendu 1, obtenu %d", len(resultat))
	}
	if resultat[0].Texte != "Texte valide" {
		t.Errorf("mauvais texte: got %q", resultat[0].Texte)
	}
}

func TestFusionnerBlocsHybride_DimensionsInvalides(t *testing.T) {
	tessBlocs := []BlocTesseract{
		{Texte: "Test", X: 10, Y: 10, Largeur: 100, Hauteur: 50, Confiance: 90},
	}

	resultat := FusionnerBlocsHybride(nil, tessBlocs, 0, 1000)
	if resultat != nil {
		t.Errorf("largeur 0 devrait retourner nil")
	}

	resultat = FusionnerBlocsHybride(nil, tessBlocs, 1000, 0)
	if resultat != nil {
		t.Errorf("hauteur 0 devrait retourner nil")
	}
}
