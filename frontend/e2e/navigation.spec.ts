import { test, expect } from '@playwright/test';

/**
 * Tests E2E pour la navigation et l'integration entre les pages
 * Verifie que toutes les pages sont accessibles depuis la sidebar,
 * que les nouveaux liens sont presents, et que le parcours utilisateur
 * complet fonctionne de bout en bout.
 */

test.describe('Navigation - Sidebar', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('la sidebar contient tous les liens du menu principal', async ({ page }) => {
    const nav = page.getByRole('navigation', { name: 'Menu principal' });
    await expect(nav).toBeVisible();

    // Menu
    await expect(page.getByRole('link', { name: /tableau de bord/i })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Scanner un cours', exact: true })).toBeVisible();
    await expect(page.getByRole('link', { name: /mes cours/i })).toBeVisible();
  });

  test('la sidebar contient les liens de la section Bibliotheque', async ({ page }) => {
    await expect(page.getByRole('link', { name: /mes cours/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /analyser une copie/i })).toBeVisible();
  });

  test('tous les liens de navigation sont fonctionnels', async ({ page }) => {
    const liens = [
      { nom: /tableau de bord/i, url: /localhost:3000\/$/ },
      { nom: 'Scanner un cours', url: /.*scanner/ },
      { nom: /mes cours/i, url: /.*cours/ },
      { nom: /analyser une copie/i, url: /.*analyser/ },
    ];

    for (const lien of liens) {
      await page.goto('/');
      const linkElement = typeof lien.nom === 'string'
        ? page.getByRole('link', { name: lien.nom, exact: true })
        : page.getByRole('link', { name: lien.nom });
      await linkElement.click();
      await expect(page).toHaveURL(lien.url);
    }
  });
});

test.describe('Navigation - Depuis le detail d\'un cours', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/cours');
    await expect(page.getByRole('heading', { name: /mes cours/i })).toBeVisible({ timeout: 10000 });

    const carteCours = page.locator('[data-testid="cours-card"]').first();
    const carteVisible = await carteCours.isVisible().catch(() => false);

    if (carteVisible) {
      await carteCours.click();
      await expect(page).toHaveURL(/\/cours\?id=/);
      await expect(page.getByRole('heading', { name: /contenu du cours/i })).toBeVisible({ timeout: 10000 });
    }
  });

  test('le lien vers les fiches est present dans le detail du cours', async ({ page }) => {
    if (!page.url().includes('?id=')) {
      test.skip();
      return;
    }

    const lienFiches = page.getByRole('link', { name: /réviser les fiches/i });
    await expect(lienFiches).toBeVisible();

    // Verifier que le lien pointe vers /fiches?cours=
    const href = await lienFiches.getAttribute('href');
    expect(href).toMatch(/\/fiches\?cours=/);
  });

  test('le lien vers le quiz est present dans le detail du cours', async ({ page }) => {
    if (!page.url().includes('?id=')) {
      test.skip();
      return;
    }

    const lienQuiz = page.getByRole('link', { name: /lancer un quiz/i });
    await expect(lienQuiz).toBeVisible();

    const href = await lienQuiz.getAttribute('href');
    expect(href).toMatch(/\/quiz\?cours=/);
  });

  test('le bouton retour aux cours fonctionne', async ({ page }) => {
    if (!page.url().includes('?id=')) {
      test.skip();
      return;
    }

    await page.getByRole('button', { name: /retour aux cours/i }).click();
    await expect(page).toHaveURL(/\/cours$/);
  });

  test('le bouton modifier est present et fonctionnel', async ({ page }) => {
    if (!page.url().includes('?id=')) {
      test.skip();
      return;
    }

    const boutonModifier = page.getByRole('button', { name: /modifier/i });
    await expect(boutonModifier).toBeVisible();

    await boutonModifier.click();

    // En mode edition, les boutons Enregistrer et Annuler doivent etre visibles
    await expect(page.getByRole('button', { name: /enregistrer/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /annuler/i })).toBeVisible();
  });
});

test.describe('Navigation - Depuis la liste des cours', () => {
  test('les liens rapides fiches/quiz/mindmap sont visibles sur les cartes de cours', async ({ page }) => {
    await page.goto('/cours');
    await expect(page.getByRole('heading', { name: /mes cours/i })).toBeVisible({ timeout: 10000 });

    const carteCours = page.locator('[data-testid="cours-card"]').first();
    const carteVisible = await carteCours.isVisible().catch(() => false);

    if (!carteVisible) {
      test.skip();
      return;
    }

    // Verifier les liens rapides dans la carte
    const lienFiches = carteCours.getByRole('link', { name: /fiches/i });
    const lienQuiz = carteCours.getByRole('link', { name: /quiz/i });
    const lienMindmap = carteCours.getByRole('link', { name: /mindmap/i });

    await expect(lienFiches).toBeVisible();
    await expect(lienQuiz).toBeVisible();
    await expect(lienMindmap).toBeVisible();
  });
});

test.describe('Navigation - Parcours utilisateur complet (API)', () => {
  test.describe.configure({ timeout: 180000 });
  test.skip(({ }) => !process.env.RUN_API_TESTS, 'Skipped: Backend API non disponible. Lancez avec RUN_API_TESTS=1 pour activer.');

  test('parcours complet: dashboard vers cours vers outils de revision', async ({ page }) => {
    // 1. Commencer sur le Dashboard
    await page.goto('/');
    await expect(page.getByRole('link', { name: /revise.*mieux.*accueil/i })).toBeVisible();

    // 2. Naviguer vers la liste des cours
    await page.getByRole('link', { name: /mes cours/i }).click();
    await expect(page).toHaveURL(/.*cours/);

    // 3. Si des cours existent, ouvrir le premier
    const carteCours = page.locator('[data-testid="cours-card"]').first();
    const carteVisible = await carteCours.isVisible().catch(() => false);

    if (!carteVisible) {
      // Pas de cours, on ne peut pas continuer le parcours complet
      test.skip();
      return;
    }

    await carteCours.click();
    await expect(page).toHaveURL(/\/cours\?id=/);
    await expect(page.getByText(/contenu du cours/i)).toBeVisible({ timeout: 10000 });

    // 4. Naviguer vers les fiches depuis le cours
    const lienFiches = page.getByRole('link', { name: /réviser les fiches/i });
    const fichesHref = await lienFiches.getAttribute('href');
    await lienFiches.click();
    await expect(page).toHaveURL(/\/fiches\?cours=/);

    // 5. Retourner via la sidebar pour aller au lexique
    await page.getByRole('link', { name: /lexique/i }).click();
    await expect(page).toHaveURL(/.*lexique/);

    // 6. Si des cours sont listes, ouvrir le lexique du premier
    const lienLexique = page.locator('a[href^="/lexique?cours="]').first();
    const lexiqueVisible = await lienLexique.isVisible().catch(() => false);

    if (lexiqueVisible) {
      await lienLexique.click();
      await expect(page).toHaveURL(/\/lexique\?cours=/);

      // Verifier que le titre Lexique est visible
      await expect(page.getByRole('heading', { name: /lexique/i })).toBeVisible({ timeout: 10000 });
    }

    // 7. Naviguer vers le quiz via la sidebar
    await page.getByRole('link', { name: /^quiz$/i }).click();
    await expect(page).toHaveURL(/.*quiz/);

    // 8. Retourner au dashboard
    await page.getByRole('link', { name: /tableau de bord/i }).click();
    await expect(page).toHaveURL('/');
  });
});

test.describe('Navigation - Coherence des URLs', () => {
  test('les pages sans donnees affichent un etat vide coherent', async ({ page }) => {
    // Verifier que chaque page gere correctement l'absence de donnees
    const pages = [
      { url: '/fiches', indicateur: /h1/i },
      { url: '/quiz', indicateur: /h1/i },
      { url: '/mindmap', indicateur: /h1/i },
      { url: '/lexique', indicateur: /lexique|aucun cours/i },
    ];

    for (const p of pages) {
      await page.goto(p.url);
      await expect(page).toHaveURL(new RegExp(`.*${p.url.slice(1)}`));

      // Verifier que la page se charge sans erreur (pas de crash)
      // On verifie qu'un h1 est present (charge correctement)
      await expect(page.locator('h1').first()).toBeAttached({ timeout: 10000 });
    }
  });

  test('le logo Revisemieux ramene a l\'accueil', async ({ page }) => {
    // Aller sur une page autre que l'accueil
    await page.goto('/cours');
    await expect(page).toHaveURL(/.*cours/);

    // Cliquer sur le logo
    const logo = page.getByRole('link', { name: /revise.*mieux.*accueil/i });
    await logo.click();

    await expect(page).toHaveURL('/');
  });
});
