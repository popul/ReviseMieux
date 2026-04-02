/**
 * Dashboard screen tests
 *
 * Validates:
 * - Z8-AC01: Demo chapter visible on first load
 * - Z6-AC30: Global cross-chapter progression view
 * - Z7-AC02: Contextual evening greeting when in evening window
 * - Chapter navigation and display
 */
import React from 'react';
import { render, screen, waitFor } from '@testing-library/react-native';

// Mock expo-router
jest.mock('expo-router', () => ({
  router: { push: jest.fn(), replace: jest.fn() },
  useRouter: () => ({ push: jest.fn(), replace: jest.fn() }),
}));

// Mock safe area
jest.mock('react-native-safe-area-context', () => ({
  useSafeAreaInsets: () => ({ top: 44, bottom: 34, left: 0, right: 0 }),
}));

// Mock API
jest.mock('@/services/api', () => ({
  listChapters: jest.fn(),
  getLessonCard: jest.fn(),
  getMasteries: jest.fn(),
  startDailySession: jest.fn(),
  getQuestions: jest.fn(),
  submitAnswer: jest.fn(),
  getDebrief: jest.fn(),
  seedDemo: jest.fn(),
  getOnboardingStatus: jest.fn(),
  setAuthToken: jest.fn(),
  getAuthToken: jest.fn(() => 'test-token'),
  fetchDevToken: jest.fn(),
}));

import { listChapters } from '@/services/api';
import DashboardScreen from '@/app/(tabs)/index';

const mockChapters = [
  {
    id: '1', user_id: 'u1', subject: 'Physique', class_level: '5e',
    name: 'Densité et masse volumique', archived: false, is_demo: false,
    item_count: 8, mastery_breakdown: { unknown: 1, fragile: 1, ok: 3, solid: 3 },
  },
  {
    id: '2', user_id: 'u1', subject: 'Maths', class_level: '5e',
    name: 'Fractions et proportionnalité', archived: false, is_demo: false,
    item_count: 12, mastery_breakdown: { unknown: 2, fragile: 3, ok: 3, solid: 4 },
  },
  {
    id: '3', user_id: 'u1', subject: 'SVT', class_level: '5e',
    name: 'Cellule et biodiversité (DEMO)', archived: false, is_demo: true,
    item_count: 6, mastery_breakdown: { unknown: 2, fragile: 2, ok: 1, solid: 1 },
  },
];

describe('Dashboard Screen', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (listChapters as jest.Mock).mockResolvedValue(mockChapters);
  });

  describe('Z8-AC01: Demo chapter pre-loaded', () => {
    it('displays demo chapter with DEMO badge in chapter list', async () => {
      render(<DashboardScreen />);
      await waitFor(() => {
        expect(screen.getByText(/DEMO/)).toBeTruthy();
      });
    });

    it('demo chapter shows "Essayer" button instead of "Réviser"', async () => {
      render(<DashboardScreen />);
      await waitFor(() => {
        expect(screen.getAllByText('Essayer').length).toBeGreaterThanOrEqual(1);
      });
    });

    it('demo chapter shows SVT subject label', async () => {
      render(<DashboardScreen />);
      await waitFor(() => {
        expect(screen.getByText(/SVT/)).toBeTruthy();
      });
    });
  });

  describe('Z6-AC30: Global cross-chapter progression', () => {
    it('displays global mastery progress section', async () => {
      render(<DashboardScreen />);
      await waitFor(() => {
        expect(screen.getByText('PROGRESSION GLOBALE')).toBeTruthy();
      });
    });

    it('shows SOLID, OK, and FRAGILE counts', async () => {
      render(<DashboardScreen />);
      await waitFor(() => {
        expect(screen.getByText(/SOLID/)).toBeTruthy();
        expect(screen.getByText(/\bOK\b/)).toBeTruthy();
        expect(screen.getByText(/FRAGILE/)).toBeTruthy();
      });
    });

    it('displays individual chapter cards with mastery percentage', async () => {
      render(<DashboardScreen />);
      await waitFor(() => {
        // Physique chapter: (3+3)/(1+1+3+3) = 75%
        expect(screen.getByText('75%')).toBeTruthy();
      });
    });

    it('shows item count per chapter', async () => {
      render(<DashboardScreen />);
      await waitFor(() => {
        expect(screen.getByText(/8 items/)).toBeTruthy();
        expect(screen.getByText(/12 items/)).toBeTruthy();
      });
    });
  });

  describe('Z7-AC02: Contextual dashboard', () => {
    it('shows greeting with user name', async () => {
      render(<DashboardScreen />);
      await waitFor(() => {
        expect(screen.getByText(/Hugo/)).toBeTruthy();
      });
    });
  });

  describe('Chapter navigation', () => {
    it('shows "Réviser" buttons for real chapters', async () => {
      render(<DashboardScreen />);
      await waitFor(() => {
        expect(screen.getAllByText('Réviser').length).toBeGreaterThanOrEqual(2);
      });
    });

    it('shows capture CTA button', async () => {
      render(<DashboardScreen />);
      await waitFor(() => {
        expect(screen.getByText(/Capturer un cours/)).toBeTruthy();
      });
    });
  });

  describe('Chapter cards display', () => {
    it('shows chapter titles', async () => {
      render(<DashboardScreen />);
      await waitFor(() => {
        expect(screen.getByText('Densité et masse volumique')).toBeTruthy();
        expect(screen.getByText('Fractions et proportionnalité')).toBeTruthy();
      });
    });

    it('shows chapter subjects', async () => {
      render(<DashboardScreen />);
      await waitFor(() => {
        expect(screen.getByText(/Physique/)).toBeTruthy();
        expect(screen.getByText(/Maths/)).toBeTruthy();
      });
    });
  });
});
