import { defineConfig, devices } from '@playwright/test';

/**
 * Configuration Playwright pour tests E2E
 * @see https://playwright.dev/docs/test-configuration
 */
export default defineConfig({
  // Dossier contenant les tests
  testDir: './e2e',

  // Timeout global pour chaque test
  timeout: 30 * 1000,

  // Timeout pour les assertions expect
  expect: {
    timeout: 5 * 1000,
  },

  // Reporter pour les résultats
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['list'],
  ],

  // Exécution en parallèle
  fullyParallel: true,

  // Échouer le build CI si des tests sont marqués test.only
  forbidOnly: !!process.env.CI,

  // Réessayer sur CI uniquement
  retries: process.env.CI ? 2 : 0,

  // Nombre de workers (parallélisme)
  workers: process.env.CI ? 1 : undefined,

  // Configuration partagée pour tous les projets
  use: {
    // URL de base pour les tests
    baseURL: 'http://localhost:3000',

    // Collecter les traces en cas d'échec
    trace: 'on-first-retry',

    // Screenshots en cas d'échec
    screenshot: 'only-on-failure',

    // Vidéo en cas d'échec
    video: 'on-first-retry',
  },

  // Projets (navigateurs) à tester
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  // Serveur web à démarrer avant les tests
  // Note: En local avec Docker (make dev), réutiliser le serveur existant
  // Le serveur Docker tourne sur localhost:3000
  webServer: {
    command: 'npm run dev -- --port 3000',
    url: 'http://localhost:3000',
    reuseExistingServer: true, // Toujours réutiliser si Docker tourne
    timeout: 120 * 1000,
  },
});
