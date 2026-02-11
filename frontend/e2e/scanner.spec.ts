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
