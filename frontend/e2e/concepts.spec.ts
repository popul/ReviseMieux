import { test, expect } from '@playwright/test';

/**
 * Tests E2E pour les Concepts
 * Les concepts sont extraits via l'API backend.
 * Les endpoints sont: POST /api/cours/:id/concepts/extraire, GET /api/cours/:id/concepts,
 * PUT /api/concepts/:id, DELETE /api/concepts/:id
 *
 * Note: Il n'existe pas encore de page UI dediee aux concepts.
 * Ces tests verifient uniquement les appels API.
 * Tous les tests requierent RUN_API_TESTS=1 car ils appellent le backend directement.
 */

test.describe('Concepts - API', () => {
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

  test('extrait les concepts d\'un cours via l\'API', async ({ request }) => {
    if (!coursId) {
      test.skip();
      return;
    }

    const response = await request.post(`/api/cours/${coursId}/concepts/extraire`);
    expect(response.ok()).toBeTruthy();

    const data = await response.json();
    expect(data.succes).toBeTruthy();
    expect(data.concepts).toBeDefined();
    expect(Array.isArray(data.concepts)).toBeTruthy();
  });

  test('recupere les concepts d\'un cours via l\'API', async ({ request }) => {
    if (!coursId) {
      test.skip();
      return;
    }

    const response = await request.get(`/api/cours/${coursId}/concepts`);
    expect(response.ok()).toBeTruthy();

    const data = await response.json();
    expect(data.succes).toBeTruthy();
    expect(data.concepts).toBeDefined();
    expect(Array.isArray(data.concepts)).toBeTruthy();

    // Verifier la structure d'un concept si present
    if (data.concepts.length > 0) {
      const concept = data.concepts[0];
      expect(concept.id).toBeDefined();
      expect(concept.nom).toBeDefined();
      expect(concept.definition).toBeDefined();
      expect(['essentiel', 'important', 'secondaire']).toContain(concept.importance);
    }
  });

  test('met a jour un concept via l\'API', async ({ request }) => {
    if (!coursId) {
      test.skip();
      return;
    }

    // Recuperer les concepts existants
    const getResponse = await request.get(`/api/cours/${coursId}/concepts`);
    const getData = await getResponse.json();

    if (!getData.concepts || getData.concepts.length === 0) {
      test.skip();
      return;
    }

    const conceptId = getData.concepts[0].id;

    // Mettre a jour le concept
    const updateResponse = await request.put(`/api/concepts/${conceptId}`, {
      data: {
        nom: 'Concept modifie test E2E',
        definition: 'Definition modifiee pour le test',
      },
    });

    expect(updateResponse.ok()).toBeTruthy();

    const updateData = await updateResponse.json();
    expect(updateData.succes).toBeTruthy();
    if (updateData.concept) {
      expect(updateData.concept.nom).toBe('Concept modifie test E2E');
    }
  });

  test('supprime un concept via l\'API', async ({ request }) => {
    if (!coursId) {
      test.skip();
      return;
    }

    // Recuperer les concepts existants
    const getResponse = await request.get(`/api/cours/${coursId}/concepts`);
    const getData = await getResponse.json();

    if (!getData.concepts || getData.concepts.length === 0) {
      test.skip();
      return;
    }

    // Prendre le dernier concept pour eviter de casser les autres tests
    const conceptId = getData.concepts[getData.concepts.length - 1].id;

    const deleteResponse = await request.delete(`/api/concepts/${conceptId}`);
    expect(deleteResponse.ok()).toBeTruthy();
  });

  test('les concepts ont les bons niveaux d\'importance', async ({ request }) => {
    if (!coursId) {
      test.skip();
      return;
    }

    const response = await request.get(`/api/cours/${coursId}/concepts`);
    const data = await response.json();

    if (!data.concepts || data.concepts.length === 0) {
      test.skip();
      return;
    }

    // Verifier que chaque concept a une importance valide
    for (const concept of data.concepts) {
      expect(['essentiel', 'important', 'secondaire']).toContain(concept.importance);
    }
  });
});
