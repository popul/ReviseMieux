// Package api contient les handlers pour la gestion des images
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/revisemieux/backend/internal/services"
	"github.com/revisemieux/backend/internal/store"
)

// HandlersImages contient les handlers pour les images avec leurs dépendances
type HandlersImages struct {
	serviceStorage *services.ServiceStorage
	coursRepo      store.CoursRepository
}

// NouveauHandlersImages crée une nouvelle instance des handlers Images
func NouveauHandlersImages(serviceStorage *services.ServiceStorage, coursRepo store.CoursRepository) *HandlersImages {
	return &HandlersImages{
		serviceStorage: serviceStorage,
		coursRepo:      coursRepo,
	}
}

// ServirImageHandler sert une image d'un cours
// GET /api/cours/:id/images/:filename
func (h *HandlersImages) ServirImageHandler(c *gin.Context) {
	if h.serviceStorage == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de stockage n'est pas configuré",
			},
		})
		return
	}

	coursID := c.Param("id")
	filename := c.Param("filename")

	if coursID == "" || filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "PARAMETRES_MANQUANTS",
				"message": "L'ID du cours et le nom du fichier sont requis",
			},
		})
		return
	}

	cheminComplet, err := h.serviceStorage.ObtenirImage(coursID, filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "IMAGE_NON_TROUVEE",
				"message": err.Error(),
			},
		})
		return
	}

	c.Header("Cache-Control", "no-store, must-revalidate")
	c.File(cheminComplet)
}

// AjouterImageHandler ajoute une image à un cours
// POST /api/cours/:id/images
func (h *HandlersImages) AjouterImageHandler(c *gin.Context) {
	if h.serviceStorage == nil || h.coursRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de stockage n'est pas configuré",
			},
		})
		return
	}

	coursID := c.Param("id")
	if coursID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ID_MANQUANT",
				"message": "L'ID du cours est requis",
			},
		})
		return
	}

	// Récupérer le cours
	cours, err := h.coursRepo.ObtenirParID(c.Request.Context(), coursID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "COURS_NON_TROUVE",
				"message": "Cours non trouvé",
			},
		})
		return
	}

	// Récupérer le fichier uploadé
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "FICHIER_MANQUANT",
				"message": "Aucun fichier image fourni",
			},
		})
		return
	}

	// Sauvegarder l'image
	nomFichier, err := h.serviceStorage.SauvegarderImage(coursID, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ERREUR_SAUVEGARDE",
				"message": err.Error(),
			},
		})
		return
	}

	// Ajouter l'image à la liste du cours
	cours.Images = append(cours.Images, nomFichier)
	if err := h.coursRepo.MettreAJour(c.Request.Context(), cours); err != nil {
		// Supprimer l'image si la mise à jour échoue
		h.serviceStorage.SupprimerImage(coursID, nomFichier)
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ERREUR_MISE_A_JOUR",
				"message": "Erreur lors de la mise à jour du cours",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succes":     true,
		"nomFichier": nomFichier,
		"images":     cours.Images,
	})
}

// SupprimerImageHandler supprime une image d'un cours
// DELETE /api/cours/:id/images/:filename
func (h *HandlersImages) SupprimerImageHandler(c *gin.Context) {
	if h.serviceStorage == nil || h.coursRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de stockage n'est pas configuré",
			},
		})
		return
	}

	coursID := c.Param("id")
	filename := c.Param("filename")

	if coursID == "" || filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "PARAMETRES_MANQUANTS",
				"message": "L'ID du cours et le nom du fichier sont requis",
			},
		})
		return
	}

	// Récupérer le cours
	cours, err := h.coursRepo.ObtenirParID(c.Request.Context(), coursID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "COURS_NON_TROUVE",
				"message": "Cours non trouvé",
			},
		})
		return
	}

	// Supprimer l'image de la liste
	nouvellesImages := make([]string, 0, len(cours.Images))
	imageTrouvee := false
	for _, img := range cours.Images {
		if img != filename {
			nouvellesImages = append(nouvellesImages, img)
		} else {
			imageTrouvee = true
		}
	}

	if !imageTrouvee {
		c.JSON(http.StatusNotFound, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "IMAGE_NON_TROUVEE",
				"message": "L'image n'existe pas dans ce cours",
			},
		})
		return
	}

	// Supprimer le fichier
	if err := h.serviceStorage.SupprimerImage(coursID, filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ERREUR_SUPPRESSION",
				"message": err.Error(),
			},
		})
		return
	}

	// Mettre à jour le cours
	cours.Images = nouvellesImages
	if err := h.coursRepo.MettreAJour(c.Request.Context(), cours); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ERREUR_MISE_A_JOUR",
				"message": "Erreur lors de la mise à jour du cours",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succes": true,
		"images": cours.Images,
	})
}

// ReordonnerImagesHandler réordonne les images d'un cours
// PUT /api/cours/:id/images/ordre
func (h *HandlersImages) ReordonnerImagesHandler(c *gin.Context) {
	if h.coursRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service n'est pas configuré",
			},
		})
		return
	}

	coursID := c.Param("id")
	if coursID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ID_MANQUANT",
				"message": "L'ID du cours est requis",
			},
		})
		return
	}

	var req struct {
		Images []string `json:"images" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "REQUETE_INVALIDE",
				"message": "La liste des images est requise",
			},
		})
		return
	}

	// Récupérer le cours
	cours, err := h.coursRepo.ObtenirParID(c.Request.Context(), coursID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "COURS_NON_TROUVE",
				"message": "Cours non trouvé",
			},
		})
		return
	}

	// Vérifier que toutes les images fournies existent dans le cours
	imagesExistantes := make(map[string]bool)
	for _, img := range cours.Images {
		imagesExistantes[img] = true
	}

	for _, img := range req.Images {
		if !imagesExistantes[img] {
			c.JSON(http.StatusBadRequest, gin.H{
				"succes": false,
				"erreur": gin.H{
					"code":    "IMAGE_INVALIDE",
					"message": "L'image " + img + " n'existe pas dans ce cours",
				},
			})
			return
		}
	}

	// Mettre à jour l'ordre
	cours.Images = req.Images
	if err := h.coursRepo.MettreAJour(c.Request.Context(), cours); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ERREUR_MISE_A_JOUR",
				"message": "Erreur lors de la mise à jour du cours",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succes": true,
		"images": cours.Images,
	})
}

// DeplacerImageHandler déplace une image vers le haut ou vers le bas
// POST /api/cours/:id/images/:filename/deplacer
func (h *HandlersImages) DeplacerImageHandler(c *gin.Context) {
	if h.coursRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service n'est pas configuré",
			},
		})
		return
	}

	coursID := c.Param("id")
	filename := c.Param("filename")

	if coursID == "" || filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "PARAMETRES_MANQUANTS",
				"message": "L'ID du cours et le nom du fichier sont requis",
			},
		})
		return
	}

	var req struct {
		Direction string `json:"direction" binding:"required"` // "haut" ou "bas"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "REQUETE_INVALIDE",
				"message": "La direction est requise (haut ou bas)",
			},
		})
		return
	}

	if req.Direction != "haut" && req.Direction != "bas" {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "DIRECTION_INVALIDE",
				"message": "La direction doit être 'haut' ou 'bas'",
			},
		})
		return
	}

	// Récupérer le cours
	cours, err := h.coursRepo.ObtenirParID(c.Request.Context(), coursID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "COURS_NON_TROUVE",
				"message": "Cours non trouvé",
			},
		})
		return
	}

	// Trouver l'index de l'image
	index := -1
	for i, img := range cours.Images {
		if img == filename {
			index = i
			break
		}
	}

	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "IMAGE_NON_TROUVEE",
				"message": "L'image n'existe pas dans ce cours",
			},
		})
		return
	}

	// Déplacer l'image
	newIndex := index
	if req.Direction == "haut" && index > 0 {
		newIndex = index - 1
	} else if req.Direction == "bas" && index < len(cours.Images)-1 {
		newIndex = index + 1
	}

	if newIndex != index {
		// Échanger les positions
		cours.Images[index], cours.Images[newIndex] = cours.Images[newIndex], cours.Images[index]

		if err := h.coursRepo.MettreAJour(c.Request.Context(), cours); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"succes": false,
				"erreur": gin.H{
					"code":    "ERREUR_MISE_A_JOUR",
					"message": "Erreur lors de la mise à jour du cours",
				},
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"succes": true,
		"images": cours.Images,
	})
}

// ListerImagesHandler retourne la liste des images d'un cours
// GET /api/cours/:id/images
func (h *HandlersImages) ListerImagesHandler(c *gin.Context) {
	if h.coursRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service n'est pas configuré",
			},
		})
		return
	}

	coursID := c.Param("id")
	if coursID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ID_MANQUANT",
				"message": "L'ID du cours est requis",
			},
		})
		return
	}

	// Récupérer le cours
	cours, err := h.coursRepo.ObtenirParID(c.Request.Context(), coursID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "COURS_NON_TROUVE",
				"message": "Cours non trouvé",
			},
		})
		return
	}

	// Construire les URLs des images
	baseURL := "/api/cours/" + coursID + "/images/"
	imagesAvecURL := make([]gin.H, len(cours.Images))
	for i, img := range cours.Images {
		imagesAvecURL[i] = gin.H{
			"nom": img,
			"url": baseURL + img,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"succes": true,
		"images": imagesAvecURL,
		"total":  len(cours.Images),
	})
}

// parseInt convertit une chaîne en entier
func parseIntImages(s string) (int, error) {
	return strconv.Atoi(s)
}
