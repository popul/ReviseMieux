import { test, expect } from '@playwright/test'

/**
 * Tests E2E pour l'overlay OCR sur la page Cours
 * - Boutons bbox (minimiser, supprimer) dans la bbox
 * - Toggle minimiser/agrandir toutes les bbox
 * - Suppression "Overlay OCR non disponible" (ne doit plus exister)
 * - Bouton × affiche le bon caractère
 */

test.describe('Cours - Overlay OCR (API réelle)', () => {
  test.describe.configure({ timeout: 120000 })
  test.skip(({ }) => !process.env.RUN_API_TESTS, 'Skipped: Backend API non disponible.')

  let coursId: string

  // Trouver un cours existant avec des blocsTexte (évite de relancer un OCR complet)
  test.beforeAll(async ({ browser }) => {
    const context = await browser.newContext()
    const page = await context.newPage()

    const response = await page.request.get('/api/cours?page=1&limite=20')
    const data = await response.json()

    const coursAvecBlocs = data.cours?.find(
      (c: { blocsTexte: unknown[]; statutOCR: string }) =>
        c.blocsTexte && c.blocsTexte.length > 0 && c.statutOCR === 'termine'
    )

    if (!coursAvecBlocs) {
      await page.close()
      test.skip(true, 'Aucun cours avec blocsTexte trouvé en base')
      return
    }

    coursId = coursAvecBlocs.id
    await page.close()
  })

  test('les boutons de bbox sont dans la bbox, pas en overflow au-dessus', async ({ page }) => {
    await page.goto(`/cours?id=${coursId}`)
    await page.waitForSelector('[data-testid="bbox-bloc"]', { timeout: 10000 })

    // Hover sur le premier bloc
    const premierBloc = page.locator('[data-testid="bbox-bloc"]').first()
    await premierBloc.hover()

    // Les boutons minimiser et supprimer doivent être visibles
    const btnMinimiser = page.locator('[data-testid="btn-minimiser-bloc"]').first()
    const btnSupprimer = page.locator('[data-testid="btn-supprimer-bloc"]').first()
    await expect(btnMinimiser).toBeVisible({ timeout: 3000 })
    await expect(btnSupprimer).toBeVisible()

    // Vérifier que le haut du bouton est >= au haut du bloc (dans la bbox, pas au-dessus)
    const blocBox = await premierBloc.boundingBox()
    const btnBox = await btnSupprimer.boundingBox()
    expect(blocBox).toBeTruthy()
    expect(btnBox).toBeTruthy()
    expect(btnBox!.y).toBeGreaterThanOrEqual(blocBox!.y - 2)
  })

  test('le bouton supprimer affiche × correctement', async ({ page }) => {
    await page.goto(`/cours?id=${coursId}`)
    await page.waitForSelector('[data-testid="bbox-bloc"]', { timeout: 10000 })

    const premierBloc = page.locator('[data-testid="bbox-bloc"]').first()
    await premierBloc.hover()

    const btnSupprimer = page.locator('[data-testid="btn-supprimer-bloc"]').first()
    await expect(btnSupprimer).toBeVisible()
    const texte = await btnSupprimer.textContent()
    expect(texte).toBe('×')
    expect(texte).not.toContain('u00d7')
  })

  test('"Masquer le texte OCR" minimise toutes les bbox, "Voir" les restaure', async ({ page }) => {
    await page.goto(`/cours?id=${coursId}`)
    await page.waitForSelector('[data-testid="bbox-bloc"]', { timeout: 10000 })

    const blocs = page.locator('[data-testid="bbox-bloc"]')
    const nbBlocs = await blocs.count()
    expect(nbBlocs).toBeGreaterThan(0)

    // Somme des hauteurs initiales (non minimisées)
    let totalHauteurAvant = 0
    for (let i = 0; i < nbBlocs; i++) {
      totalHauteurAvant += (await blocs.nth(i).boundingBox())!.height
    }
    expect(totalHauteurAvant).toBeGreaterThan(50)

    // Cliquer sur "Masquer le texte OCR"
    const toggleBtn = page.locator('[data-testid="toggle-minimiser-bbox"]')
    await expect(toggleBtn).toHaveText('Masquer le texte OCR')
    await toggleBtn.click()

    // Le label doit changer
    await expect(toggleBtn).toHaveText('Voir le texte OCR')

    // La somme des hauteurs doit être significativement réduite
    let totalHauteurApres = 0
    for (let i = 0; i < nbBlocs; i++) {
      totalHauteurApres += (await blocs.nth(i).boundingBox())!.height
    }
    expect(totalHauteurApres).toBeLessThan(totalHauteurAvant / 2)

    // Cliquer sur "Voir le texte OCR" pour restaurer
    await toggleBtn.click()
    await expect(toggleBtn).toHaveText('Masquer le texte OCR')

    let totalHauteurRestauree = 0
    for (let i = 0; i < nbBlocs; i++) {
      totalHauteurRestauree += (await blocs.nth(i).boundingBox())!.height
    }
    expect(totalHauteurRestauree).toBeGreaterThan(totalHauteurAvant / 2)
  })

  test('supprimer un bloc met à jour le contenu du cours', async ({ page }) => {
    await page.goto(`/cours?id=${coursId}`)
    await page.waitForSelector('[data-testid="bbox-bloc"]', { timeout: 10000 })

    const blocs = page.locator('[data-testid="bbox-bloc"]')
    const nbBlocsAvant = await blocs.count()

    // Lire le texte du premier bloc
    const premierBloc = blocs.first()
    const textarea = premierBloc.locator('textarea')
    const texteBlocSupprime = await textarea.inputValue()

    // Hover et supprimer
    await premierBloc.hover()
    const btnSupprimer = page.locator('[data-testid="btn-supprimer-bloc"]').first()
    await expect(btnSupprimer).toBeVisible()
    await btnSupprimer.click()

    // Un bloc en moins
    await expect(blocs).toHaveCount(nbBlocsAvant - 1)

    // Le texte du bloc supprimé ne doit plus être dans le contenu du cours
    if (texteBlocSupprime.trim().length > 10) {
      const contenuCours = page.locator('pre')
      await expect(contenuCours).not.toContainText(texteBlocSupprime.trim().slice(0, 30))
    }
  })

  test('"Overlay OCR non disponible" n\'apparait jamais', async ({ page }) => {
    await page.goto(`/cours?id=${coursId}`)
    await page.waitForSelector('h1', { timeout: 10000 })
    await expect(page.getByText('Overlay OCR non disponible')).not.toBeVisible()
  })

  test('le bouton Re-OCR apparait quand pas de blocsTexte', async ({ page }) => {
    // Chercher un cours SANS blocsTexte
    const response = await page.request.get('/api/cours?page=1&limite=20')
    const data = await response.json()
    const coursSansBlocs = data.cours?.find(
      (c: { blocsTexte: unknown[] | null; statutOCR: string; images: string[] }) =>
        (!c.blocsTexte || c.blocsTexte.length === 0) && c.statutOCR === 'termine' && c.images && c.images.length > 0
    )

    if (!coursSansBlocs) {
      test.skip(true, 'Aucun cours sans blocsTexte trouvé')
      return
    }

    await page.goto(`/cours?id=${coursSansBlocs.id}`)
    await page.waitForSelector('h1', { timeout: 10000 })

    // Pas de message "Overlay OCR non disponible"
    await expect(page.getByText('Overlay OCR non disponible')).not.toBeVisible()

    // Le bouton Re-OCR doit être visible
    const btnReocr = page.locator('[data-testid="btn-reocr"]')
    await expect(btnReocr).toBeVisible()
  })
})
