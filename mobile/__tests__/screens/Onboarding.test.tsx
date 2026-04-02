/**
 * Onboarding screen tests
 *
 * Validates:
 * - Z8-AC08: Deterministic onboarding sequence
 * - Z8-AC02: Guided empty state (3-step checklist)
 * - Z8-AC01: Demo chapter CTA visible in onboarding
 */
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react-native';

jest.mock('expo-router', () => ({
  router: { push: jest.fn(), replace: jest.fn() },
}));

jest.mock('react-native-safe-area-context', () => ({
  useSafeAreaInsets: () => ({ top: 44, bottom: 34, left: 0, right: 0 }),
}));

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

import { seedDemo } from '@/services/api';
import OnboardingScreen from '@/app/onboarding';

describe('Onboarding Screen', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (seedDemo as jest.Mock).mockResolvedValue(undefined);
  });

  describe('Z8-AC08: Deterministic onboarding sequence', () => {
    it('displays welcome message with user name', () => {
      render(<OnboardingScreen />);
      expect(screen.getByText(/Bienvenue Hugo/)).toBeTruthy();
    });

    it('shows 3-step checklist', () => {
      render(<OnboardingScreen />);
      expect(screen.getByText(/Compte créé/)).toBeTruthy();
      expect(screen.getByText(/Photographie ton premier cours/)).toBeTruthy();
      expect(screen.getByText(/Ta première révision/)).toBeTruthy();
    });
  });

  describe('Z8-AC02: Guided empty state', () => {
    it('step 1 is marked as done (checked)', () => {
      render(<OnboardingScreen />);
      expect(screen.getByText('✓')).toBeTruthy();
    });

    it('step 2 shows "30 secondes chrono" indication', () => {
      render(<OnboardingScreen />);
      expect(screen.getByText(/30 secondes/)).toBeTruthy();
    });

    it('step 2 has capture button', () => {
      render(<OnboardingScreen />);
      expect(screen.getByText(/Capturer/)).toBeTruthy();
    });

    it('step 3 shows "~5 min après la capture"', () => {
      render(<OnboardingScreen />);
      expect(screen.getByText(/5 min après la capture/)).toBeTruthy();
    });
  });

  describe('Z8-AC01: Demo chapter in onboarding', () => {
    it('shows demo chapter CTA with subject and duration', () => {
      render(<OnboardingScreen />);
      expect(screen.getByText(/Chapitre démo/)).toBeTruthy();
      expect(screen.getByText(/Densité/)).toBeTruthy();
      expect(screen.getByText(/8 questions/)).toBeTruthy();
      expect(screen.getByText(/3 min/)).toBeTruthy();
    });

    it('shows "ou essaie d\'abord" separator', () => {
      render(<OnboardingScreen />);
      expect(screen.getByText(/ou essaie d'abord/)).toBeTruthy();
    });

    it('has "Essayer" button for demo', () => {
      render(<OnboardingScreen />);
      expect(screen.getByText(/Essayer/)).toBeTruthy();
    });
  });

  describe('Navigation', () => {
    it('capture button navigates to capture screen', () => {
      const { router } = require('expo-router');
      render(<OnboardingScreen />);
      fireEvent.press(screen.getByText(/Capturer/));
      expect(router.replace).toHaveBeenCalledWith('/(tabs)/capture');
    });

    it('demo "Essayer" button calls seedDemo and navigates to tabs', async () => {
      const { router } = require('expo-router');
      render(<OnboardingScreen />);
      fireEvent.press(screen.getByText(/Essayer/));

      // seedDemo should be called
      expect(seedDemo).toHaveBeenCalled();

      // After seedDemo resolves, should navigate
      await waitFor(() => {
        expect(router.replace).toHaveBeenCalledWith('/(tabs)');
      });
    });
  });
});
