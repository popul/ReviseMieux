package services

import (
	"testing"
)

func TestMelangerChoix(t *testing.T) {
	t.Run("préserve tous les choix", func(t *testing.T) {
		choix := []string{"A", "B", "C", "D"}
		indexCorrect := 0

		resultat, _ := melangerChoix(choix, indexCorrect)

		// Vérifier que tous les choix sont présents
		if len(resultat) != 4 {
			t.Errorf("attendu 4 choix, obtenu %d", len(resultat))
		}

		// Créer un set des choix originaux et résultants
		originaux := make(map[string]bool)
		for _, c := range choix {
			originaux[c] = true
		}

		for _, c := range resultat {
			if !originaux[c] {
				t.Errorf("choix inattendu dans le résultat: %s", c)
			}
		}
	})

	t.Run("la bonne réponse reste correctement indexée", func(t *testing.T) {
		choix := []string{"Bonne réponse", "Mauvaise 1", "Mauvaise 2", "Mauvaise 3"}
		indexCorrect := 0

		// Exécuter plusieurs fois pour s'assurer que l'index suit le mélange
		for i := 0; i < 100; i++ {
			resultat, nouvelIndex := melangerChoix(choix, indexCorrect)

			// Vérifier que le nouvel index pointe vers la bonne réponse
			if resultat[nouvelIndex] != "Bonne réponse" {
				t.Errorf("l'index %d ne pointe pas vers la bonne réponse, pointe vers: %s", nouvelIndex, resultat[nouvelIndex])
			}
		}
	})

	t.Run("la bonne réponse n'est pas toujours en première position", func(t *testing.T) {
		choix := []string{"Bonne réponse", "Mauvaise 1", "Mauvaise 2", "Mauvaise 3"}
		indexCorrect := 0

		// Exécuter plusieurs fois et compter les positions
		positions := make(map[int]int)
		iterations := 1000

		for i := 0; i < iterations; i++ {
			_, nouvelIndex := melangerChoix(choix, indexCorrect)
			positions[nouvelIndex]++
		}

		// Vérifier que la bonne réponse apparaît à différentes positions
		// (statistiquement, avec 1000 itérations, chaque position devrait avoir ~250 occurrences)
		if len(positions) < 2 {
			t.Error("la bonne réponse semble toujours être à la même position")
		}

		// Vérifier que la position 0 n'a pas plus de 40% des cas (avec une marge)
		if positions[0] > iterations*40/100 {
			t.Errorf("la bonne réponse est trop souvent en position 0: %d/%d fois", positions[0], iterations)
		}
	})

	t.Run("gère correctement un index correct différent de 0", func(t *testing.T) {
		choix := []string{"Mauvaise 1", "Mauvaise 2", "Bonne réponse", "Mauvaise 3"}
		indexCorrect := 2

		for i := 0; i < 100; i++ {
			resultat, nouvelIndex := melangerChoix(choix, indexCorrect)

			if resultat[nouvelIndex] != "Bonne réponse" {
				t.Errorf("l'index %d ne pointe pas vers la bonne réponse, pointe vers: %s", nouvelIndex, resultat[nouvelIndex])
			}
		}
	})

	t.Run("gère une liste vide", func(t *testing.T) {
		choix := []string{}
		indexCorrect := 0

		resultat, nouvelIndex := melangerChoix(choix, indexCorrect)

		if len(resultat) != 0 {
			t.Errorf("attendu liste vide, obtenu %d éléments", len(resultat))
		}
		if nouvelIndex != 0 {
			t.Errorf("attendu index 0, obtenu %d", nouvelIndex)
		}
	})
}
