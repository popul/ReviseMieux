// Package main - Outil CLI pour les migrations de base de données
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/revisemieux/backend/internal/store"
)

func main() {
	// Charger .env si présent
	godotenv.Load()

	// Flags
	var (
		action        = flag.String("action", "up", "Action: up, down, status")
		migrationsDir = flag.String("dir", "", "Dossier des migrations (défaut: ./migrations)")
	)
	flag.Parse()

	// Déterminer le dossier des migrations
	if *migrationsDir == "" {
		// Chercher le dossier migrations relatif à l'exécutable ou au répertoire courant
		cwd, _ := os.Getwd()
		possibles := []string{
			filepath.Join(cwd, "migrations"),
			filepath.Join(cwd, "backend", "migrations"),
			"./migrations",
		}
		for _, p := range possibles {
			if info, err := os.Stat(p); err == nil && info.IsDir() {
				*migrationsDir = p
				break
			}
		}
		if *migrationsDir == "" {
			log.Fatal("Dossier migrations introuvable. Utilisez -dir pour spécifier le chemin.")
		}
	}

	// URL de la base de données
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/revisemieux?sslmode=disable"
	}

	// Connexion à la base de données
	s, err := store.NouveauStore(databaseURL)
	if err != nil {
		log.Fatalf("Erreur connexion base de données: %v", err)
	}
	defer s.Fermer()

	// Exécuter l'action
	switch *action {
	case "up":
		fmt.Println("Exécution des migrations...")
		if err := s.Migrer(*migrationsDir); err != nil {
			log.Fatalf("Erreur migration: %v", err)
		}
		fmt.Println("Migrations terminées avec succès.")

	case "down":
		fmt.Println("Annulation de la dernière migration...")
		if err := s.Rollback(*migrationsDir); err != nil {
			log.Fatalf("Erreur rollback: %v", err)
		}
		fmt.Println("Rollback terminé avec succès.")

	case "status":
		appliquees, enAttente, err := s.StatutMigrations(*migrationsDir)
		if err != nil {
			log.Fatalf("Erreur statut: %v", err)
		}

		fmt.Println("\n=== Migrations appliquées ===")
		if len(appliquees) == 0 {
			fmt.Println("Aucune migration appliquée.")
		} else {
			for _, m := range appliquees {
				fmt.Printf("  ✓ %03d_%s (appliquée le %s)\n",
					m.Version,
					store.NettoyerNomMigration(m.Name),
					m.AppliedAt.Format("2006-01-02 15:04:05"),
				)
			}
		}

		fmt.Println("\n=== Migrations en attente ===")
		if len(enAttente) == 0 {
			fmt.Println("Aucune migration en attente.")
		} else {
			for _, m := range enAttente {
				fmt.Printf("  ○ %03d_%s\n", m.Version, store.NettoyerNomMigration(m.Name))
			}
		}
		fmt.Println()

	default:
		log.Fatalf("Action inconnue: %s (utilisez up, down, ou status)", *action)
	}
}
