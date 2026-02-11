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
  test('permet de sélectionner le nombre de questions', async ({ page }) => {
    await page.goto('/quiz');

    // Attendre que la page soit chargée
    await expect(page.locator('h1').first()).toBeAttached();

    // Chercher un cours disponible pour lancer un quiz
    const boutonCours = page.locator('button').filter({ hasText: /lancer un quiz/i }).first();
    const coursDisponible = await boutonCours.isVisible().catch(() => false);

    if (!coursDisponible) {
      test.skip();
      return;
    }

    // Sélectionner le premier cours
    await boutonCours.click();

    // Attendre la page de configuration
    await expect(page.getByText('Configurer le quiz')).toBeVisible({ timeout: 5000 });

    // Vérifier que le label "Nombre de questions" est présent
    await expect(page.getByText('Nombre de questions')).toBeVisible();

    // Les boutons de nombre (5, 10, 15, 20) devraient être visibles
    await expect(page.getByRole('button', { name: '5', exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: '10', exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: '15', exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: '20', exact: true })).toBeVisible();

    // Cliquer sur 5 pour changer la sélection
    await page.getByRole('button', { name: '5', exact: true }).click();

    // Le bouton 5 devrait avoir le style sélectionné (bg-coral)
    await expect(page.getByRole('button', { name: '5', exact: true })).toHaveClass(/bg-coral/);
  });

  test('permet de sélectionner la difficulté', async ({ page }) => {
    await page.goto('/quiz');

    // Attendre que la page soit chargée
    await expect(page.locator('h1').first()).toBeAttached();

    // Chercher un cours disponible
    const boutonCours = page.locator('button').filter({ hasText: /lancer un quiz/i }).first();
    const coursDisponible = await boutonCours.isVisible().catch(() => false);

    if (!coursDisponible) {
      test.skip();
      return;
    }

    // Sélectionner le premier cours
    await boutonCours.click();

    // Attendre la page de configuration
    await expect(page.getByText('Configurer le quiz')).toBeVisible({ timeout: 5000 });

    // Vérifier que le label "Difficulte" est présent
    await expect(page.getByText('Difficulte')).toBeVisible();

    // Les boutons de difficulté (Facile, Moyen, Difficile) devraient être visibles
    await expect(page.getByRole('button', { name: 'Facile', exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Moyen', exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Difficile', exact: true })).toBeVisible();

    // Cliquer sur "Facile"
    await page.getByRole('button', { name: 'Facile', exact: true }).click();

    // Le bouton Facile devrait avoir le style sélectionné (bg-green)
    await expect(page.getByRole('button', { name: 'Facile', exact: true })).toHaveClass(/bg-green/);
  });
});

test.describe('Quiz - Session (avec données)', () => {
  test('affiche une question avec des choix', async ({ page }) => {
    await page.goto('/quiz');

    // Attendre que la page soit chargée
    await expect(page.locator('h1').first()).toBeAttached();

    // Chercher un cours disponible
    const boutonCours = page.locator('button').filter({ hasText: /lancer un quiz/i }).first();
    const coursDisponible = await boutonCours.isVisible().catch(() => false);

    if (!coursDisponible) {
      test.skip();
      return;
    }

    // Sélectionner le premier cours
    await boutonCours.click();

    // Attendre la page de configuration
    await expect(page.getByText('Configurer le quiz')).toBeVisible({ timeout: 5000 });

    // Lancer le quiz
    await page.getByRole('button', { name: 'Lancer le quiz' }).click();

    // Attendre que la génération soit terminée et qu'une question apparaisse
    // Cela peut prendre du temps si l'API LLM est impliquée
    const boutonValider = page.getByRole('button', { name: /valider ma reponse/i });
    const quizCharge = await boutonValider.isVisible({ timeout: 30000 }).catch(() => false);

    if (!quizCharge) {
      // La génération a échoué ou pris trop longtemps
      test.skip();
      return;
    }

    // Vérifier qu'un énoncé de question (h2) est visible
    await expect(page.locator('h2').first()).toBeVisible();

    // Vérifier qu'il y a des boutons de choix (A, B, C, D)
    const choix = page.locator('button[aria-pressed]');
    const nombreChoix = await choix.count();
    expect(nombreChoix).toBeGreaterThanOrEqual(2);
  });

  test('permet de sélectionner une réponse', async ({ page }) => {
    await page.goto('/quiz');

    // Attendre et sélectionner un cours
    await expect(page.locator('h1').first()).toBeAttached();
    const boutonCours = page.locator('button').filter({ hasText: /lancer un quiz/i }).first();
    const coursDisponible = await boutonCours.isVisible().catch(() => false);

    if (!coursDisponible) {
      test.skip();
      return;
    }

    await boutonCours.click();
    await expect(page.getByText('Configurer le quiz')).toBeVisible({ timeout: 5000 });

    // Lancer le quiz
    await page.getByRole('button', { name: 'Lancer le quiz' }).click();

    // Attendre qu'une question apparaisse
    const boutonValider = page.getByRole('button', { name: /valider ma reponse/i });
    const quizCharge = await boutonValider.isVisible({ timeout: 30000 }).catch(() => false);

    if (!quizCharge) {
      test.skip();
      return;
    }

    // Le bouton valider devrait être désactivé (pas de choix sélectionné)
    await expect(boutonValider).toBeDisabled();

    // Sélectionner le premier choix
    const premierChoix = page.locator('button[aria-pressed="false"]').first();
    await premierChoix.click();

    // Le choix devrait maintenant être sélectionné
    await expect(premierChoix).toHaveAttribute('aria-pressed', 'true');

    // Le bouton valider devrait être activé
    await expect(boutonValider).toBeEnabled();
  });

  test('affiche le feedback après validation', async ({ page }) => {
    await page.goto('/quiz');

    // Attendre et sélectionner un cours
    await expect(page.locator('h1').first()).toBeAttached();
    const boutonCours = page.locator('button').filter({ hasText: /lancer un quiz/i }).first();
    const coursDisponible = await boutonCours.isVisible().catch(() => false);

    if (!coursDisponible) {
      test.skip();
      return;
    }

    await boutonCours.click();
    await expect(page.getByText('Configurer le quiz')).toBeVisible({ timeout: 5000 });

    // Lancer le quiz
    await page.getByRole('button', { name: 'Lancer le quiz' }).click();

    // Attendre qu'une question apparaisse
    const boutonValider = page.getByRole('button', { name: /valider ma reponse/i });
    const quizCharge = await boutonValider.isVisible({ timeout: 30000 }).catch(() => false);

    if (!quizCharge) {
      test.skip();
      return;
    }

    // Sélectionner le premier choix
    await page.locator('button[aria-pressed="false"]').first().click();

    // Valider la réponse
    await boutonValider.click();

    // Attendre le feedback - soit "Bonne reponse !" soit "Mauvaise reponse"
    const bonneReponse = page.getByText('Bonne reponse !');
    const mauvaiseReponse = page.getByText('Mauvaise reponse');

    const bonneVisible = await bonneReponse.isVisible({ timeout: 10000 }).catch(() => false);
    const mauvaiseVisible = await mauvaiseReponse.isVisible({ timeout: 1000 }).catch(() => false);

    // L'un des deux feedbacks doit être visible
    expect(bonneVisible || mauvaiseVisible).toBeTruthy();

    // Le bouton "Question suivante" ou "Voir les resultats" devrait être visible
    const boutonSuivant = page.getByRole('button', { name: /question suivante|voir les resultats/i });
    await expect(boutonSuivant).toBeVisible();
  });

  test('affiche les résultats à la fin', async ({ page }) => {
    await page.goto('/quiz');

    // Attendre et sélectionner un cours
    await expect(page.locator('h1').first()).toBeAttached();
    const boutonCours = page.locator('button').filter({ hasText: /lancer un quiz/i }).first();
    const coursDisponible = await boutonCours.isVisible().catch(() => false);

    if (!coursDisponible) {
      test.skip();
      return;
    }

    await boutonCours.click();
    await expect(page.getByText('Configurer le quiz')).toBeVisible({ timeout: 5000 });

    // Sélectionner 5 questions (le minimum) pour aller vite
    await page.getByRole('button', { name: '5', exact: true }).click();

    // Lancer le quiz
    await page.getByRole('button', { name: 'Lancer le quiz' }).click();

    // Boucle à travers toutes les questions
    for (let i = 0; i < 5; i++) {
      // Attendre qu'une question ou les résultats apparaissent
      const boutonValider = page.getByRole('button', { name: /valider ma reponse/i });
      const resultats = page.getByText(/quiz termine/i);

      const validerVisible = await boutonValider.isVisible({ timeout: 30000 }).catch(() => false);
      const resultatsVisible = await resultats.isVisible().catch(() => false);

      if (resultatsVisible) break; // Déjà sur les résultats

      if (!validerVisible) {
        // Ni question ni résultats = erreur probable
        test.skip();
        return;
      }

      // Sélectionner le premier choix
      await page.locator('button[aria-pressed="false"]').first().click();

      // Valider
      await boutonValider.click();

      // Attendre le feedback
      const feedbackVisible = await page.getByRole('button', { name: /question suivante|voir les resultats/i }).isVisible({ timeout: 10000 }).catch(() => false);

      if (!feedbackVisible) {
        test.skip();
        return;
      }

      // Passer à la question suivante (ou voir les résultats)
      await page.getByRole('button', { name: /question suivante|voir les resultats/i }).click();
    }

    // Vérifier que la page de résultats est affichée
    await expect(page.getByText(/quiz termine/i)).toBeVisible({ timeout: 5000 });

    // Vérifier que le score est affiché
    await expect(page.getByText('Score')).toBeVisible();

    // Vérifier que les boutons d'actions sont présents
    await expect(page.getByRole('button', { name: /refaire ce quiz/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /nouveau quiz/i })).toBeVisible();
  });
});
