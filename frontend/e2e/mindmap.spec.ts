import { test, expect } from '@playwright/test';

/**
 * Tests E2E pour la page Mindmap
 * Vérifie la visualisation des cartes mentales
 */

test.describe('Mindmap', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/mindmap');
  });

  test('charge la page mindmap', async ({ page }) => {
    // Vérifie qu'on est sur la bonne URL
    await expect(page).toHaveURL(/.*mindmap/);

    // Attend que la page soit chargée (any h1 is attached to DOM)
    await expect(page.locator('h1').first()).toBeAttached();
  });

  test('navigation clavier fonctionne', async ({ page }) => {
    // Vérifie que la page est navigable au clavier
    await page.keyboard.press('Tab');

    const focusedElement = page.locator(':focus');
    await expect(focusedElement).toBeVisible();
  });
});

test.describe('Mindmap - Visualisation (avec données mock)', () => {
  // Ces tests nécessiteraient une mindmap existante

  test.skip('affiche le SVG de visualisation', async () => {
    // TODO: Implémenter avec données mock
  });

  test.skip('affiche les contrôles de zoom', async () => {
    // TODO: Implémenter avec données mock
  });

  test.skip('le zoom fonctionne avec les boutons', async () => {
    // TODO: Implémenter avec données mock
  });

  test.skip('le pan fonctionne au drag', async () => {
    // TODO: Implémenter avec données mock
  });

  test.skip('affiche la légende des types de noeuds', async () => {
    // TODO: Implémenter avec données mock
  });
});
