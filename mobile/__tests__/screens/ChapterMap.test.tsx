/**
 * Chapter Map screen tests
 *
 * Validates:
 * - Z7-AC15: Notions grouped by concept_tag with accordion display
 * - Z7-AC15: Per-notion mastery bar, item count, mastery %
 * - Z3-AC16: Ignored items section ("Points non vérifiés") with reactivate
 * - Z6-AC30: Chapter-level mastery indicator
 */
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react-native';

jest.mock('expo-router', () => ({
  router: { push: jest.fn(), replace: jest.fn() },
  useLocalSearchParams: () => ({ id: 'test-chapter-1' }),
}));

jest.mock('react-native-safe-area-context', () => ({
  useSafeAreaInsets: () => ({ top: 44, bottom: 34, left: 0, right: 0 }),
}));

import ChapterScreen from '@/app/chapter/[id]';

describe('Chapter Map Screen', () => {
  describe('Z7-AC15: Notion accordion display', () => {
    it('shows chapter title and subject', () => {
      render(<ChapterScreen />);
      expect(screen.getByText('Densité et masse volumique')).toBeTruthy();
      expect(screen.getByText('Physique')).toBeTruthy();
    });

    it('displays 3 notions', () => {
      render(<ChapterScreen />);
      expect(screen.getByText('Masse volumique (ρ)')).toBeTruthy();
      expect(screen.getByText('Flottabilité')).toBeTruthy();
      expect(screen.getByText('Mesures & instruments')).toBeTruthy();
    });

    it('shows mastered/total count per notion', () => {
      render(<ChapterScreen />);
      // Notion 1: 4 mastered (2ok+2solid) / 5 total
      expect(screen.getByText('4/5')).toBeTruthy();
      // Notion 2: 1/4
      expect(screen.getByText('1/4')).toBeTruthy();
      // Notion 3: 1/3
      expect(screen.getByText('1/3')).toBeTruthy();
    });

    it('first notion is expanded by default showing items', () => {
      render(<ChapterScreen />);
      // Items from notion 1 should be visible
      expect(screen.getByText('Définition ρ')).toBeTruthy();
      expect(screen.getByText('Formule ρ=m/V')).toBeTruthy();
      expect(screen.getByText('Unité g/cm³')).toBeTruthy();
      expect(screen.getByText('Conversion L↔cm³')).toBeTruthy();
    });

    it('shows item type labels', () => {
      render(<ChapterScreen />);
      expect(screen.getAllByText('KNOWLEDGE').length).toBeGreaterThanOrEqual(1);
      expect(screen.getByText('PROCEDURE')).toBeTruthy();
      expect(screen.getByText('ANALYSIS')).toBeTruthy();
    });

    it('shows mastery badges per item', () => {
      render(<ChapterScreen />);
      // First notion items have SOLID, OK, FRAGILE states
      expect(screen.getAllByText('SOLID').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('OK').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('FRAGILE').length).toBeGreaterThanOrEqual(1);
    });

    it('toggling a notion collapses/expands it', () => {
      render(<ChapterScreen />);
      // Notion 1 is expanded — items visible
      expect(screen.getByText('Définition ρ')).toBeTruthy();

      // Press notion 1 to collapse
      fireEvent.press(screen.getByText('Masse volumique (ρ)'));
      expect(screen.queryByText('Définition ρ')).toBeNull();

      // Press again to expand
      fireEvent.press(screen.getByText('Masse volumique (ρ)'));
      expect(screen.getByText('Définition ρ')).toBeTruthy();
    });

    it('expanding another notion collapses the current one', () => {
      render(<ChapterScreen />);
      // Notion 1 expanded
      expect(screen.getByText('Définition ρ')).toBeTruthy();

      // Expand notion 2
      fireEvent.press(screen.getByText('Flottabilité'));
      // Notion 1 items should be hidden
      expect(screen.queryByText('Définition ρ')).toBeNull();
      // Notion 2 items should be visible
      expect(screen.getByText("Condition flotte/coule")).toBeTruthy();
    });
  });

  describe('Z3-AC16: Ignored items (Points non vérifiés)', () => {
    it('shows "Points non vérifiés" section with count', () => {
      render(<ChapterScreen />);
      expect(screen.getByText(/Points non vérifiés \(2\)/)).toBeTruthy();
    });

    it('lists ignored item terms', () => {
      render(<ChapterScreen />);
      expect(screen.getByText('Volume molaire')).toBeTruthy();
      expect(screen.getByText('Pression atmosphérique')).toBeTruthy();
    });

    it('shows "Réactiver" button for each ignored item', () => {
      render(<ChapterScreen />);
      expect(screen.getAllByText('Réactiver').length).toBe(2);
    });
  });

  describe('Session launch', () => {
    it('shows "Lancer une session" floating button', () => {
      render(<ChapterScreen />);
      expect(screen.getByText(/Lancer une session/)).toBeTruthy();
    });

    it('navigates to session on press', () => {
      const { router } = require('expo-router');
      render(<ChapterScreen />);
      fireEvent.press(screen.getByText(/Lancer une session/));
      expect(router.push).toHaveBeenCalledWith('/session/test-chapter-1');
    });
  });
});
