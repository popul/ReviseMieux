package ocr

import (
	"math"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// normaliserTexte met en minuscules, supprime les accents et collapse les espaces
func normaliserTexte(s string) string {
	s = strings.ToLower(s)
	// Supprimer les accents via décomposition Unicode NFD + suppression des marques
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, s)
	if err != nil {
		result = s
	}
	// Collapser les espaces
	champs := strings.Fields(result)
	return strings.Join(champs, " ")
}

// trigrammes retourne l'ensemble des trigrammes de caractères d'une chaîne
func trigrammes(s string) map[string]struct{} {
	runes := []rune(s)
	result := make(map[string]struct{})
	for i := 0; i <= len(runes)-3; i++ {
		tri := string(runes[i : i+3])
		result[tri] = struct{}{}
	}
	return result
}

// similariteTexte calcule la similarité Jaccard sur les trigrammes de caractères.
// Retourne un score entre 0.0 et 1.0.
func similariteTexte(a, b string) float64 {
	na := normaliserTexte(a)
	nb := normaliserTexte(b)

	if na == "" || nb == "" {
		return 0
	}
	if na == nb {
		return 1
	}

	triA := trigrammes(na)
	triB := trigrammes(nb)

	if len(triA) == 0 || len(triB) == 0 {
		return 0
	}

	// Intersection
	inter := 0
	for t := range triA {
		if _, ok := triB[t]; ok {
			inter++
		}
	}

	// Union
	union := len(triA) + len(triB) - inter

	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

// Rect représente un rectangle pour le calcul d'IoU
type Rect struct {
	X, Y, Largeur, Hauteur float64
}

// calculerIoU calcule l'Intersection over Union de deux rectangles.
// Les coordonnées peuvent être en pixels ou en pourcentages (même unité pour les deux).
func calculerIoU(r1, r2 Rect) float64 {
	// Coins des rectangles
	x1Min, y1Min := r1.X, r1.Y
	x1Max, y1Max := r1.X+r1.Largeur, r1.Y+r1.Hauteur

	x2Min, y2Min := r2.X, r2.Y
	x2Max, y2Max := r2.X+r2.Largeur, r2.Y+r2.Hauteur

	// Intersection
	xOverlap := math.Max(0, math.Min(x1Max, x2Max)-math.Max(x1Min, x2Min))
	yOverlap := math.Max(0, math.Min(y1Max, y2Max)-math.Max(y1Min, y2Min))
	intersection := xOverlap * yOverlap

	// Union
	aire1 := r1.Largeur * r1.Hauteur
	aire2 := r2.Largeur * r2.Hauteur
	union := aire1 + aire2 - intersection

	if union <= 0 {
		return 0
	}
	return intersection / union
}
