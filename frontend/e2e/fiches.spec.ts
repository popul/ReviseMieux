import { test, expect } from '@playwright/test';

/**
 * Tests E2E pour la page Fiches
 * Vérifie l'interface de révision des fiches
 */

test.describe('Fiches', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/fiches');
  });

  test('charge la page fiches', async ({ page }) => {
    // Vérifie qu'on est sur la bonne URL
    await expect(page).toHaveURL(/.*fiches/);

    // Attend que la page soit chargée (any h1 is attached to DOM)
    await expect(page.locator('h1').first()).toBeAttached();
  });

  test('navigation clavier fonctionne', async ({ page }) => {
    // Vérifie que la page est navigable au clavier
    await page.keyboard.press('Tab');

    // Un élément devrait être focusé
    const focusedElement = page.locator(':focus');
    await expect(focusedElement).toBeVisible();
  });
});

test.describe('Fiches - Mode révision (avec données mock)', () => {
  // Ces tests nécessiteraient un backend ou des mocks
  // Ils sont préparés pour une implémentation future

  test.skip('affiche une fiche avec question visible', async () => {
    // TODO: Implémenter avec données mock
  });

  test.skip('retourne la fiche au clic', async () => {
    // TODO: Implémenter avec données mock
  });

  test.skip('navigue entre les fiches avec les boutons', async () => {
    // TODO: Implémenter avec données mock
  });
});
