// Package llm fournit les adaptateurs pour les services LLM (OpenAI, Mistral)
package llm

import (
	"context"
	"log"
)

// GestionnaireLLM gère les adaptateurs LLM avec fallback automatique
type GestionnaireLLM struct {
	primaire AdaptateurLLM
	fallback AdaptateurLLM
}

// NouveauGestionnaire crée un nouveau gestionnaire LLM
// primaire est l'adaptateur principal (généralement OpenAI)
// fallback est l'adaptateur de secours (généralement Mistral), peut être nil
func NouveauGestionnaire(primaire, fallback AdaptateurLLM) *GestionnaireLLM {
	return &GestionnaireLLM{
		primaire: primaire,
		fallback: fallback,
	}
}

// NouveauGestionnaireAvecCles crée un gestionnaire avec les clés API fournies
// Retourne nil si aucune clé n'est fournie
func NouveauGestionnaireAvecCles(cleOpenAI, cleMistral string) *GestionnaireLLM {
	var primaire, fallback AdaptateurLLM

	if cleOpenAI != "" {
		primaire = NouveauClientOpenAI(cleOpenAI)
	}

	if cleMistral != "" {
		fallback = NouveauClientMistral(cleMistral)
	}

	// Si pas de primaire mais un fallback, utiliser le fallback comme primaire
	if primaire == nil && fallback != nil {
		primaire = fallback
		fallback = nil
	}

	// Si aucun adaptateur disponible, retourner nil
	if primaire == nil {
		return nil
	}

	return NouveauGestionnaire(primaire, fallback)
}

// GenererTexte génère du texte avec fallback automatique
func (g *GestionnaireLLM) GenererTexte(ctx context.Context, prompt string, options OptionsGeneration) (string, error) {
	// Essaye d'abord le fournisseur primaire
	resultat, err := g.primaire.GenererTexte(ctx, prompt, options)
	if err == nil {
		return resultat, nil
	}

	// Vérifie si on peut faire un fallback
	if !g.doitFallback(err) {
		return "", err
	}

	// Fallback vers le fournisseur secondaire
	log.Printf("[LLM] Fallback de %s vers %s: %s", g.primaire.Nom(), g.fallback.Nom(), err.Error())
	return g.fallback.GenererTexte(ctx, prompt, options)
}

// GenererJSON génère du JSON structuré avec fallback automatique
func (g *GestionnaireLLM) GenererJSON(ctx context.Context, prompt string, schema interface{}, options OptionsGeneration) ([]byte, error) {
	// Essaye d'abord le fournisseur primaire
	resultat, err := g.primaire.GenererJSON(ctx, prompt, schema, options)
	if err == nil {
		return resultat, nil
	}

	// Vérifie si on peut faire un fallback
	if !g.doitFallback(err) {
		return nil, err
	}

	// Fallback vers le fournisseur secondaire
	log.Printf("[LLM] Fallback de %s vers %s: %s", g.primaire.Nom(), g.fallback.Nom(), err.Error())
	return g.fallback.GenererJSON(ctx, prompt, schema, options)
}

// ExtraireTexteImage extrait du texte d'une image avec fallback automatique
func (g *GestionnaireLLM) ExtraireTexteImage(ctx context.Context, image []byte, options OptionsOCR) (*ResultatOCR, error) {
	// Essaye d'abord le fournisseur primaire
	resultat, err := g.primaire.ExtraireTexteImage(ctx, image, options)
	if err == nil {
		return resultat, nil
	}

	// Vérifie si on peut faire un fallback
	if !g.doitFallback(err) {
		return nil, err
	}

	// Fallback vers le fournisseur secondaire
	log.Printf("[LLM] Fallback de %s vers %s: %s", g.primaire.Nom(), g.fallback.Nom(), err.Error())
	return g.fallback.ExtraireTexteImage(ctx, image, options)
}

// EstDisponible vérifie si au moins un adaptateur est disponible
func (g *GestionnaireLLM) EstDisponible(ctx context.Context) bool {
	if g.primaire.EstDisponible(ctx) {
		return true
	}
	if g.fallback != nil {
		return g.fallback.EstDisponible(ctx)
	}
	return false
}

// Nom retourne le nom du fournisseur actif
func (g *GestionnaireLLM) Nom() string {
	return g.primaire.Nom()
}

// FournisseurActif retourne l'adaptateur primaire
func (g *GestionnaireLLM) FournisseurActif() AdaptateurLLM {
	return g.primaire
}

// FournisseurFallback retourne l'adaptateur de fallback (peut être nil)
func (g *GestionnaireLLM) FournisseurFallback() AdaptateurLLM {
	return g.fallback
}

// doitFallback détermine si on doit basculer vers le fallback
func (g *GestionnaireLLM) doitFallback(err error) bool {
	// Pas de fallback disponible
	if g.fallback == nil {
		return false
	}

	// Vérifie si c'est une erreur récupérable (rate limit, erreur serveur, etc.)
	if errLLM, ok := err.(*ErreurLLM); ok {
		return errLLM.Recuperable
	}

	// Pour les autres erreurs (network, timeout), on tente le fallback
	return true
}

// Vérifie que GestionnaireLLM implémente AdaptateurLLM
var _ AdaptateurLLM = (*GestionnaireLLM)(nil)
