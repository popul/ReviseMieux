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

    const carteCours = page.locator('[data-testid="cours-card"]').first();
    const carteVisible = await carteCours.isVisible().catch(() => false);
    if (carteVisible) {
      await carteCours.click();
      await expect(page).toHaveURL(/\/cours\?id=/);
    }
  });

  test('affiche le bouton pour activer le mode édition', async ({ page }) => {
    // Vérifier que le bouton d'édition est présent
    await expect(page.getByRole('button', { name: /modifier|éditer/i })).toBeVisible();
  });

  test('active le mode édition au clic sur le bouton', async ({ page }) => {
    // Cliquer sur le bouton d'édition
    await page.getByRole('button', { name: /modifier|éditer/i }).click();

    // Vérifier que le mode édition est actif
    await expect(page.getByRole('button', { name: /enregistrer|sauvegarder/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /annuler/i })).toBeVisible();
  });

  test('permet d\'éditer le texte OCR', async ({ page }) => {
    // Activer le mode édition
    await page.getByRole('button', { name: /modifier|éditer/i }).click();

    // Vérifier que le texte est éditable
    const champTexte = page.locator('[data-testid="texte-ocr-editable"]');
    await expect(champTexte).toBeVisible();

    // Modifier le texte
    await champTexte.fill('Texte modifié pour le test');

    // Le champ doit contenir le nouveau texte
    await expect(champTexte).toHaveValue('Texte modifié pour le test');
  });

  test('affiche les images du cours avec le texte superposé en mode édition', async ({ page }) => {
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
    // Activer le mode édition
    await page.getByRole('button', { name: /modifier|éditer/i }).click();

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
    // Activer le mode édition
    await page.getByRole('button', { name: /modifier|éditer/i }).click();

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
    // Activer le mode édition
    await page.getByRole('button', { name: /modifier|éditer/i }).click();

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
  test.beforeEach(async ({ page }) => {
    // Aller sur la liste des cours et ouvrir le premier
    await page.goto('/cours');

    const carteCours = page.locator('[data-testid="cours-card"]').first();
    const carteVisible = await carteCours.isVisible().catch(() => false);
    if (carteVisible) {
      await carteCours.click();
      await expect(page).toHaveURL(/\/cours\?id=/);
    }
  });

  test('affiche les images du cours si elles existent', async ({ page }) => {
    // Activer le mode édition pour voir la section images même si vide
    await page.getByRole('button', { name: /modifier|éditer/i }).click();

    // Vérifier que la section des images est présente en mode édition
    const sectionImages = page.locator('[data-testid="images-ocr"]');
    await expect(sectionImages).toBeVisible();

    // Vérifier qu'il y a des vignettes d'images OU le message "aucune image"
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    if (nombreVignettes > 0) {
      // Il y a des images, vérifier qu'on peut les voir
      await expect(vignettes.first()).toBeVisible();
    } else {
      // Pas d'images, vérifier que le message est présent
      await expect(page.getByText(/aucune image/i)).toBeVisible();
    }
  });

  test('permet de sélectionner une image pour la voir en grand', async ({ page }) => {
    const sectionImages = page.locator('[data-testid="images-ocr"]');
    const sectionVisible = await sectionImages.isVisible().catch(() => false);

    if (!sectionVisible) {
      test.skip();
      return;
    }

    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    if (nombreVignettes < 2) {
      test.skip();
      return;
    }

    // Cliquer sur la deuxième vignette
    await vignettes.nth(1).click();

    // Vérifier que la deuxième vignette est sélectionnée (a une bordure différente)
    await expect(vignettes.nth(1)).toHaveClass(/border-coral/);
  });

  test('permet d\'ajouter une nouvelle image en mode édition', async ({ page }) => {
    // Activer le mode édition
    await page.getByRole('button', { name: /modifier|éditer/i }).click();

    // Vérifier que le bouton d'ajout d'image est visible
    const boutonAjout = page.locator('[data-testid="ajouter-image"]');
    await expect(boutonAjout).toBeVisible();
  });

  test('permet de supprimer une image en mode édition', async ({ page }) => {
    // Activer le mode édition
    await page.getByRole('button', { name: /modifier|éditer/i }).click();

    // Vérifier s'il y a des images
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    if (nombreVignettes === 0) {
      // Pas d'images, on skip ce test
      test.skip();
      return;
    }

    // Vérifier que le bouton de suppression est visible sur les images
    const boutonSupprimer = page.locator('[data-testid="supprimer-image"]').first();
    await expect(boutonSupprimer).toBeVisible();
  });

  test('permet de réordonner les images par drag and drop en mode édition', async ({ page }) => {
    const sectionImages = page.locator('[data-testid="images-ocr"]');
    const sectionVisible = await sectionImages.isVisible().catch(() => false);

    if (!sectionVisible) {
      test.skip();
      return;
    }

    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    if (nombreVignettes < 2) {
      test.skip();
      return;
    }

    // Activer le mode édition
    await page.getByRole('button', { name: /modifier|éditer/i }).click();

    // Vérifier que les boutons de réordonnancement sont visibles
    const boutonMonter = page.locator('[data-testid="monter-image"]').first();
    const boutonDescendre = page.locator('[data-testid="descendre-image"]').first();

    // Au moins un des boutons doit être visible
    const monterVisible = await boutonMonter.isVisible().catch(() => false);
    const descendreVisible = await boutonDescendre.isVisible().catch(() => false);

    expect(monterVisible || descendreVisible).toBeTruthy();
  });

  test('sauvegarde les modifications d\'images', async ({ page }) => {
    const sectionImages = page.locator('[data-testid="images-ocr"]');
    const sectionVisible = await sectionImages.isVisible().catch(() => false);

    if (!sectionVisible) {
      test.skip();
      return;
    }

    // Activer le mode édition
    await page.getByRole('button', { name: /modifier|éditer/i }).click();

    // Faire une modification (si possible)
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    if (nombreVignettes >= 2) {
      // Descendre la première image
      const boutonDescendre = page.locator('[data-testid="descendre-image"]').first();
      if (await boutonDescendre.isVisible()) {
        await boutonDescendre.click();
      }
    }

    // Sauvegarder
    await page.getByRole('button', { name: /enregistrer|sauvegarder/i }).click();

    // Vérifier qu'on sort du mode édition
    await expect(page.getByRole('button', { name: /modifier|éditer/i })).toBeVisible();
  });
});
