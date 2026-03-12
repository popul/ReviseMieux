import { test, expect } from '@playwright/test';
import path from 'path';
import { fileURLToPath } from 'url';

/**
 * Tests E2E pour la page Scanner et le flux de création de cours avec OCR
 * Ces tests utilisent le vrai backend avec l'API OpenAI
 */

// Chemin absolu vers le fichier de test
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const TEST_IMAGE = path.join(__dirname, 'fixtures/test-notes.jpg');
const TEST_IMAGE_2 = path.join(__dirname, 'fixtures/test-notes-2.jpg');
const TEST_IMAGE_3 = path.join(__dirname, 'fixtures/test-notes-3.jpg');

test.describe('Scanner - Interface', () => {
  // Timeout plus long pour les appels API réels
  test.describe.configure({ timeout: 120000 });
  test.beforeEach(async ({ page }) => {
    await page.goto('/scanner');
  });

  test('charge la page scanner', async ({ page }) => {
    await expect(page).toHaveURL(/.*scanner/);
    await expect(page.getByText('Import', { exact: true })).toBeVisible();
  });

  test('navigation clavier fonctionne', async ({ page }) => {
    await expect(page.getByText('Glisse tes fichiers ici')).toBeVisible();
    await page.keyboard.press('Tab');
    await page.keyboard.press('Tab');
    const focusedElement = page.locator(':focus');
    await expect(focusedElement).toBeAttached();
  });

  test('affiche la zone de téléchargement', async ({ page }) => {
    await expect(page.getByText('Glisse tes fichiers ici')).toBeVisible();
    await expect(page.getByRole('button', { name: /Choisir des fichiers/i })).toBeVisible();
  });
});

test.describe('Scanner - Upload de fichiers', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/scanner');
  });

  test('permet de sélectionner un fichier', async ({ page }) => {
    const fileInput = page.locator('input[type="file"]').first();
    await fileInput.setInputFiles(TEST_IMAGE);

    await expect(page.getByText('Options')).toBeVisible();
    await expect(page.getByText('Personnalise ta génération')).toBeVisible();
  });

  test('affiche le preview après upload', async ({ page }) => {
    const fileInput = page.locator('input[type="file"]').first();
    await fileInput.setInputFiles(TEST_IMAGE);

    await expect(page.getByText('test-notes.jpg')).toBeVisible();
  });

  test('permet de supprimer un fichier uploadé', async ({ page }) => {
    const fileInput = page.locator('input[type="file"]').first();
    await fileInput.setInputFiles(TEST_IMAGE);

    await expect(page.getByText('test-notes.jpg')).toBeVisible();

    const deleteButton = page.getByRole('button', { name: /supprimer/i }).first();
    if (await deleteButton.isVisible()) {
      await deleteButton.click();
      await expect(page.getByText('test-notes.jpg')).not.toBeVisible();
    }
  });
});

test.describe('Scanner - Options de génération', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/scanner');
    const fileInput = page.locator('input[type="file"]').first();
    await fileInput.setInputFiles(TEST_IMAGE);
  });

  test('affiche les options de génération après upload', async ({ page }) => {
    const mainContent = page.locator('#main-content');
    await expect(mainContent.getByText('Personnalise ta génération')).toBeVisible();
    await expect(mainContent.getByText('Fiches de révision')).toBeVisible();
    await expect(mainContent.getByText('🎯')).toBeVisible();
    await expect(mainContent.getByText('Carte mentale')).toBeVisible();
  });

  test('permet de saisir un titre', async ({ page }) => {
    const titreInput = page.getByPlaceholder(/laisser vide pour détection automatique/i);
    await titreInput.fill('Mon cours de test');
    await expect(titreInput).toHaveValue('Mon cours de test');
  });

  test('permet de sélectionner une matière', async ({ page }) => {
    const matiereSelect = page.getByRole('combobox');
    await matiereSelect.selectOption('histoire');
    await expect(matiereSelect).toHaveValue('histoire');
  });

  test('le bouton de soumission est actif quand au moins une option est cochée', async ({ page }) => {
    const submitButton = page.getByRole('button', { name: /Générer mes supports/i });
    await expect(submitButton).toBeEnabled();
  });

  test('le bouton de soumission est désactivé si aucune option n\'est cochée', async ({ page }) => {
    const mainContent = page.locator('#main-content');

    const fichesLabel = mainContent.getByText('Fiches de révision');
    const quizLabel = mainContent.locator('label').filter({ hasText: '🎯' });

    await fichesLabel.click();
    await quizLabel.click();

    const submitButton = page.getByRole('button', { name: /Générer mes supports/i });
    await expect(submitButton).toBeDisabled();

    await expect(page.getByText(/Sélectionne au moins un type/i)).toBeVisible();
  });
});

test.describe('Scanner - Flux OCR complet (API réelle)', () => {
  // Ces tests nécessitent un backend fonctionnel avec l'API OCR
  test.describe.configure({ timeout: 120000 });
  test.skip(({ }) => !process.env.RUN_API_TESTS, 'Skipped: Backend API non disponible. Lancez avec RUN_API_TESTS=1 pour activer.');

  test('flux complet: upload → options → OCR → redirection vers le cours', async ({ page }) => {
    await page.goto('/scanner');

    // Upload d'un fichier
    const fileInput = page.locator('input[type="file"]').first();
    await fileInput.setInputFiles(TEST_IMAGE);

    // Remplir les options
    const titreInput = page.getByPlaceholder(/laisser vide pour détection automatique/i);
    await titreInput.fill('Test OCR E2E');

    const matiereSelect = page.getByRole('combobox');
    await matiereSelect.selectOption('histoire');

    // Soumettre
    await page.getByRole('button', { name: /Générer mes supports/i }).click();

    // Le scanner redirige vers la page du cours cree
    await expect(page).toHaveURL(/\/cours\?id=/, { timeout: 90000 });

    // La page du cours doit afficher le contenu
    await expect(page.getByRole('heading', { name: /contenu du cours/i })).toBeVisible({ timeout: 10000 });
  });
});

test.describe('Scanner - Indicateurs de progression OCR (API réelle)', () => {
  // Ces tests nécessitent un backend fonctionnel avec l'API OCR
  // Timeout élevé : l'OCR de 3 pages prend ~90s+
  test.describe.configure({ timeout: 300000 });
  test.skip(({ }) => !process.env.RUN_API_TESTS, 'Skipped: Backend API non disponible. Lancez avec RUN_API_TESTS=1 pour activer.');

  test('upload 3 pages: affiche la bannière de progression OCR', async ({ page }) => {
    await page.goto('/scanner');

    // Upload de 3 fichiers différents
    const fileInput = page.locator('input[type="file"]').first();
    await fileInput.setInputFiles([TEST_IMAGE, TEST_IMAGE_2, TEST_IMAGE_3]);

    // Soumettre
    await page.getByRole('button', { name: /Générer mes supports/i }).click();

    // Redirection vers la page du cours
    await expect(page).toHaveURL(/\/cours\?id=/, { timeout: 30000 });

    // La bannière de progression OCR doit apparaître
    const banniere = page.locator('[data-testid="ocr-progress-banner"]');
    await expect(banniere).toBeVisible({ timeout: 10000 });

    // Elle doit contenir le texte de progression
    await expect(banniere.getByText(/extraction du texte en cours/i)).toBeVisible();
    await expect(banniere.getByText(/sur 3/)).toBeVisible();
  });

  test('upload 3 pages: les vignettes affichent les indicateurs OCR', async ({ page }) => {
    await page.goto('/scanner');

    // Upload de 3 fichiers
    const fileInput = page.locator('input[type="file"]').first();
    await fileInput.setInputFiles([TEST_IMAGE, TEST_IMAGE_2, TEST_IMAGE_3]);

    // Soumettre
    await page.getByRole('button', { name: /Générer mes supports/i }).click();

    // Redirection vers le cours
    await expect(page).toHaveURL(/\/cours\?id=/, { timeout: 30000 });

    // Attendre que les vignettes apparaissent
    const vignettes = page.locator('[data-testid="image-vignette"]');
    await expect(vignettes).toHaveCount(3, { timeout: 10000 });

    // Vérifier que les indicateurs OCR sont présents sur les vignettes
    const indicateurInProgress = page.locator('[data-testid="ocr-page-in-progress"]');
    const indicateurWaiting = page.locator('[data-testid="ocr-page-waiting"]');
    const indicateurDone = page.locator('[data-testid="ocr-page-done"]');

    // Attendre qu'au moins un indicateur apparaisse (polling met à jour les statuts)
    await page.waitForSelector(
      '[data-testid="ocr-page-in-progress"], [data-testid="ocr-page-waiting"], [data-testid="ocr-page-done"]',
      { timeout: 10000 }
    );

    // Le total des indicateurs doit être 3 (un pour chaque page)
    const totalIndicateurs = await indicateurDone.count() + await indicateurInProgress.count() + await indicateurWaiting.count();
    expect(totalIndicateurs).toBe(3);

    // Au maximum 1 spinner actif à la fois
    const nbSpinners = await indicateurInProgress.count();
    expect(nbSpinners).toBeLessThanOrEqual(1);
  });

  test('upload 3 pages: les indicateurs évoluent au fur et à mesure', async ({ page }) => {
    await page.goto('/scanner');

    // Upload de 3 fichiers
    const fileInput = page.locator('input[type="file"]').first();
    await fileInput.setInputFiles([TEST_IMAGE, TEST_IMAGE_2, TEST_IMAGE_3]);

    // Soumettre
    await page.getByRole('button', { name: /Générer mes supports/i }).click();

    // Redirection vers le cours
    await expect(page).toHaveURL(/\/cours\?id=/, { timeout: 30000 });

    // Attendre que les vignettes et les indicateurs apparaissent
    await expect(page.locator('[data-testid="image-vignette"]')).toHaveCount(3, { timeout: 10000 });

    const indicateurDone = page.locator('[data-testid="ocr-page-done"]');
    const indicateurInProgress = page.locator('[data-testid="ocr-page-in-progress"]');

    // Phase 1 : il ne doit y avoir maximum 1 spinner (une seule page à la fois est "en cours")
    const nbSpinners = await indicateurInProgress.count();
    expect(nbSpinners).toBeLessThanOrEqual(1);

    // Phase 2 : attendre qu'au moins une page soit terminée (check vert)
    await expect(indicateurDone.first()).toBeVisible({ timeout: 120000 });

    // Quand une page est terminée, le spinner doit avoir avancé
    const nbDone = await indicateurDone.count();
    expect(nbDone).toBeGreaterThanOrEqual(1);
  });

  test('upload 3 pages: tous les indicateurs disparaissent quand OCR terminé', async ({ page }) => {
    await page.goto('/scanner');

    // Upload de 3 fichiers
    const fileInput = page.locator('input[type="file"]').first();
    await fileInput.setInputFiles([TEST_IMAGE, TEST_IMAGE_2, TEST_IMAGE_3]);

    // Soumettre
    await page.getByRole('button', { name: /Générer mes supports/i }).click();

    // Redirection vers le cours
    await expect(page).toHaveURL(/\/cours\?id=/, { timeout: 30000 });

    // Attendre que les vignettes apparaissent
    await expect(page.locator('[data-testid="image-vignette"]')).toHaveCount(3, { timeout: 10000 });

    // Attendre que la bannière de progression disparaisse (OCR terminé)
    const banniere = page.locator('[data-testid="ocr-progress-banner"]');
    await expect(banniere).not.toBeVisible({ timeout: 240000 });

    // Après OCR terminé : plus aucun indicateur de progression sur les vignettes
    await expect(page.locator('[data-testid="ocr-page-in-progress"]')).toHaveCount(0);
    await expect(page.locator('[data-testid="ocr-page-waiting"]')).toHaveCount(0);
    await expect(page.locator('[data-testid="ocr-page-done"]')).toHaveCount(0);

    // Le bouton Modifier doit être visible (OCR terminé)
    await expect(page.getByRole('button', { name: /modifier/i })).toBeVisible({ timeout: 5000 });

    // Le texte OCR doit avoir du contenu
    const contenuCours = page.locator('pre');
    await expect(contenuCours).not.toBeEmpty();
  });
});
