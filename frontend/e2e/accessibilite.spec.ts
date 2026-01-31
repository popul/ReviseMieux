import { test, expect } from '@playwright/test';

/**
 * Tests E2E pour l'accessibilité
 * Vérifie les fonctionnalités WCAG 2.1 AA
 */

test.describe('Accessibilité', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('le skip-to-content link est fonctionnel', async ({ page }) => {
    // Focus le skip link avec Tab
    await page.keyboard.press('Tab');

    const skipLink = page.getByRole('link', { name: /aller au contenu principal/i });
    await expect(skipLink).toBeFocused();

    // Active le skip link
    await page.keyboard.press('Enter');

    // Le focus devrait être sur le main content
    const mainContent = page.locator('#main-content');
    await expect(mainContent).toBeFocused();
  });

  test('la navigation principale a les attributs ARIA corrects', async ({ page }) => {
    // Vérifie la navigation principale (pas toutes les navs)
    const mainNav = page.getByRole('navigation', { name: 'Menu principal' });
    await expect(mainNav).toBeVisible();

    // Vérifie que la nav a un aria-label
    await expect(mainNav).toHaveAttribute('aria-label', 'Menu principal');
  });

  test('les contrôles d\'accessibilité sont présents', async ({ page }) => {
    // Le bouton d'accessibilité a un aria-label spécifique
    const accessibilityButton = page.getByRole('button', { name: /options d'accessibilite/i });
    await expect(accessibilityButton).toBeVisible();
  });

  test('tous les liens ont un texte accessible', async ({ page }) => {
    // Récupère tous les liens
    const links = page.getByRole('link');
    const linkCount = await links.count();

    // Vérifie que chaque lien a du contenu textuel
    for (let i = 0; i < linkCount; i++) {
      const link = links.nth(i);
      const text = await link.textContent();
      const ariaLabel = await link.getAttribute('aria-label');

      // Chaque lien doit avoir soit du texte, soit un aria-label
      expect(text || ariaLabel).toBeTruthy();
    }
  });

  test('les boutons ont un texte accessible', async ({ page }) => {
    const buttons = page.getByRole('button');
    const buttonCount = await buttons.count();

    for (let i = 0; i < buttonCount; i++) {
      const button = buttons.nth(i);
      const text = await button.textContent();
      const ariaLabel = await button.getAttribute('aria-label');

      // Chaque bouton doit avoir soit du texte, soit un aria-label
      expect(text || ariaLabel).toBeTruthy();
    }
  });

  test('le focus est visible sur tous les éléments interactifs', async ({ page }) => {
    // Tab à travers plusieurs éléments et vérifie le focus
    for (let i = 0; i < 5; i++) {
      await page.keyboard.press('Tab');

      // Un élément devrait être focusé
      const focused = page.locator(':focus');
      await expect(focused).toBeVisible();
    }
  });
});

test.describe('Accessibilité - Modes', () => {
  test('change de taille de texte', async ({ page }) => {
    await page.goto('/');

    // Ouvre le panneau d'accessibilité
    await page.getByRole('button', { name: /options d'accessibilite/i }).click();

    // Clique sur le bouton "Grand" (exact match pour éviter "Tres grand")
    const largeSizeButton = page.getByRole('radio', { name: 'Grand', exact: true });
    await largeSizeButton.click();

    // Vérifie que la classe est appliquée au HTML
    const htmlClass = await page.locator('html').getAttribute('class');
    expect(htmlClass).toContain('text-scale-large');
  });

  test('active le contraste élevé', async ({ page }) => {
    await page.goto('/');

    // Ouvre le panneau d'accessibilité
    await page.getByRole('button', { name: /options d'accessibilite/i }).click();

    // Active le switch de contraste élevé
    const contrastSwitch = page.getByRole('switch', { name: /contraste/i });
    await contrastSwitch.click();

    // Vérifie que la classe est appliquée au HTML
    const htmlClass = await page.locator('html').getAttribute('class');
    expect(htmlClass).toContain('contraste-eleve');
  });

  test.skip('active le mode daltonien', async () => {
    // TODO: Implémenter - nécessite de tester le select
  });
});
