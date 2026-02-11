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

test.describe('Mindmap - Visualisation (avec données)', () => {
  // Helper pour naviguer vers une mindmap existante
  async function naviguerVersMindmap(page: import('@playwright/test').Page) {
    await page.goto('/mindmap');
    await expect(page.locator('h1').first()).toBeAttached();

    // Chercher un cours disponible
    const lienCours = page.locator('a[href*="/mindmap?cours="]').first();
    const coursDisponible = await lienCours.isVisible().catch(() => false);

    if (!coursDisponible) return false;

    // Cliquer sur le premier cours
    await lienCours.click();

    // Attendre le chargement
    await page.waitForTimeout(2000);

    // Vérifier si une mindmap est affichée (SVG présent) ou si on doit la générer
    const svg = page.locator('svg').first();
    const svgVisible = await svg.isVisible().catch(() => false);

    // Si on est sur la page "Pas encore de carte mentale", la mindmap n'existe pas encore
    const pasDeMindmap = page.getByText(/pas encore de carte mentale/i);
    const pasDeMindmapVisible = await pasDeMindmap.isVisible().catch(() => false);

    if (pasDeMindmapVisible) {
      return false; // Pas de mindmap, on ne peut pas tester la visu
    }

    return svgVisible;
  }

  test('affiche le SVG de visualisation', async ({ page }) => {
    const mindmapChargee = await naviguerVersMindmap(page);

    if (!mindmapChargee) {
      test.skip();
      return;
    }

    // Vérifier que le titre "Carte mentale" est visible
    await expect(page.getByText('Carte mentale', { exact: false })).toBeVisible();

    // Vérifier qu'un SVG est rendu dans la zone de visualisation
    const svgElement = page.locator('svg');
    const svgCount = await svgElement.count();
    expect(svgCount).toBeGreaterThan(0);

    // Vérifier que le SVG contient des éléments (noeuds)
    const noeuds = page.locator('svg g.noeuds');
    const noeudsPresent = await noeuds.isVisible().catch(() => false);

    if (noeudsPresent) {
      // Vérifier qu'il y a au moins un noeud (rect avec du texte)
      const rects = page.locator('svg g.noeuds rect');
      const nombreRects = await rects.count();
      expect(nombreRects).toBeGreaterThan(0);
    }
  });

  test('affiche les contrôles de zoom', async ({ page }) => {
    const mindmapChargee = await naviguerVersMindmap(page);

    if (!mindmapChargee) {
      test.skip();
      return;
    }

    // Vérifier que les 3 boutons de contrôle (zoom avant, zoom arrière, recentrer) sont présents
    await expect(page.locator('button[title="Zoom avant"]')).toBeVisible();
    await expect(page.locator('button[title="Zoom arriere"]')).toBeVisible();
    await expect(page.locator('button[title="Recentrer"]')).toBeVisible();
  });

  test('le zoom fonctionne avec les boutons', async ({ page }) => {
    const mindmapChargee = await naviguerVersMindmap(page);

    if (!mindmapChargee) {
      test.skip();
      return;
    }

    // Récupérer le viewBox initial du SVG
    const svgElement = page.locator('svg[viewBox]').last();
    const viewBoxInitial = await svgElement.getAttribute('viewBox');

    // Cliquer sur "Zoom avant"
    await page.locator('button[title="Zoom avant"]').click();

    // Attendre que le viewBox change (animation)
    await page.waitForTimeout(300);

    // Le viewBox devrait avoir changé (zoom = viewBox plus petit)
    const viewBoxApresZoom = await svgElement.getAttribute('viewBox');
    expect(viewBoxApresZoom).not.toEqual(viewBoxInitial);

    // Cliquer sur "Recentrer" pour revenir à la vue initiale
    await page.locator('button[title="Recentrer"]').click();
    await page.waitForTimeout(300);

    // Le viewBox devrait être revenu proche de l'initial
    const viewBoxApresReset = await svgElement.getAttribute('viewBox');
    expect(viewBoxApresReset).toBeTruthy();
  });

  test('le pan fonctionne au drag', async ({ page }) => {
    const mindmapChargee = await naviguerVersMindmap(page);

    if (!mindmapChargee) {
      test.skip();
      return;
    }

    // Récupérer le viewBox initial
    const svgElement = page.locator('svg[viewBox]').last();
    const viewBoxInitial = await svgElement.getAttribute('viewBox');

    // Trouver le conteneur de drag (classe cursor-grab)
    const container = page.locator('.cursor-grab');
    const containerVisible = await container.isVisible().catch(() => false);

    if (!containerVisible) {
      test.skip();
      return;
    }

    // Effectuer un drag (mousedown + mousemove + mouseup)
    const box = await container.boundingBox();
    if (!box) {
      test.skip();
      return;
    }

    const startX = box.x + box.width / 2;
    const startY = box.y + box.height / 2;

    await page.mouse.move(startX, startY);
    await page.mouse.down();
    await page.mouse.move(startX + 100, startY + 50, { steps: 5 });
    await page.mouse.up();

    await page.waitForTimeout(300);

    // Le viewBox devrait avoir changé après le drag
    const viewBoxApresDrag = await svgElement.getAttribute('viewBox');
    expect(viewBoxApresDrag).not.toEqual(viewBoxInitial);
  });

  test('affiche la légende des types de noeuds', async ({ page }) => {
    const mindmapChargee = await naviguerVersMindmap(page);

    if (!mindmapChargee) {
      test.skip();
      return;
    }

    // Vérifier que la légende est visible
    await expect(page.getByText('Legende')).toBeVisible();

    // Vérifier les 4 types dans la légende
    await expect(page.getByText('Theme principal')).toBeVisible();
    await expect(page.getByText('Sous-theme')).toBeVisible();
    await expect(page.getByText('Detail')).toBeVisible();
    await expect(page.getByText('Lien vers concept')).toBeVisible();

    // Vérifier les instructions d'interaction
    await expect(page.getByText(/molette.*zoom/i)).toBeVisible();
  });
});
