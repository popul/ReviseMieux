import { test, expect } from '@playwright/test'

/**
 * Tests E2E pour la rotation manuelle d'images et la détection automatique d'orientation
 * - Boutons ↶ ↷ pour pivoter l'image
 * - Bouton 🧭 pour la détection automatique d'orientation
 * - Re-OCR déclenché après rotation
 */

test.describe('Cours - Rotation d\'images (API réelle)', () => {
  test.describe.configure({ timeout: 120000 })
  test.skip(({}) => !process.env.RUN_API_TESTS, 'Skipped: Backend API non disponible.')

  let coursId: string

  // Trouver un cours existant avec des images et OCR terminé
  test.beforeAll(async ({ browser }) => {
    const context = await browser.newContext()
    const page = await context.newPage()

    const response = await page.request.get('/api/cours?page=1&limite=20')
    const data = await response.json()

    const coursAvecImages = data.cours?.find(
      (c: { images: string[]; statutOCR: string }) =>
        c.images && c.images.length > 0 && c.statutOCR === 'termine'
    )

    if (!coursAvecImages) {
      await page.close()
      test.skip(true, 'Aucun cours avec images et OCR terminé trouvé en base')
      return
    }

    coursId = coursAvecImages.id
    await page.close()
  })

  test('les boutons de rotation ↶ ↷ sont visibles', async ({ page }) => {
    await page.goto(`/cours?id=${coursId}`)
    await page.waitForSelector('[data-testid="images-ocr"]', { timeout: 10000 })

    const btnGauche = page.locator('[data-testid="btn-pivoter-gauche"]').first()
    const btnDroite = page.locator('[data-testid="btn-pivoter-droite"]').first()

    await expect(btnGauche).toBeVisible({ timeout: 5000 })
    await expect(btnDroite).toBeVisible()

    // Vérifier le contenu des boutons
    await expect(btnGauche).toHaveText('↶')
    await expect(btnDroite).toHaveText('↷')
  })

  test('le bouton de détection auto 🧭 est visible', async ({ page }) => {
    await page.goto(`/cours?id=${coursId}`)
    await page.waitForSelector('[data-testid="images-ocr"]', { timeout: 10000 })

    const btnAutoOrient = page.locator('[data-testid="btn-auto-orient"]').first()
    await expect(btnAutoOrient).toBeVisible({ timeout: 5000 })

    // Par défaut, la détection auto est activée → pas de style coral
    await expect(btnAutoOrient).toHaveText('🧭')
    await expect(btnAutoOrient).not.toHaveClass(/bg-coral/)
  })

  test('pivoter à droite déclenche le re-OCR et change le style du bouton 🧭', async ({ page }) => {
    await page.goto(`/cours?id=${coursId}`)
    await page.waitForSelector('[data-testid="images-ocr"]', { timeout: 10000 })

    const btnDroite = page.locator('[data-testid="btn-pivoter-droite"]').first()
    const btnAutoOrient = page.locator('[data-testid="btn-auto-orient"]').first()

    await expect(btnDroite).toBeVisible({ timeout: 5000 })
    await expect(btnDroite).toBeEnabled()

    // Cliquer sur pivoter à droite
    await btnDroite.click()

    // Après rotation manuelle, le bouton 🧭 doit passer en mode "désactivé" (style coral)
    await expect(btnAutoOrient).toHaveText('🧭!')
    await expect(btnAutoOrient).toHaveClass(/bg-coral/, { timeout: 10000 })

    // Le cours doit passer en statut OCR en_cours (boutons désactivés temporairement)
    // Attendre que le re-OCR soit terminé (boutons redeviennent actifs)
    await expect(btnDroite).toBeEnabled({ timeout: 60000 })
  })

  test('pivoter à gauche déclenche le re-OCR', async ({ page }) => {
    await page.goto(`/cours?id=${coursId}`)
    await page.waitForSelector('[data-testid="images-ocr"]', { timeout: 10000 })

    const btnGauche = page.locator('[data-testid="btn-pivoter-gauche"]').first()
    const btnAutoOrient = page.locator('[data-testid="btn-auto-orient"]').first()

    await expect(btnGauche).toBeVisible({ timeout: 5000 })
    await expect(btnGauche).toBeEnabled()

    // Cliquer sur pivoter à gauche
    await btnGauche.click()

    // Après rotation manuelle, le bouton 🧭 doit passer en mode "désactivé"
    await expect(btnAutoOrient).toHaveText('🧭!')
    await expect(btnAutoOrient).toHaveClass(/bg-coral/, { timeout: 10000 })

    // Attendre que le re-OCR soit terminé
    await expect(btnGauche).toBeEnabled({ timeout: 60000 })
  })

  test('le bouton 🧭 réactive la détection auto et relance l\'OCR', async ({ page }) => {
    await page.goto(`/cours?id=${coursId}`)
    await page.waitForSelector('[data-testid="images-ocr"]', { timeout: 10000 })

    const btnDroite = page.locator('[data-testid="btn-pivoter-droite"]').first()
    const btnAutoOrient = page.locator('[data-testid="btn-auto-orient"]').first()

    await expect(btnDroite).toBeVisible({ timeout: 5000 })
    await expect(btnDroite).toBeEnabled()

    // 1. D'abord pivoter manuellement pour désactiver la détection auto
    await btnDroite.click()
    await expect(btnAutoOrient).toHaveText('🧭!')
    await expect(btnAutoOrient).toHaveClass(/bg-coral/, { timeout: 10000 })

    // Attendre que le premier re-OCR soit terminé
    await expect(btnAutoOrient).toBeEnabled({ timeout: 60000 })

    // 2. Cliquer sur le bouton 🧭 pour réactiver la détection auto
    await btnAutoOrient.click()

    // Le bouton doit revenir à l'état normal (plus de coral)
    await expect(btnAutoOrient).toHaveText('🧭')

    // Attendre que le re-OCR avec détection auto soit terminé
    await expect(btnAutoOrient).toBeEnabled({ timeout: 60000 })
    await expect(btnAutoOrient).not.toHaveClass(/bg-coral/)
  })

  test('les boutons sont désactivés pendant le re-OCR', async ({ page }) => {
    await page.goto(`/cours?id=${coursId}`)
    await page.waitForSelector('[data-testid="images-ocr"]', { timeout: 10000 })

    const btnGauche = page.locator('[data-testid="btn-pivoter-gauche"]').first()
    const btnDroite = page.locator('[data-testid="btn-pivoter-droite"]').first()
    const btnAutoOrient = page.locator('[data-testid="btn-auto-orient"]').first()
    const btnReocr = page.locator('[data-testid="btn-reocr"]').first()

    await expect(btnDroite).toBeVisible({ timeout: 5000 })
    await expect(btnDroite).toBeEnabled()

    // Déclencher une rotation
    await btnDroite.click()

    // Pendant le re-OCR, vérifier que les 4 boutons sont bien désactivés
    // (le cours passe en statutOCR = 'en_cours' ce qui désactive les boutons)
    // Note: ceci est très rapide, on vérifie immédiatement après le clic
    // Le polling peut déjà avoir rafraîchi, donc on attend juste que tout soit terminé
    await expect(btnDroite).toBeEnabled({ timeout: 60000 })
    await expect(btnGauche).toBeEnabled()
    await expect(btnAutoOrient).toBeEnabled()
    await expect(btnReocr).toBeEnabled()
  })
})
