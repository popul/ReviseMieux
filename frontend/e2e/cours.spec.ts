import { test, expect } from '@playwright/test';

/**
 * Tests E2E pour la page Cours
 * - Liste des cours cliquables
 * - Édition du contenu d'un cours (texte, images, zones incertaines)
 */

test.describe('Liste des cours', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/cours');
  });

  test('affiche la liste des cours', async ({ page }) => {
    // Vérifie le titre de la page
    await expect(page.getByRole('heading', { name: /mes cours/i })).toBeVisible();
  });

  test('cliquer sur une carte de cours ouvre le détail du cours', async ({ page }) => {
    // Attendre que les cours soient chargés
    const carteCours = page.locator('[data-testid="cours-card"]').first();

    // Si aucun cours, skip
    const carteVisible = await carteCours.isVisible().catch(() => false);
    if (!carteVisible) {
      test.skip();
      return;
    }

    // Cliquer sur la carte
    await carteCours.click();

    // Vérifier qu'on est sur la page de détail
    await expect(page).toHaveURL(/\/cours\?id=/);
    await expect(page.getByRole('heading', { name: /contenu du cours/i })).toBeVisible();
  });
});

test.describe('Détail et édition d\'un cours', () => {
  test.beforeEach(async ({ page }) => {
    // Aller sur la liste des cours et ouvrir le premier
    await page.goto('/cours');

    // Attendre que la page soit chargée
    await expect(page.getByRole('heading', { name: /mes cours/i })).toBeVisible({ timeout: 10000 });

    const carteCours = page.locator('[data-testid="cours-card"]').first();
    const carteVisible = await carteCours.isVisible().catch(() => false);
    if (carteVisible) {
      await carteCours.click();
      await expect(page).toHaveURL(/\/cours\?id=/);
      // Attendre que la page de détail soit complètement chargée
      await expect(page.getByRole('heading', { name: /contenu du cours/i })).toBeVisible({ timeout: 10000 });
    }
  });

  test('affiche le bouton pour activer le mode édition', async ({ page }) => {
    // Skip si pas sur une page de détail de cours
    const surPageDetail = page.url().includes('?id=');
    if (!surPageDetail) {
      test.skip();
      return;
    }
    // Vérifier que le bouton d'édition est présent
    await expect(page.getByRole('button', { name: /modifier|éditer/i })).toBeVisible({ timeout: 10000 });
  });

  test('active le mode édition au clic sur le bouton', async ({ page }) => {
    // Skip si pas sur une page de détail de cours
    if (!page.url().includes('?id=')) {
      test.skip();
      return;
    }
    // Attendre et cliquer sur le bouton d'édition
    const boutonModifier = page.getByRole('button', { name: /modifier|éditer/i });
    await expect(boutonModifier).toBeVisible({ timeout: 10000 });
    await boutonModifier.click();

    // Vérifier que le mode édition est actif
    await expect(page.getByRole('button', { name: /enregistrer|sauvegarder/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /annuler/i })).toBeVisible();
  });

  test('permet d\'éditer le texte OCR', async ({ page }) => {
    // Skip si pas sur une page de détail de cours
    if (!page.url().includes('?id=')) {
      test.skip();
      return;
    }
    // Attendre et activer le mode édition
    const boutonModifier = page.getByRole('button', { name: /modifier|éditer/i });
    await expect(boutonModifier).toBeVisible({ timeout: 10000 });
    await boutonModifier.click();

    // Vérifier que le texte est éditable
    const champTexte = page.locator('[data-testid="texte-ocr-editable"]');
    await expect(champTexte).toBeVisible();

    // Modifier le texte
    await champTexte.fill('Texte modifié pour le test');

    // Le champ doit contenir le nouveau texte
    await expect(champTexte).toHaveValue('Texte modifié pour le test');
  });

  test('affiche les images du cours avec le texte superposé en mode édition', async ({ page }) => {
    // Skip si pas sur une page de détail de cours
    if (!page.url().includes('?id=')) {
      test.skip();
      return;
    }
    // Vérifier si des images sont présentes
    const sectionImages = page.locator('[data-testid="images-ocr"]');
    const sectionVisible = await sectionImages.isVisible().catch(() => false);

    if (!sectionVisible) {
      // Pas d'images pour ce cours, skip le test
      test.skip();
      return;
    }

    // Vérifier qu'il y a au moins une image
    const images = sectionImages.locator('img');
    const nombreImages = await images.count();

    if (nombreImages === 0) {
      test.skip();
      return;
    }

    // Activer le mode édition pour voir le texte superposé
    await page.getByRole('button', { name: /modifier|éditer/i }).click();

    // Vérifier que le texte est superposé sur l'image
    const overlayTexte = page.locator('[data-testid="texte-overlay"]').first();
    await expect(overlayTexte).toBeVisible({ timeout: 5000 });
  });

  test('permet de corriger les zones incertaines', async ({ page }) => {
    // Skip si pas sur une page de détail de cours
    if (!page.url().includes('?id=')) {
      test.skip();
      return;
    }
    // Attendre que le bouton modifier soit visible
    const boutonModifier = page.getByRole('button', { name: /modifier|éditer/i });
    await expect(boutonModifier).toBeVisible({ timeout: 10000 });
    await boutonModifier.click();

    // Vérifier si des zones incertaines sont présentes
    const zonesIncertaines = page.locator('[data-testid="zone-incertaine"]');
    const nombreZones = await zonesIncertaines.count();

    if (nombreZones > 0) {
      // Cliquer sur une zone incertaine pour la corriger
      await zonesIncertaines.first().click();

      // Vérifier qu'un champ d'édition apparaît
      const champCorrection = page.locator('[data-testid="correction-zone"]');
      await expect(champCorrection).toBeVisible();
    }
  });

  test('sauvegarde les modifications', async ({ page }) => {
    // Skip si pas sur une page de détail de cours
    if (!page.url().includes('?id=')) {
      test.skip();
      return;
    }
    // Attendre et activer le mode édition
    const boutonModifier = page.getByRole('button', { name: /modifier|éditer/i });
    await expect(boutonModifier).toBeVisible({ timeout: 10000 });
    await boutonModifier.click();

    // Faire une modification
    const champTexte = page.locator('[data-testid="texte-ocr-editable"]');
    const texteOriginal = await champTexte.inputValue();
    await champTexte.fill(texteOriginal + ' - modifié');

    // Sauvegarder
    await page.getByRole('button', { name: /enregistrer|sauvegarder/i }).click();

    // Vérifier qu'on sort du mode édition
    await expect(page.getByRole('button', { name: /modifier|éditer/i })).toBeVisible();

    // Vérifier que le texte a été modifié
    await expect(page.getByText(/modifié/)).toBeVisible();
  });

  test('annule les modifications', async ({ page }) => {
    // Skip si pas sur une page de détail de cours
    if (!page.url().includes('?id=')) {
      test.skip();
      return;
    }
    // Attendre et activer le mode édition
    const boutonModifier = page.getByRole('button', { name: /modifier|éditer/i });
    await expect(boutonModifier).toBeVisible({ timeout: 10000 });
    await boutonModifier.click();

    // Faire une modification
    const champTexte = page.locator('[data-testid="texte-ocr-editable"]');
    const texteOriginal = await champTexte.inputValue();
    await champTexte.fill('Texte totalement différent');

    // Annuler
    await page.getByRole('button', { name: /annuler/i }).click();

    // Vérifier qu'on sort du mode édition sans sauvegarder
    await expect(page.getByRole('button', { name: /modifier|éditer/i })).toBeVisible();

    // Le texte original doit être affiché (pas "Texte totalement différent")
    await expect(page.getByText('Texte totalement différent')).not.toBeVisible();
  });
});

test.describe('Gestion des images du cours', () => {
  // Helper pour trouver un cours avec au moins N images
  async function trouverCoursAvecImages(page, nombreMinImages = 2) {
    await page.goto('/cours');

    const cartesCours = page.locator('[data-testid="cours-card"]');
    const nombreCours = await cartesCours.count();

    for (let i = 0; i < nombreCours; i++) {
      // Cliquer sur le cours
      await cartesCours.nth(i).click();
      await expect(page).toHaveURL(/\/cours\?id=/);

      // Vérifier s'il a assez d'images
      const vignettes = page.locator('[data-testid="image-vignette"]');
      const nombreVignettes = await vignettes.count();

      if (nombreVignettes >= nombreMinImages) {
        return true; // Trouvé un cours avec assez d'images
      }

      // Sinon, retourner à la liste
      await page.goto('/cours');
    }

    return false; // Aucun cours trouvé avec assez d'images
  }

  test.beforeEach(async ({ page }) => {
    // Chercher un cours avec au moins 2 images
    const coursAvecImages = await trouverCoursAvecImages(page, 2);
    if (!coursAvecImages) {
      // Si aucun cours avec 2 images, prendre le premier cours disponible
      await page.goto('/cours');
      const carteCours = page.locator('[data-testid="cours-card"]').first();
      const carteVisible = await carteCours.isVisible().catch(() => false);
      if (carteVisible) {
        await carteCours.click();
        await expect(page).toHaveURL(/\/cours\?id=/);
      }
    }
  });

  test('affiche les vignettes d\'images avec numéros de page', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    // Skip si pas d'images
    if (nombreVignettes === 0) {
      test.skip();
      return;
    }

    // Vérifier que la première vignette est visible
    await expect(vignettes.first()).toBeVisible();

    // Vérifier que le numéro de page 1 est affiché
    await expect(page.getByText('1', { exact: true }).first()).toBeVisible();

    // Si plus d'une image, vérifier le numéro 2
    if (nombreVignettes >= 2) {
      await expect(page.getByText('2', { exact: true }).first()).toBeVisible();
    }
  });

  test('affiche l\'indicateur de page actuelle', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    // Skip si pas assez d'images pour avoir un indicateur
    if (nombreVignettes < 2) {
      test.skip();
      return;
    }

    // Vérifier que l'indicateur "Page X sur Y" est visible (boutons de navigation)
    // OU le badge "Page X / Y" sur l'image
    const indicateurSur = page.getByText(/page \d+ sur \d+/i);
    const indicateurSlash = page.getByText(/page \d+ \/ \d+/i);

    // L'un ou l'autre doit être visible
    const surVisible = await indicateurSur.isVisible().catch(() => false);
    const slashVisible = await indicateurSlash.isVisible().catch(() => false);

    expect(surVisible || slashVisible).toBeTruthy();
  });

  test('permet de sélectionner une image pour la voir en grand', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    // Skip si pas assez d'images pour tester la sélection
    if (nombreVignettes < 2) {
      test.skip();
      return;
    }

    // Vérifier que la première vignette est sélectionnée par défaut
    await expect(vignettes.first()).toHaveClass(/border-coral/);

    // Cliquer sur la deuxième vignette
    await vignettes.nth(1).click();

    // Vérifier que la deuxième vignette est maintenant sélectionnée
    await expect(vignettes.nth(1)).toHaveClass(/border-coral/);

    // Vérifier que l'indicateur de page a changé
    await expect(page.getByText(/page 2 sur/i)).toBeVisible();
  });

  test('permet de naviguer avec les boutons Précédent/Suivant', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    // Skip si pas assez d'images pour tester la navigation
    if (nombreVignettes < 2) {
      test.skip();
      return;
    }

    // Vérifier que les boutons de navigation sont présents
    const boutonPrecedent = page.getByRole('button', { name: /précédent/i });
    const boutonSuivant = page.getByRole('button', { name: /suivant/i });

    await expect(boutonPrecedent).toBeVisible();
    await expect(boutonSuivant).toBeVisible();

    // Le bouton Précédent doit être désactivé sur la première page
    await expect(boutonPrecedent).toBeDisabled();
    await expect(boutonSuivant).toBeEnabled();

    // Cliquer sur Suivant
    await boutonSuivant.click();

    // Vérifier que la page a changé
    await expect(page.getByText(/page 2 sur/i)).toBeVisible();

    // Vérifier que la deuxième vignette est sélectionnée
    await expect(vignettes.nth(1)).toHaveClass(/border-coral/);

    // Maintenant Précédent doit être activé
    await expect(boutonPrecedent).toBeEnabled();
  });

  test('le bouton Suivant est désactivé sur la dernière page', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    // Skip si pas assez d'images pour tester la navigation
    if (nombreVignettes < 2) {
      test.skip();
      return;
    }

    // Aller à la dernière page en cliquant sur la dernière vignette
    await vignettes.last().click();

    // Vérifier que le bouton Suivant est désactivé
    const boutonSuivant = page.getByRole('button', { name: /suivant/i });
    await expect(boutonSuivant).toBeDisabled();

    // Le bouton Précédent doit être activé
    const boutonPrecedent = page.getByRole('button', { name: /précédent/i });
    await expect(boutonPrecedent).toBeEnabled();
  });

  test('affiche l\'image principale correspondant à la sélection', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    // Skip si pas d'images
    if (nombreVignettes === 0) {
      test.skip();
      return;
    }

    // Vérifier qu'une image principale est affichée
    const imagePrincipale = page.locator('[data-testid="images-ocr"] img').last();
    await expect(imagePrincipale).toBeVisible();

    // Si plusieurs images, vérifier le badge "Page X / Y"
    if (nombreVignettes >= 2) {
      await expect(page.getByText(/page \d+ \/ \d+/i)).toBeVisible();
    }
  });

  test('affiche le message de correspondance texte-image', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    // Skip si pas d'images
    if (nombreVignettes === 0) {
      test.skip();
      return;
    }

    // Attendre que la section images soit visible
    const sectionImages = page.locator('[data-testid="images-ocr"]');
    await expect(sectionImages).toBeVisible({ timeout: 10000 });

    // Vérifier que le message de correspondance est affiché
    await expect(page.getByText(/le texte ci-dessous correspond/i)).toBeVisible();
  });

  test('permet d\'ajouter une nouvelle image en mode édition', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    // Skip si pas d'images (le bouton d'ajout n'est visible que s'il y a déjà des images)
    if (nombreVignettes === 0) {
      test.skip();
      return;
    }

    // Attendre que le bouton modifier soit visible
    const boutonModifier = page.getByRole('button', { name: /modifier/i });
    await expect(boutonModifier).toBeVisible({ timeout: 10000 });

    // Activer le mode édition
    await boutonModifier.click();

    // Vérifier que le bouton d'ajout d'image est visible
    const boutonAjout = page.locator('[data-testid="ajouter-image"]');
    await expect(boutonAjout).toBeVisible();
  });

  test('permet de supprimer une image en mode édition', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    // Skip si pas d'images
    if (nombreVignettes === 0) {
      test.skip();
      return;
    }

    // Attendre que le bouton modifier soit visible
    const boutonModifier = page.getByRole('button', { name: /modifier/i });
    await expect(boutonModifier).toBeVisible({ timeout: 10000 });

    // Activer le mode édition
    await boutonModifier.click();

    // Vérifier que le bouton de suppression est visible sur les images
    const boutonSupprimer = page.locator('[data-testid="supprimer-image"]').first();
    await expect(boutonSupprimer).toBeVisible();
  });

  test('les boutons de navigation changent la page sélectionnée', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    // Skip si pas assez d'images pour tester la navigation
    if (nombreVignettes < 2) {
      test.skip();
      return;
    }

    // Vérifier qu'on est sur la page 1
    await expect(page.getByText(/page 1 sur/i)).toBeVisible();

    // Cliquer sur "Suivant"
    const boutonSuivant = page.getByRole('button', { name: /suivant/i });
    await boutonSuivant.click();

    // Vérifier que la sélection a changé (on est sur la page 2)
    await expect(page.getByText(/page 2 sur/i)).toBeVisible();

    // Cliquer sur "Précédent" pour revenir
    const boutonPrecedent = page.getByRole('button', { name: /précédent/i });
    await boutonPrecedent.click();

    // Vérifier qu'on est revenu à la page 1
    await expect(page.getByText(/page 1 sur/i)).toBeVisible();
  });

  test('sauvegarde les modifications d\'images en mode édition', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    // Skip si pas d'images
    if (nombreVignettes === 0) {
      test.skip();
      return;
    }

    // Attendre que le bouton modifier soit visible
    const boutonModifier = page.getByRole('button', { name: /modifier/i });
    await expect(boutonModifier).toBeVisible({ timeout: 10000 });

    // Activer le mode édition
    await boutonModifier.click();

    // Sauvegarder
    await page.getByRole('button', { name: /enregistrer/i }).click();

    // Vérifier qu'on sort du mode édition
    await expect(page.getByRole('button', { name: /modifier/i })).toBeVisible();
  });
});
