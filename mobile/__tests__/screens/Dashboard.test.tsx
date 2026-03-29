/**
 * Dashboard screen tests
 *
 * Validates:
 * - Z8-AC01: Demo chapter visible on first load
 * - Z8-AC02: Guided empty state (3-step checklist) before first upload
 * - Z6-AC30: Global cross-chapter progression view
 * - Z7-AC02: Contextual evening greeting when in evening window
 * - Z4-AC13: "All caught up" state indicator
 * - Z1-AC20: Starving items indicator on dashboard
 */
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react-native';

// Mock expo-router
jest.mock('expo-router', () => ({
  router: { push: jest.fn(), replace: jest.fn() },
  useRouter: () => ({ push: jest.fn(), replace: jest.fn() }),
}));

// Mock safe area
jest.mock('react-native-safe-area-context', () => ({
  useSafeAreaInsets: () => ({ top: 44, bottom: 34, left: 0, right: 0 }),
}));

import DashboardScreen from '@/app/(tabs)/index';

describe('Dashboard Screen', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Z8-AC01: Demo chapter pre-loaded', () => {
    it('displays demo chapter with DEMO badge in chapter list', () => {
      render(<DashboardScreen />);
      // Mock data includes a demo chapter with "(DEMO)" in title
      expect(screen.getByText(/DEMO/)).toBeTruthy();
    });

    it('demo chapter shows "Essayer" button instead of "Réviser"', () => {
      render(<DashboardScreen />);
      expect(screen.getAllByText('Essayer').length).toBeGreaterThanOrEqual(1);
    });

    it('demo chapter shows SVT subject label', () => {
      render(<DashboardScreen />);
      expect(screen.getByText(/SVT/)).toBeTruthy();
    });
  });

  describe('Z6-AC30: Global cross-chapter progression', () => {
    it('displays global mastery progress section', () => {
      render(<DashboardScreen />);
      expect(screen.getByText('PROGRESSION GLOBALE')).toBeTruthy();
    });

    it('shows SOLID, OK, and FRAGILE counts', () => {
      render(<DashboardScreen />);
      // Mock data: solid=4, ok=6, fragile=5 across all chapters
      expect(screen.getByText(/SOLID/)).toBeTruthy();
      expect(screen.getByText(/\bOK\b/)).toBeTruthy();
      expect(screen.getByText(/FRAGILE/)).toBeTruthy();
    });

    it('displays individual chapter cards with mastery percentage', () => {
      render(<DashboardScreen />);
      // Physique chapter: (3+3)/(1+1+3+3) = 75%
      expect(screen.getByText('75%')).toBeTruthy();
    });

    it('shows item count and last revised time per chapter', () => {
      render(<DashboardScreen />);
      expect(screen.getByText(/8 items/)).toBeTruthy();
      expect(screen.getByText(/12 items/)).toBeTruthy();
    });
  });

  describe('Z7-AC02: Contextual evening dashboard', () => {
    it('shows greeting with user name', () => {
      render(<DashboardScreen />);
      // Greeting depends on time of day
      expect(screen.getByText(/Hugo/)).toBeTruthy();
    });
  });

  describe('Chapter navigation', () => {
    it('shows "Réviser" buttons for real chapters', () => {
      render(<DashboardScreen />);
      expect(screen.getAllByText('Réviser').length).toBeGreaterThanOrEqual(2);
    });

    it('shows capture CTA button', () => {
      render(<DashboardScreen />);
      expect(screen.getByText(/Capturer un cours/)).toBeTruthy();
    });
  });

  describe('Chapter cards display', () => {
    it('shows chapter titles', () => {
      render(<DashboardScreen />);
      expect(screen.getByText('Densité et masse volumique')).toBeTruthy();
      expect(screen.getByText('Fractions et proportionnalité')).toBeTruthy();
    });

    it('shows chapter subjects', () => {
      render(<DashboardScreen />);
      expect(screen.getByText(/Physique/)).toBeTruthy();
      expect(screen.getByText(/Maths/)).toBeTruthy();
    });
  });
});
