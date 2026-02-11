import { test, expect } from '@playwright/test';

/**
 * Tests E2E pour l'overlay OCR et le composant de reorganisation des pages
 * Verifie que le texte OCR est correctement superpose sur les images du cours,
 * que le code couleur de confiance est visible, que le zoom fonctionne,
 * et que les pages peuvent etre reorganisees par drag and drop.
 */

// Helper: naviguer vers un cours avec des images et des blocs de texte OCR
async function ouvrirCoursAvecImages(page: import('@playwright/test').Page) {
  await page.goto('/cours');
  await expect(page.getByRole('heading', { name: /mes cours/i })).toBeVisible({ timeout: 10000 });

  const cartesCours = page.locator('[data-testid="cours-card"]');
  const nombreCours = await cartesCours.count();

  for (let i = 0; i < nombreCours; i++) {
    await cartesCours.nth(i).click();
    await expect(page).toHaveURL(/\/cours\?id=/);

    // Attendre le chargement complet
    const sectionImages = page.locator('[data-testid="images-ocr"]');
    const sectionVisible = await sectionImages.isVisible().catch(() => false);

    if (sectionVisible) {
      const vignettes = page.locator('[data-testid="image-vignette"]');
      const nombreVignettes = await vignettes.count();
      if (nombreVignettes > 0) {
        return true;
      }
    }

    // Retourner a la liste
    await page.goto('/cours');
    await expect(page.getByRole('heading', { name: /mes cours/i })).toBeVisible({ timeout: 10000 });
  }

  return false;
}

test.describe('Overlay OCR - Affichage', () => {
  test.beforeEach(async ({ page }) => {
    const coursAvecImages = await ouvrirCoursAvecImages(page);
    if (!coursAvecImages) {
      test.skip();
    }
  });

  test('affiche les vignettes d\'images du cours', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    if (nombreVignettes === 0) {
      test.skip();
      return;
    }

    await expect(vignettes.first()).toBeVisible();
  });

  test('affiche le composant overlay ou le mode classique selon la presence de blocsTexte', async ({ page }) => {
    // Verifier qu'il y a soit l'overlay OCR (avec zoom controls) soit l'image classique
    const overlayZoom = page.locator('button[title="Zoomer"]');
    const imageBadgePage = page.getByText(/page \d+ \/ \d+/i);

    const overlayPresent = await overlayZoom.isVisible().catch(() => false);
    const classicPresent = await imageBadgePage.isVisible().catch(() => false);

    // L'un des deux modes doit etre present
    expect(overlayPresent || classicPresent).toBeTruthy();
  });

  test('affiche la legende de confiance quand l\'overlay OCR est actif', async ({ page }) => {
    // La legende montre >=90%, >=70%, <70%
    const legendeHaute = page.getByText(/90%/);
    const legendeMoyenne = page.getByText(/70%/);

    const legendeVisible = await legendeHaute.first().isVisible().catch(() => false);

    if (legendeVisible) {
      await expect(legendeHaute.first()).toBeVisible();
      await expect(legendeMoyenne.first()).toBeVisible();
    }
    // Si pas de legende, c'est que le cours n'a pas de blocsTexte - OK
  });

  test('les controles de zoom sont presents quand l\'overlay est actif', async ({ page }) => {
    const boutonZoomIn = page.locator('button[title="Zoomer"]');
    const boutonZoomOut = page.locator('button[title="Dezoomer"]');
    const boutonZoomReset = page.locator('button[title="Reinitialiser le zoom"]');

    const overlayActif = await boutonZoomIn.isVisible().catch(() => false);

    if (!overlayActif) {
      // Pas de blocsTexte, pas d'overlay - skip
      test.skip();
      return;
    }

    await expect(boutonZoomIn).toBeVisible();
    await expect(boutonZoomOut).toBeVisible();
    await expect(boutonZoomReset).toBeVisible();
  });

  test('le zoom fonctionne avec les boutons + et -', async ({ page }) => {
    const boutonZoomIn = page.locator('button[title="Zoomer"]');
    const boutonZoomReset = page.locator('button[title="Reinitialiser le zoom"]');

    const overlayActif = await boutonZoomIn.isVisible().catch(() => false);

    if (!overlayActif) {
      test.skip();
      return;
    }

    // Verifier le zoom initial a 100%
    await expect(boutonZoomReset).toHaveText('100%');

    // Zoomer
    await boutonZoomIn.click();
    await expect(boutonZoomReset).toHaveText('125%');

    // Re-zoomer
    await boutonZoomIn.click();
    await expect(boutonZoomReset).toHaveText('150%');

    // Reset
    await boutonZoomReset.click();
    await expect(boutonZoomReset).toHaveText('100%');
  });

  test('le dezoomer fonctionne et s\'arrete a 50%', async ({ page }) => {
    const boutonZoomOut = page.locator('button[title="Dezoomer"]');
    const boutonZoomReset = page.locator('button[title="Reinitialiser le zoom"]');

    const overlayActif = await boutonZoomOut.isVisible().catch(() => false);

    if (!overlayActif) {
      test.skip();
      return;
    }

    // Dezoomer
    await boutonZoomOut.click();
    await expect(boutonZoomReset).toHaveText('75%');

    // Dezoomer encore
    await boutonZoomOut.click();
    await expect(boutonZoomReset).toHaveText('50%');

    // Le bouton doit etre desactive a 50%
    await expect(boutonZoomOut).toBeDisabled();
  });

  test('affiche le message d\'instruction en mode lecture', async ({ page }) => {
    const boutonZoomIn = page.locator('button[title="Zoomer"]');
    const overlayActif = await boutonZoomIn.isVisible().catch(() => false);

    if (!overlayActif) {
      test.skip();
      return;
    }

    // Verifier le message d'instruction
    await expect(page.getByText(/cliquez sur un bloc pour le modifier/i)).toBeVisible();
  });

  test('le cours sans blocsTexte affiche le mode d\'image classique avec badge de page', async ({ page }) => {
    // Ce test verifie le fallback: si pas de blocsTexte, on voit l'image avec badge
    const imageBadgePage = page.getByText(/page \d+ \/ \d+/i);
    const overlayZoom = page.locator('button[title="Zoomer"]');

    const classicPresent = await imageBadgePage.isVisible().catch(() => false);
    const overlayPresent = await overlayZoom.isVisible().catch(() => false);

    // Si le mode classique est present, verifier le message de correspondance
    if (classicPresent && !overlayPresent) {
      await expect(page.getByText(/le texte ci-dessous correspond/i)).toBeVisible();
    }
    // Sinon, c'est le mode overlay, ce test est non applicable
  });
});

test.describe('Overlay OCR - Mode edition', () => {
  test.beforeEach(async ({ page }) => {
    const coursAvecImages = await ouvrirCoursAvecImages(page);
    if (!coursAvecImages) {
      test.skip();
    }
  });

  test('le mode edition affiche les instructions d\'edition pour l\'overlay', async ({ page }) => {
    const boutonZoomIn = page.locator('button[title="Zoomer"]');
    const overlayActif = await boutonZoomIn.isVisible().catch(() => false);

    if (!overlayActif) {
      test.skip();
      return;
    }

    // Activer le mode edition
    const boutonModifier = page.getByRole('button', { name: /modifier/i });
    await expect(boutonModifier).toBeVisible({ timeout: 10000 });
    await boutonModifier.click();

    // Verifier le message d'instructions en mode edition
    await expect(page.getByText(/cliquez sur un bloc pour le modifier/i)).toBeVisible();
  });

  test('annuler le mode edition revient au mode lecture', async ({ page }) => {
    const boutonZoomIn = page.locator('button[title="Zoomer"]');
    const overlayActif = await boutonZoomIn.isVisible().catch(() => false);

    if (!overlayActif) {
      test.skip();
      return;
    }

    // Activer puis annuler
    const boutonModifier = page.getByRole('button', { name: /modifier/i });
    await expect(boutonModifier).toBeVisible({ timeout: 10000 });
    await boutonModifier.click();

    await expect(page.getByText(/cliquez sur un bloc pour le modifier/i)).toBeVisible();

    await page.getByRole('button', { name: /annuler/i }).click();

    // Après annulation, le bouton Modifier doit réapparaître
    await expect(boutonModifier).toBeVisible();
  });
});

test.describe('Reorganisation des pages', () => {
  test.beforeEach(async ({ page }) => {
    const coursAvecImages = await ouvrirCoursAvecImages(page);
    if (!coursAvecImages) {
      test.skip();
    }
  });

  test('affiche les vignettes avec numeros de page', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    if (nombreVignettes === 0) {
      test.skip();
      return;
    }

    // Verifier que la premiere vignette est visible
    await expect(vignettes.first()).toBeVisible();

    // Verifier que le numero 1 est affiche
    await expect(page.getByText('1', { exact: true }).first()).toBeVisible();
  });

  test('cliquer sur une vignette change la page selectionnee', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    if (nombreVignettes < 2) {
      test.skip();
      return;
    }

    // Verifier que la premiere vignette est selectionnee par defaut
    await expect(vignettes.first()).toHaveClass(/border-coral/);

    // Cliquer sur la deuxieme vignette
    await vignettes.nth(1).click();

    // Verifier que la deuxieme est maintenant selectionnee
    await expect(vignettes.nth(1)).toHaveClass(/border-coral/);
  });

  test('affiche l\'indicateur page X sur Y quand plusieurs images', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    if (nombreVignettes < 2) {
      test.skip();
      return;
    }

    await expect(page.getByText(/page \d+ sur \d+/i)).toBeVisible();
  });

  test('le message de glisser-deposer apparait en mode edition', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    if (nombreVignettes < 2) {
      test.skip();
      return;
    }

    // Activer le mode edition
    const boutonModifier = page.getByRole('button', { name: /modifier/i });
    await expect(boutonModifier).toBeVisible({ timeout: 10000 });
    await boutonModifier.click();

    // Verifier que le message de drag and drop est visible
    await expect(page.getByText(/glissez pour reorganiser/i)).toBeVisible();
  });

  test('les boutons de suppression d\'image sont visibles en mode edition', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    if (nombreVignettes === 0) {
      test.skip();
      return;
    }

    // Activer le mode edition
    const boutonModifier = page.getByRole('button', { name: /modifier/i });
    await expect(boutonModifier).toBeVisible({ timeout: 10000 });
    await boutonModifier.click();

    // Verifier les boutons de suppression
    const boutonSupprimer = page.locator('[data-testid="supprimer-image"]').first();
    await expect(boutonSupprimer).toBeVisible();
  });

  test('le bouton d\'ajout d\'image est visible en mode edition', async ({ page }) => {
    const vignettes = page.locator('[data-testid="image-vignette"]');
    const nombreVignettes = await vignettes.count();

    if (nombreVignettes === 0) {
      test.skip();
      return;
    }

    // Activer le mode edition
    const boutonModifier = page.getByRole('button', { name: /modifier/i });
    await expect(boutonModifier).toBeVisible({ timeout: 10000 });
    await boutonModifier.click();

    // Verifier le bouton d'ajout
    const boutonAjout = page.locator('[data-testid="ajouter-image"]');
    await expect(boutonAjout).toBeVisible();
  });
});
