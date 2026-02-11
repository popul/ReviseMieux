import { test, expect } from '@playwright/test';

/**
 * Tests E2E pour la page Lexique
 * Verifie l'affichage de la liste des termes, le quiz vocabulaire,
 * et le mode flashcards.
 * Route: /lexique (selection de cours) et /lexique?cours=<id> (lexique d'un cours)
 */

test.describe('Lexique - Selection de cours', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/lexique');
  });

  test('charge la page lexique', async ({ page }) => {
    await expect(page).toHaveURL(/.*lexique/);
    // La page affiche soit le titre "Lexique" soit un message quand il n'y a pas de cours
    const titreLexique = page.getByRole('heading', { name: /lexique/i });
    const messagePasDeCours = page.getByText(/aucun cours disponible/i);
    const messageScanner = page.getByText(/scannez d'abord un cours/i);

    const titreVisible = await titreLexique.isVisible().catch(() => false);
    const pasDeCoursVisible = await messagePasDeCours.isVisible().catch(() => false);
    const scannerVisible = await messageScanner.isVisible().catch(() => false);

    // L'un des trois doit etre visible
    expect(titreVisible || pasDeCoursVisible || scannerVisible).toBeTruthy();
  });

  test('affiche la liste des cours pour selection', async ({ page }) => {
    // Attendre que la page soit chargee
    await page.waitForTimeout(2000);

    const titreLexique = page.getByRole('heading', { name: /lexique/i }).first();
    const titreVisible = await titreLexique.isVisible().catch(() => false);

    if (!titreVisible) {
      // Pas de cours disponibles
      test.skip();
      return;
    }

    // Verifier le message de selection
    await expect(page.getByText(/selectionnez un cours/i)).toBeVisible();
  });

  test('affiche les cours avec lien vers leur lexique', async ({ page }) => {
    await page.waitForTimeout(2000);

    // Chercher les liens "Voir le lexique"
    const liensLexique = page.getByText(/voir le lexique/i);
    const nombreLiens = await liensLexique.count();

    if (nombreLiens === 0) {
      // Pas de cours disponibles
      test.skip();
      return;
    }

    // Verifier qu'au moins un lien est visible
    await expect(liensLexique.first()).toBeVisible();
  });

  test('cliquer sur un cours navigue vers le lexique de ce cours', async ({ page }) => {
    await page.waitForTimeout(2000);

    const liensCours = page.locator('a[href^="/lexique?cours="]');
    const nombreLiens = await liensCours.count();

    if (nombreLiens === 0) {
      test.skip();
      return;
    }

    // Cliquer sur le premier cours
    await liensCours.first().click();

    // Verifier l'URL
    await expect(page).toHaveURL(/\/lexique\?cours=/);
  });

  test('affiche le lien vers le scanner quand aucun cours', async ({ page }) => {
    await page.waitForTimeout(2000);

    const messagePasDeCours = page.getByText(/aucun cours disponible/i);
    const pasDeCoursVisible = await messagePasDeCours.isVisible().catch(() => false);

    if (!pasDeCoursVisible) {
      // Des cours existent, ce test n'est pas applicable
      test.skip();
      return;
    }

    // Verifier le lien vers le scanner
    await expect(page.getByRole('link', { name: /scanner un cours/i })).toBeVisible();
  });
});

test.describe('Lexique - Vue des termes', () => {
  test.beforeEach(async ({ page }) => {
    // Naviguer vers le lexique du premier cours disponible
    await page.goto('/lexique');
    await page.waitForTimeout(2000);

    const liensCours = page.locator('a[href^="/lexique?cours="]');
    const nombreLiens = await liensCours.count();

    if (nombreLiens > 0) {
      await liensCours.first().click();
      await expect(page).toHaveURL(/\/lexique\?cours=/);
    }
  });

  test('affiche le titre Lexique et les onglets', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    await expect(page.getByRole('heading', { name: 'Lexique', exact: true })).toBeVisible({ timeout: 10000 });

    // Verifier les trois onglets
    await expect(page.getByRole('button', { name: /lexique/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /quiz vocabulaire/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /flashcards rapides/i })).toBeVisible();
  });

  test('affiche le bouton d\'extraction quand aucun terme', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    // Attendre le chargement
    await page.waitForTimeout(3000);

    // Verifier si le bouton d'extraction est visible (aucun terme)
    const boutonExtraire = page.getByRole('button', { name: /extraire les termes/i });
    const termesPresents = page.getByText(/terme.*au total/i);

    const boutonVisible = await boutonExtraire.isVisible().catch(() => false);
    const termesVisible = await termesPresents.isVisible().catch(() => false);

    // Soit le bouton d'extraction, soit des termes existent
    expect(boutonVisible || termesVisible).toBeTruthy();
  });

  test('affiche les cartes de termes quand des termes existent', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    await page.waitForTimeout(3000);

    const compteurTermes = page.getByText(/\d+ termes? au total/i);
    const termesVisible = await compteurTermes.isVisible().catch(() => false);

    if (!termesVisible) {
      test.skip();
      return;
    }

    // Verifier qu'il y a des cartes de termes (h3 avec le nom du terme)
    const cartesTermes = page.locator('.grid h3');
    const nombreCartes = await cartesTermes.count();
    expect(nombreCartes).toBeGreaterThan(0);
  });

  test('la barre de recherche filtre les termes', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    await page.waitForTimeout(3000);

    const champRecherche = page.getByPlaceholder(/rechercher un terme/i);
    const champVisible = await champRecherche.isVisible().catch(() => false);

    if (!champVisible) {
      // Pas de termes, pas de barre de recherche
      test.skip();
      return;
    }

    // Saisir un texte de recherche
    await champRecherche.fill('xyz_inexistant_test');

    // Verifier que le compteur change
    await expect(page.getByText(/0 termes? trouves?/i)).toBeVisible();

    // Effacer la recherche
    await champRecherche.fill('');

    // Verifier que les termes sont a nouveau visibles
    await expect(page.getByText(/\d+ termes? au total/i)).toBeVisible();
  });

  test('le select de tri est present avec les options', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    await page.waitForTimeout(3000);

    const selectTri = page.locator('select');
    const selectVisible = await selectTri.isVisible().catch(() => false);

    if (!selectVisible) {
      // Pas de termes, pas de select de tri
      test.skip();
      return;
    }

    // Verifier les options de tri
    await expect(selectTri.locator('option[value="alphabetique"]')).toBeAttached();
    await expect(selectTri.locator('option[value="maitrise"]')).toBeAttached();
    await expect(selectTri.locator('option[value="categorie"]')).toBeAttached();
  });

  test('le bouton re-extraire est visible quand des termes existent', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    await page.waitForTimeout(3000);

    const compteurTermes = page.getByText(/\d+ termes? au total/i);
    const termesVisible = await compteurTermes.isVisible().catch(() => false);

    if (!termesVisible) {
      test.skip();
      return;
    }

    // Verifier le bouton Re-extraire
    await expect(page.getByRole('button', { name: /re-extraire/i })).toBeVisible();
  });

  test('le lien retour aux cours est visible', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    await page.waitForTimeout(3000);

    await expect(page.getByText(/retour aux cours/i)).toBeVisible();
  });

  test('l\'indicateur de maitrise est affiche pour chaque terme', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    await page.waitForTimeout(3000);

    // Verifier si des termes existent
    const compteurTermes = page.getByText(/\d+ termes? au total/i);
    const termesVisible = await compteurTermes.isVisible().catch(() => false);

    if (!termesVisible) {
      test.skip();
      return;
    }

    // Verifier qu'il y a des indicateurs de maitrise (groupes de boutons avec aria-label)
    const indicateursMaitrise = page.locator('[role="group"][aria-label*="Maitrise"]');
    const nombreIndicateurs = await indicateursMaitrise.count();
    expect(nombreIndicateurs).toBeGreaterThan(0);
  });
});

test.describe('Lexique - Quiz vocabulaire', () => {
  test.beforeEach(async ({ page }) => {
    // Naviguer vers le lexique d'un cours avec des termes
    await page.goto('/lexique');
    await page.waitForTimeout(2000);

    const liensCours = page.locator('a[href^="/lexique?cours="]');
    const nombreLiens = await liensCours.count();

    if (nombreLiens > 0) {
      await liensCours.first().click();
      await expect(page).toHaveURL(/\/lexique\?cours=/);
      await page.waitForTimeout(3000);
    }
  });

  test('le mode quiz est accessible via l\'onglet', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    const ongletQuiz = page.getByRole('button', { name: /quiz vocabulaire/i });
    await expect(ongletQuiz).toBeVisible({ timeout: 10000 });
    await ongletQuiz.click();

    // Verifier qu'on est en mode quiz
    // Soit le selecteur de mode, soit un message "extrayez d'abord les termes"
    const titreQuiz = page.getByRole('heading', { name: /quiz vocabulaire/i });
    const messageExtraire = page.getByText(/extrayez d'abord les termes/i);

    const quizVisible = await titreQuiz.isVisible().catch(() => false);
    const extraireVisible = await messageExtraire.isVisible().catch(() => false);

    expect(quizVisible || extraireVisible).toBeTruthy();
  });

  test('affiche le selecteur de mode avec les deux options', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    // Aller a l'onglet quiz
    await page.getByRole('button', { name: /quiz vocabulaire/i }).click();

    // Verifier les options de mode
    const modeTermeDefinition = page.getByText(/terme.*definition/i).first();
    const modeDefinitionTerme = page.getByText(/definition.*terme/i).first();

    const termeDefVisible = await modeTermeDefinition.isVisible().catch(() => false);

    if (!termeDefVisible) {
      // Pas de termes, le quiz ne peut pas demarrer
      test.skip();
      return;
    }

    await expect(modeTermeDefinition).toBeVisible();
    await expect(modeDefinitionTerme).toBeVisible();
  });

  test('le bouton lancer le quiz est present', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    await page.getByRole('button', { name: /quiz vocabulaire/i }).click();

    const boutonLancer = page.getByRole('button', { name: /lancer le quiz/i });
    const boutonVisible = await boutonLancer.isVisible().catch(() => false);

    if (!boutonVisible) {
      // Pas de termes disponibles
      test.skip();
      return;
    }

    await expect(boutonLancer).toBeVisible();
    await expect(boutonLancer).toBeEnabled();
  });
});

test.describe('Lexique - Quiz vocabulaire (API)', () => {
  test.describe.configure({ timeout: 120000 });
  test.skip(({ }) => !process.env.RUN_API_TESTS, 'Skipped: Backend API non disponible. Lancez avec RUN_API_TESTS=1 pour activer.');

  test('lance un quiz et affiche les questions', async ({ page }) => {
    await page.goto('/lexique');
    await page.waitForTimeout(2000);

    const liensCours = page.locator('a[href^="/lexique?cours="]');
    const nombreLiens = await liensCours.count();

    if (nombreLiens === 0) {
      test.skip();
      return;
    }

    await liensCours.first().click();
    await expect(page).toHaveURL(/\/lexique\?cours=/);
    await page.waitForTimeout(3000);

    // Aller a l'onglet quiz
    await page.getByRole('button', { name: /quiz vocabulaire/i }).click();

    const boutonLancer = page.getByRole('button', { name: /lancer le quiz/i });
    const boutonVisible = await boutonLancer.isVisible().catch(() => false);

    if (!boutonVisible) {
      test.skip();
      return;
    }

    // Lancer le quiz
    await boutonLancer.click();

    // Attendre que les questions se chargent
    await expect(page.getByText(/question \d+ \/ \d+/i)).toBeVisible({ timeout: 60000 });

    // Verifier que la zone de reponse est visible
    await expect(page.locator('textarea')).toBeVisible();

    // Verifier les boutons d'action
    await expect(page.getByRole('button', { name: /verifier/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /je ne sais pas/i })).toBeVisible();
  });

  test('affiche le feedback apres verification', async ({ page }) => {
    await page.goto('/lexique');
    await page.waitForTimeout(2000);

    const liensCours = page.locator('a[href^="/lexique?cours="]');
    if ((await liensCours.count()) === 0) { test.skip(); return; }

    await liensCours.first().click();
    await page.waitForTimeout(3000);

    await page.getByRole('button', { name: /quiz vocabulaire/i }).click();

    const boutonLancer = page.getByRole('button', { name: /lancer le quiz/i });
    if (!(await boutonLancer.isVisible().catch(() => false))) { test.skip(); return; }

    await boutonLancer.click();
    await expect(page.getByText(/question \d+ \/ \d+/i)).toBeVisible({ timeout: 60000 });

    // Saisir une reponse
    await page.locator('textarea').fill('Ma reponse de test');

    // Verifier
    await page.getByRole('button', { name: /verifier/i }).click();

    // Le feedback doit apparaitre (correct ou incorrect)
    const feedbackCorrect = page.getByText(/correct/i);
    const feedbackIncorrect = page.getByText(/incorrect/i);

    const correctVisible = await feedbackCorrect.first().isVisible().catch(() => false);
    const incorrectVisible = await feedbackIncorrect.first().isVisible().catch(() => false);

    expect(correctVisible || incorrectVisible).toBeTruthy();

    // Le bouton "Question suivante" ou "Voir les resultats" doit apparaitre
    const boutonSuivant = page.getByRole('button', { name: /question suivante|voir les resultats/i });
    await expect(boutonSuivant).toBeVisible();
  });
});

test.describe('Lexique - Flashcards', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/lexique');
    await page.waitForTimeout(2000);

    const liensCours = page.locator('a[href^="/lexique?cours="]');
    const nombreLiens = await liensCours.count();

    if (nombreLiens > 0) {
      await liensCours.first().click();
      await expect(page).toHaveURL(/\/lexique\?cours=/);
      await page.waitForTimeout(3000);
    }
  });

  test('le mode flashcards est accessible via l\'onglet', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    const ongletFlashcards = page.getByRole('button', { name: /flashcards rapides/i });
    await expect(ongletFlashcards).toBeVisible({ timeout: 10000 });
    await ongletFlashcards.click();

    // Verifier qu'on est en mode flashcards
    // Soit la carte, soit le message "extrayez d'abord"
    const messageTerm = page.getByText(/terme/i).first();
    const messageExtraire = page.getByText(/extrayez d'abord les termes/i);

    const termeVisible = await messageTerm.isVisible().catch(() => false);
    const extraireVisible = await messageExtraire.isVisible().catch(() => false);

    expect(termeVisible || extraireVisible).toBeTruthy();
  });

  test('affiche une flashcard avec le terme', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    await page.getByRole('button', { name: /flashcards rapides/i }).click();

    const messageExtraire = page.getByText(/extrayez d'abord les termes/i);
    const extraireVisible = await messageExtraire.isVisible().catch(() => false);

    if (extraireVisible) {
      // Pas de termes
      test.skip();
      return;
    }

    // Verifier la carte flashcard
    await expect(page.getByText('Terme').first()).toBeVisible();
    await expect(page.getByText(/cliquez pour reveler/i)).toBeVisible();
  });

  test('cliquer sur la carte revele la definition', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    await page.getByRole('button', { name: /flashcards rapides/i }).click();

    const messageExtraire = page.getByText(/extrayez d'abord les termes/i);
    if (await messageExtraire.isVisible().catch(() => false)) {
      test.skip();
      return;
    }

    // Cliquer sur la carte pour la retourner
    const carte = page.locator('.cursor-pointer.min-h-\\[250px\\]');
    const carteVisible = await carte.isVisible().catch(() => false);

    if (!carteVisible) {
      test.skip();
      return;
    }

    await carte.click();

    // La definition doit etre visible
    await expect(page.getByText('Definition').first()).toBeVisible();
  });

  test('les boutons je sais et je ne sais pas sont presents', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    await page.getByRole('button', { name: /flashcards rapides/i }).click();

    const messageExtraire = page.getByText(/extrayez d'abord les termes/i);
    if (await messageExtraire.isVisible().catch(() => false)) {
      test.skip();
      return;
    }

    // Verifier les boutons d'action
    await expect(page.getByRole('button', { name: /je ne sais pas/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /je sais/i })).toBeVisible();
  });

  test('les boutons de navigation precedent/suivant sont presents', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    await page.getByRole('button', { name: /flashcards rapides/i }).click();

    const messageExtraire = page.getByText(/extrayez d'abord les termes/i);
    if (await messageExtraire.isVisible().catch(() => false)) {
      test.skip();
      return;
    }

    // Verifier les boutons de navigation
    const boutonPrecedent = page.getByRole('button', { name: /precedent/i });
    const boutonSuivant = page.getByRole('button', { name: /suivant/i });

    await expect(boutonPrecedent).toBeVisible();
    await expect(boutonSuivant).toBeVisible();

    // Le bouton Precedent doit etre desactive sur la premiere carte
    await expect(boutonPrecedent).toBeDisabled();
  });

  test('la barre de progression est affichee', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    await page.getByRole('button', { name: /flashcards rapides/i }).click();

    const messageExtraire = page.getByText(/extrayez d'abord les termes/i);
    if (await messageExtraire.isVisible().catch(() => false)) {
      test.skip();
      return;
    }

    // Verifier la barre de progression
    await expect(page.getByText(/progression/i)).toBeVisible();
    await expect(page.getByText(/\d+ \/ \d+/)).toBeVisible();
  });

  test('l\'indicateur de maitrise est visible', async ({ page }) => {
    if (!page.url().includes('cours=')) {
      test.skip();
      return;
    }

    await page.getByRole('button', { name: /flashcards rapides/i }).click();

    const messageExtraire = page.getByText(/extrayez d'abord les termes/i);
    if (await messageExtraire.isVisible().catch(() => false)) {
      test.skip();
      return;
    }

    // Verifier l'indicateur de maitrise
    await expect(page.getByText(/maitrise/i)).toBeVisible();
  });
});
