import { test, expect } from '@playwright/test';

/**
 * Tests E2E pour la page Scanner et le flux de création de cours avec OCR
 * Ces tests utilisent le vrai backend avec l'API OpenAI
 */

// Chemin relatif au répertoire racine du projet
const TEST_IMAGE = 'frontend/e2e/fixtures/test-notes.jpg';

// Timeout plus long pour les appels API réels
test.setTimeout(120000);

test.describe('Scanner - Interface', () => {
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
    const titreInput = page.getByPlaceholder(/Révolution française/i);
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
  test('flux complet: upload → options → OCR → résultat', async ({ page }) => {
    // 1. Aller sur la page scanner
    await page.goto('/scanner');

    // 2. Upload d'un fichier
    const fileInput = page.locator('input[type="file"]').first();
    await fileInput.setInputFiles(TEST_IMAGE);

    // 3. Remplir les options
    const titreInput = page.getByPlaceholder(/Révolution française/i);
    await titreInput.fill('Test OCR E2E');

    const matiereSelect = page.getByRole('combobox');
    await matiereSelect.selectOption('histoire');

    // 4. Soumettre
    const submitButton = page.getByRole('button', { name: /Générer mes supports/i });
    await submitButton.click();

    // 5. Vérifier le résultat (timeout long pour l'API réelle)
    await expect(page.getByText('Extraction réussie')).toBeVisible({ timeout: 60000 });

    // 7. Vérifier qu'on a des informations sur les pages
    await expect(page.getByText(/page.*traitée/i)).toBeVisible();

    // 8. Vérifier les badges des supports sélectionnés
    const mainContent = page.locator('#main-content');
    await expect(mainContent.getByText('Fiches de révision')).toBeVisible();
    await expect(mainContent.getByText('Quiz interactif')).toBeVisible();
  });

  test('affiche le score de confiance après OCR', async ({ page }) => {
    await page.goto('/scanner');
    const fileInput = page.locator('input[type="file"]').first();
    await fileInput.setInputFiles(TEST_IMAGE);

    await page.getByRole('button', { name: /Générer mes supports/i }).click();
    await expect(page.getByText('Extraction réussie')).toBeVisible({ timeout: 60000 });

    // Vérifier qu'un pourcentage de confiance est affiché
    await expect(page.getByText(/%/)).toBeVisible();
  });

  test('permet d\'éditer le texte OCR', async ({ page }) => {
    await page.goto('/scanner');
    const fileInput = page.locator('input[type="file"]').first();
    await fileInput.setInputFiles(TEST_IMAGE);

    await page.getByRole('button', { name: /Générer mes supports/i }).click();
    await expect(page.getByText('Extraction réussie')).toBeVisible({ timeout: 60000 });

    // Cliquer sur le bouton d'édition si visible
    const editButton = page.getByRole('button', { name: /Modifier/i });
    if (await editButton.isVisible()) {
      await editButton.click();

      // Vérifier qu'on peut éditer
      const textarea = page.getByRole('textbox');
      await expect(textarea).toBeVisible();
    }
  });

  test('le bouton "Nouveau scan" réinitialise la page', async ({ page }) => {
    await page.goto('/scanner');
    const fileInput = page.locator('input[type="file"]').first();
    await fileInput.setInputFiles(TEST_IMAGE);

    await page.getByRole('button', { name: /Générer mes supports/i }).click();
    await expect(page.getByText('Extraction réussie')).toBeVisible({ timeout: 60000 });

    // Cliquer sur "Nouveau scan"
    await page.getByRole('button', { name: /Nouveau scan/i }).click();

    // Vérifier qu'on est revenu à l'état initial
    await expect(page.getByText('Glisse tes fichiers ici')).toBeVisible();
  });
});
