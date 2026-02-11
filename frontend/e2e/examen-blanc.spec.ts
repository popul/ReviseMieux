import { test, expect } from '@playwright/test';

/**
 * Tests E2E pour l'Examen Blanc
 * Note: La page ExamenBlanc n'existe pas encore en tant que composant React.
 * Les types et fonctions API sont definis dans api.ts (genererExamen, demarrerSessionExamen, etc.)
 * mais il n'y a pas encore de route /examen-blanc dans App.tsx ni de lien dans la sidebar.
 *
 * Ces tests verifient l'API backend quand RUN_API_TESTS=1 est defini.
 * Les tests UI seront ajoutés quand la page sera implementee.
 */

test.describe('Examen Blanc - API', () => {
  test.describe.configure({ timeout: 120000 });
  test.skip(({ }) => !process.env.RUN_API_TESTS, 'Skipped: Backend API non disponible. Lancez avec RUN_API_TESTS=1 pour activer.');

  let coursId: string | null = null;

  test.beforeAll(async ({ request }) => {
    // Recuperer un cours existant pour les tests
    const response = await request.get('/api/cours?page=1&limite=1');
    if (response.ok()) {
      const data = await response.json();
      if (data.cours && data.cours.length > 0) {
        coursId = data.cours[0].id;
      }
    }
  });

  test('genere un examen blanc pour un cours via l\'API', async ({ request }) => {
    if (!coursId) {
      test.skip();
      return;
    }

    const response = await request.post(`/api/cours/${coursId}/examen/generer`);
    expect(response.ok()).toBeTruthy();

    const data = await response.json();
    expect(data.succes).toBeTruthy();

    if (data.examen) {
      expect(data.examen.id).toBeDefined();
      expect(data.examen.coursId).toBe(coursId);
      expect(data.examen.questions).toBeDefined();
      expect(Array.isArray(data.examen.questions)).toBeTruthy();
      expect(data.examen.dureeMinutes).toBeGreaterThan(0);

      // Verifier la structure d'une question
      if (data.examen.questions.length > 0) {
        const question = data.examen.questions[0];
        expect(question.numero).toBeDefined();
        expect(question.enonce).toBeDefined();
        expect(question.bareme).toBeGreaterThan(0);
        expect(question.reponseAttendue).toBeDefined();
        expect(['definition', 'comprehension', 'application', 'synthese']).toContain(question.type);
        expect(['facile', 'moyen', 'difficile']).toContain(question.difficulte);
      }
    }
  });

  test('demarre une session d\'examen via l\'API', async ({ request }) => {
    if (!coursId) {
      test.skip();
      return;
    }

    // D'abord generer un examen
    const genResponse = await request.post(`/api/cours/${coursId}/examen/generer`);
    const genData = await genResponse.json();

    if (!genData.examen) {
      test.skip();
      return;
    }

    const examenId = genData.examen.id;

    // Demarrer une session
    const sessionResponse = await request.post(`/api/examens/${examenId}/sessions`);
    expect(sessionResponse.ok()).toBeTruthy();

    const sessionData = await sessionResponse.json();
    expect(sessionData.succes).toBeTruthy();

    if (sessionData.session) {
      expect(sessionData.session.id).toBeDefined();
      expect(sessionData.session.examenId).toBe(examenId);
      expect(sessionData.session.termine).toBe(false);
    }
  });

  test('demande un indice pendant l\'examen via l\'API', async ({ request }) => {
    if (!coursId) {
      test.skip();
      return;
    }

    // Generer un examen et demarrer une session
    const genResponse = await request.post(`/api/cours/${coursId}/examen/generer`);
    const genData = await genResponse.json();

    if (!genData.examen || genData.examen.questions.length === 0) {
      test.skip();
      return;
    }

    const sessionResponse = await request.post(`/api/examens/${genData.examen.id}/sessions`);
    const sessionData = await sessionResponse.json();

    if (!sessionData.session) {
      test.skip();
      return;
    }

    // Demander un indice pour la premiere question
    const indiceResponse = await request.post(`/api/sessions-examen/${sessionData.session.id}/indice`, {
      data: {
        questionNumero: 1,
        niveauIndice: 1,
      },
    });

    expect(indiceResponse.ok()).toBeTruthy();

    const indiceData = await indiceResponse.json();
    expect(indiceData.succes).toBeTruthy();

    if (indiceData.indice) {
      expect(indiceData.indice.texte).toBeDefined();
      expect(indiceData.indice.niveau).toBe(1);
    }
  });

  test('corrige un examen via l\'API', async ({ request }) => {
    if (!coursId) {
      test.skip();
      return;
    }

    // Generer un examen et demarrer une session
    const genResponse = await request.post(`/api/cours/${coursId}/examen/generer`);
    const genData = await genResponse.json();

    if (!genData.examen || genData.examen.questions.length === 0) {
      test.skip();
      return;
    }

    const sessionResponse = await request.post(`/api/examens/${genData.examen.id}/sessions`);
    const sessionData = await sessionResponse.json();

    if (!sessionData.session) {
      test.skip();
      return;
    }

    // Soumettre des reponses
    const reponses = genData.examen.questions.map((q: { numero: number }) => ({
      questionNumero: q.numero,
      texte: 'Reponse de test E2E pour la question',
    }));

    const correctionResponse = await request.post(`/api/sessions-examen/${sessionData.session.id}/corriger`, {
      data: { reponses },
    });

    expect(correctionResponse.ok()).toBeTruthy();

    const correctionData = await correctionResponse.json();
    expect(correctionData.succes).toBeTruthy();

    if (correctionData.resultat) {
      expect(correctionData.resultat.noteEstimee).toBeDefined();
      expect(correctionData.resultat.pointsForts).toBeDefined();
      expect(correctionData.resultat.pointsFaibles).toBeDefined();
      expect(correctionData.resultat.planRevision).toBeDefined();
      expect(correctionData.resultat.details).toBeDefined();
    }
  });
});

test.describe('Examen Blanc - Page UI', () => {
  test('la page /examen-blanc se charge', async ({ page }) => {
    await page.goto('/examen-blanc');
    await expect(page).toHaveURL(/.*examen-blanc/);

    // La page doit afficher soit le titre "Examen blanc" soit un message de chargement/erreur
    const titreExamen = page.getByRole('heading', { name: /examen blanc/i });
    const messagePasDeCours = page.getByText(/aucun cours disponible/i);
    const messageChargement = page.getByText(/chargement/i);

    const titreVisible = await titreExamen.isVisible().catch(() => false);
    const pasDeCoursVisible = await messagePasDeCours.isVisible().catch(() => false);
    const chargementVisible = await messageChargement.isVisible().catch(() => false);

    expect(titreVisible || pasDeCoursVisible || chargementVisible).toBeTruthy();
  });

  test('affiche la liste des cours pour selection', async ({ page }) => {
    await page.goto('/examen-blanc');
    await page.waitForTimeout(3000);

    const titreExamen = page.getByRole('heading', { name: /examen blanc/i });
    const titreVisible = await titreExamen.isVisible().catch(() => false);

    if (!titreVisible) {
      // Pas de cours disponibles ou erreur
      test.skip();
      return;
    }

    // Verifier le message de selection
    await expect(page.getByText(/selectionnez un cours/i)).toBeVisible();
  });

  test('les cartes de cours ont le bouton passer un examen', async ({ page }) => {
    await page.goto('/examen-blanc');
    await page.waitForTimeout(3000);

    const boutonsExamen = page.getByText(/passer un examen blanc/i);
    const nombreBoutons = await boutonsExamen.count();

    if (nombreBoutons === 0) {
      // Pas de cours disponibles
      test.skip();
      return;
    }

    await expect(boutonsExamen.first()).toBeVisible();
  });

  test('cliquer sur un cours lance la generation', async ({ page }) => {
    await page.goto('/examen-blanc');
    await page.waitForTimeout(3000);

    const boutonsExamen = page.getByText(/passer un examen blanc/i);
    const nombreBoutons = await boutonsExamen.count();

    if (nombreBoutons === 0) {
      test.skip();
      return;
    }

    // Cliquer sur le premier cours
    await boutonsExamen.first().click();

    // Verifier que l'URL change avec le parametre cours
    await expect(page).toHaveURL(/.*examen-blanc\?cours=/);

    // Attendre que la page affiche un message de generation, une erreur, ou l'examen
    await page.waitForTimeout(2000);

    const messageGeneration = page.getByText(/generation.*en cours/i);
    const messageErreur = page.getByRole('heading', { name: /erreur/i });
    const messageExamen = page.getByRole('heading', { name: /examen blanc/i });

    const generationVisible = await messageGeneration.isVisible().catch(() => false);
    const erreurVisible = await messageErreur.isVisible().catch(() => false);
    const examenVisible = await messageExamen.isVisible().catch(() => false);

    // L'un des trois etats doit etre visible
    expect(generationVisible || erreurVisible || examenVisible).toBeTruthy();
  });
});
