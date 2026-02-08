import { test, expect } from '@playwright/test';

/**
 * Tests E2E pour le Dashboard
 * Vérifie la navigation et l'affichage des statistiques
 */

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('affiche le titre et la navigation', async ({ page }) => {
    // Vérifie le titre de l'application dans la sidebar (format "Revisemieux")
    await expect(page.getByRole('link', { name: /revise.*mieux.*accueil/i })).toBeVisible();

    // Vérifie la navigation principale
    await expect(page.getByRole('navigation', { name: 'Menu principal' })).toBeVisible();
    await expect(page.getByRole('link', { name: /tableau de bord/i })).toBeVisible();

    // Utilise le premier lien Scanner (celui dans la sidebar)
    await expect(page.getByRole('link', { name: 'Scanner un cours', exact: true })).toBeVisible();
    await expect(page.getByRole('link', { name: /fiches de révision/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /^quiz$/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /cartes mentales/i })).toBeVisible();
  });

  test('navigue vers Scanner depuis la sidebar', async ({ page }) => {
    // Clique sur le lien Scanner dans la sidebar (exact match)
    await page.getByRole('link', { name: 'Scanner un cours', exact: true }).click();

    // Vérifie qu'on est sur la page Scanner
    await expect(page).toHaveURL(/.*scanner/);
  });

  test('navigue vers Fiches depuis le Dashboard', async ({ page }) => {
    await page.getByRole('link', { name: /fiches de révision/i }).click();
    await expect(page).toHaveURL(/.*fiches/);
  });

  test('navigue vers Quiz depuis le Dashboard', async ({ page }) => {
    // Utilise le lien Quiz exact dans la sidebar
    await page.getByRole('link', { name: /^quiz$/i }).click();
    await expect(page).toHaveURL(/.*quiz/);
  });

  test('navigue vers Mindmap depuis le Dashboard', async ({ page }) => {
    await page.getByRole('link', { name: /cartes mentales/i }).click();
    await expect(page).toHaveURL(/.*mindmap/);
  });

  test('le skip-to-content link est accessible au clavier', async ({ page }) => {
    // Le skip link devrait être le premier élément focusable
    await page.keyboard.press('Tab');

    // Vérifie que le skip link est focusé
    const skipLink = page.getByRole('link', { name: /aller au contenu principal/i });
    await expect(skipLink).toBeFocused();
  });

  test('cliquer sur un cours récent navigue vers la page de détail du cours', async ({ page }) => {
    // Attendre que la section des cours récents soit visible
    const sectionCoursRecents = page.locator('h2').filter({ hasText: /tes cours récents/i });

    // Si pas de section cours récents (aucun cours), on skip
    const sectionVisible = await sectionCoursRecents.isVisible().catch(() => false);
    if (!sectionVisible) {
      test.skip();
      return;
    }

    // Trouver les cartes de cours (liens qui commencent par /cours?id=)
    const carteCours = page.locator('a[href^="/cours?id="]').first();

    // Vérifier qu'il y a au moins un cours
    const carteVisible = await carteCours.isVisible().catch(() => false);
    if (!carteVisible) {
      test.skip();
      return;
    }

    // Cliquer sur le premier cours
    await carteCours.click();

    // Vérifier qu'on est sur la page de détail du cours (pas directement sur les fiches)
    await expect(page).toHaveURL(/\/cours\?id=/);

    // Vérifier que la page de détail s'affiche correctement
    await expect(page.getByRole('heading', { name: /contenu du cours/i })).toBeVisible({ timeout: 5000 });
  });
});
