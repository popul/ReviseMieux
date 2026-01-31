import { test, expect } from '@playwright/test';

/**
 * Tests E2E pour la page Quiz
 * Vérifie l'interface de quiz interactif
 */

test.describe('Quiz', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/quiz');
  });

  test('charge la page quiz', async ({ page }) => {
    // Vérifie qu'on est sur la bonne URL
    await expect(page).toHaveURL(/.*quiz/);

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

test.describe('Quiz - Configuration', () => {
  // Tests pour le configurateur de quiz

  test.skip('permet de sélectionner le nombre de questions', async () => {
    // TODO: Implémenter avec données mock ou cours de test
  });

  test.skip('permet de sélectionner la difficulté', async () => {
    // TODO: Implémenter avec données mock
  });
});

test.describe('Quiz - Session (avec données mock)', () => {
  // Ces tests nécessiteraient un backend ou des mocks complets

  test.skip('affiche une question avec 4 choix', async () => {
    // TODO: Implémenter avec session mock
  });

  test.skip('permet de sélectionner une réponse', async () => {
    // TODO: Implémenter avec session mock
  });

  test.skip('affiche le feedback après validation', async () => {
    // TODO: Implémenter avec session mock
  });

  test.skip('affiche les résultats à la fin', async () => {
    // TODO: Implémenter avec session mock
  });
});
