import { test, expect } from '@playwright/test';

/**
 * Tests E2E pour la page Scanner
 * Vérifie l'interface d'upload et les options de génération
 */

test.describe('Scanner', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/scanner');
  });

  test('charge la page scanner', async ({ page }) => {
    // Vérifie qu'on est sur la bonne URL
    await expect(page).toHaveURL(/.*scanner/);

    // Le scanner a un indicateur d'étapes avec "Import" (exact match)
    await expect(page.getByText('Import', { exact: true })).toBeVisible();
  });

  test('navigation clavier fonctionne', async ({ page }) => {
    // Vérifie que la page est navigable au clavier
    await page.keyboard.press('Tab');

    const focusedElement = page.locator(':focus');
    await expect(focusedElement).toBeVisible();
  });
});
