package ocr

import (
	"strings"

	"github.com/revisemieux/backend/internal/llm"
)

// Seuils de correspondance
const (
	seuilSimilariteTexte  = 0.5  // similarité minimale pour remplacer le texte Tesseract par celui du LLM
	seuilConfianceMoyenne = 50.0 // confiance moyenne Tesseract minimale pour utiliser ses positions
	seuilConfianceBloc    = 40.0 // confiance minimale par bloc individuel
	longueurTexteMin      = 3    // nombre minimum de caractères significatifs par bloc
	paddingPx             = 4    // padding en pixels ajouté autour des bounding boxes Tesseract
	seuilMatchMinimal     = 0.3  // proportion minimale de blocs Tesseract matchés au LLM
)

// FusionnerBlocsHybride construit les blocs overlay à partir de Tesseract et du LLM.
//
// Stratégie :
//  1. Si Tesseract a une confiance suffisante ET que ses blocs matchent le texte LLM
//     → positions Tesseract (pixel-perfect) + texte LLM
//  2. Sinon → fallback sur les blocs LLM directement (positions approximatives)
//  3. Si ni Tesseract ni LLM n'ont de blocs → nil (pas d'overlay)
func FusionnerBlocsHybride(
	llmBlocs []llm.BlocTexte,
	tesseractBlocs []BlocTesseract,
	largeurImage, hauteurImage int,
) []llm.BlocTexte {
	if largeurImage <= 0 || hauteurImage <= 0 {
		return nil
	}

	// Calculer la confiance moyenne Tesseract
	confianceMoyenne := 0.0
	if len(tesseractBlocs) > 0 {
		var confianceTotale float64
		for _, b := range tesseractBlocs {
			confianceTotale += b.Confiance
		}
		confianceMoyenne = confianceTotale / float64(len(tesseractBlocs))
	}

	// Si Tesseract a une confiance suffisante → tenter la fusion
	if len(tesseractBlocs) > 0 && confianceMoyenne >= seuilConfianceMoyenne {
		resultat, tauxMatch := fusionnerAvecTesseract(llmBlocs, tesseractBlocs, largeurImage, hauteurImage)

		// Si suffisamment de blocs ont matché le LLM, le résultat Tesseract est fiable
		if tauxMatch >= seuilMatchMinimal || len(llmBlocs) == 0 {
			return resultat
		}
		// Sinon Tesseract a confiance OK mais texte garbage → fallback LLM
	}

	// Fallback : utiliser les blocs LLM directement (positions approximatives)
	if len(llmBlocs) > 0 {
		return llmBlocs
	}

	return nil
}

// fusionnerAvecTesseract utilise les positions pixel-perfect de Tesseract
// et remplace le texte par celui du LLM quand un match est trouvé.
// Retourne les blocs fusionnés et le taux de match (proportion de blocs matchés).
func fusionnerAvecTesseract(
	llmBlocs []llm.BlocTexte,
	tesseractBlocs []BlocTesseract,
	largeurImage, hauteurImage int,
) ([]llm.BlocTexte, float64) {
	resultat := make([]llm.BlocTexte, 0, len(tesseractBlocs))
	nbMatches := 0
	nbBlocsValides := 0

	for _, blocTess := range tesseractBlocs {
		if blocTess.Texte == "" {
			continue
		}

		// Filtrer les blocs de faible confiance ou trop courts
		if blocTess.Confiance < seuilConfianceBloc {
			continue
		}
		if len(strings.TrimSpace(blocTess.Texte)) < longueurTexteMin {
			continue
		}

		nbBlocsValides++

		// Appliquer un padding et clamper aux limites de l'image
		x := max(0, blocTess.X-paddingPx)
		y := max(0, blocTess.Y-paddingPx)
		maxX := min(largeurImage, blocTess.X+blocTess.Largeur+paddingPx)
		maxY := min(hauteurImage, blocTess.Y+blocTess.Hauteur+paddingPx)

		// Position Tesseract convertie en pourcentages
		position := llm.PositionBloc{
			X:       float64(x) / float64(largeurImage) * 100,
			Y:       float64(y) / float64(hauteurImage) * 100,
			Largeur: float64(maxX-x) / float64(largeurImage) * 100,
			Hauteur: float64(maxY-y) / float64(hauteurImage) * 100,
		}

		// Chercher le meilleur match LLM pour le texte
		texte := blocTess.Texte
		confiance := blocTess.Confiance / 100.0 // Tesseract 0-100 → 0-1

		meilleurScore := 0.0
		for _, blocLLM := range llmBlocs {
			score := similariteTexte(blocTess.Texte, blocLLM.Texte)
			if score > meilleurScore && score >= seuilSimilariteTexte {
				meilleurScore = score
				texte = blocLLM.Texte
				confiance = blocLLM.Confiance
			}
		}

		if meilleurScore >= seuilSimilariteTexte {
			nbMatches++
		}

		resultat = append(resultat, llm.BlocTexte{
			Texte:     texte,
			Position:  position,
			Confiance: confiance,
		})
	}

	tauxMatch := 0.0
	if nbBlocsValides > 0 {
		tauxMatch = float64(nbMatches) / float64(nbBlocsValides)
	}

	return resultat, tauxMatch
}
