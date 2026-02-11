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

test.describe('Fiches - Mode révision (avec données)', () => {
  test('affiche une fiche avec question visible', async ({ page }) => {
    await page.goto('/fiches');

    // Attendre que la page soit chargée
    await expect(page.locator('h1').first()).toBeAttached();

    // Chercher un cours disponible pour réviser
    const lienCours = page.locator('a[href*="/fiches?cours="]').first();
    const coursDisponible = await lienCours.isVisible().catch(() => false);

    if (!coursDisponible) {
      test.skip();
      return;
    }

    // Cliquer sur le premier cours
    await lienCours.click();

    // Attendre le chargement des fiches
    await page.waitForTimeout(2000);

    // Si aucune fiche, on peut voir le message "Aucune fiche disponible" ou le bouton de génération
    const carteFiche = page.locator('[role="button"][aria-label]').first();
    const ficheVisible = await carteFiche.isVisible().catch(() => false);

    if (!ficheVisible) {
      // Pas de fiches pour ce cours, skip
      test.skip();
      return;
    }

    // Vérifier que le texte "Question" est visible (face avant de la fiche)
    await expect(page.getByText('Question', { exact: false })).toBeVisible();
  });

  test('retourne la fiche au clic', async ({ page }) => {
    await page.goto('/fiches');

    // Attendre que la page soit chargée
    await expect(page.locator('h1').first()).toBeAttached();

    // Chercher un cours disponible
    const lienCours = page.locator('a[href*="/fiches?cours="]').first();
    const coursDisponible = await lienCours.isVisible().catch(() => false);

    if (!coursDisponible) {
      test.skip();
      return;
    }

    await lienCours.click();
    await page.waitForTimeout(2000);

    // Vérifier qu'une carte fiche est présente (CarteFiche a role="button")
    const carteFiche = page.locator('[role="button"][aria-label="Afficher la réponse"]').first();
    const ficheVisible = await carteFiche.isVisible().catch(() => false);

    if (!ficheVisible) {
      test.skip();
      return;
    }

    // Le label initial indique "Afficher la réponse" (face question)
    await expect(carteFiche).toHaveAttribute('aria-label', 'Afficher la réponse');

    // Cliquer pour retourner la fiche
    await carteFiche.click();

    // Après le clic, le label devrait devenir "Afficher la question" (face réponse)
    await expect(page.locator('[role="button"][aria-label="Afficher la question"]')).toBeVisible({ timeout: 2000 });
  });

  test('navigue entre les fiches avec les boutons', async ({ page }) => {
    await page.goto('/fiches');

    // Attendre que la page soit chargée
    await expect(page.locator('h1').first()).toBeAttached();

    // Chercher un cours disponible
    const lienCours = page.locator('a[href*="/fiches?cours="]').first();
    const coursDisponible = await lienCours.isVisible().catch(() => false);

    if (!coursDisponible) {
      test.skip();
      return;
    }

    await lienCours.click();
    await page.waitForTimeout(2000);

    // Vérifier que les contrôles de navigation sont présents
    const boutonSuivant = page.getByRole('button', { name: 'Fiche suivante' });
    const boutonPrecedent = page.getByRole('button', { name: 'Fiche précédente' });

    const controleVisible = await boutonSuivant.isVisible().catch(() => false);
    if (!controleVisible) {
      test.skip();
      return;
    }

    // Le bouton précédent devrait être désactivé sur la première fiche
    await expect(boutonPrecedent).toBeDisabled();

    // Vérifier que le compteur de progression affiche "1 /"
    await expect(page.getByText(/1 \//).first()).toBeVisible();

    // Cliquer sur suivant (si pas désactivé = il y a plus d'une fiche)
    const suivantDesactive = await boutonSuivant.isDisabled();
    if (!suivantDesactive) {
      await boutonSuivant.click();

      // Le compteur devrait maintenant afficher "2 /"
      await expect(page.getByText(/2 \//).first()).toBeVisible();

      // Le bouton précédent ne devrait plus être désactivé
      await expect(boutonPrecedent).toBeEnabled();
    }
  });
});
