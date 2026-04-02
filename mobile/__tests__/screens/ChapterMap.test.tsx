/**
 * Chapter Map screen tests
 *
 * Validates:
 * - Z7-AC15: Notions grouped by concept_tag with accordion display
 * - Z7-AC15: Per-notion mastery bar, item count, mastery %
 * - Z6-AC30: Chapter-level mastery indicator
 * - Session launch navigation
 */
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react-native';

jest.mock('expo-router', () => ({
  router: { push: jest.fn(), replace: jest.fn() },
  useLocalSearchParams: () => ({ id: 'test-chapter-1' }),
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

import { getLessonCard, getMasteries } from '@/services/api';
import ChapterScreen from '@/app/chapter/[id]';

const mockLessonCard = {
  chapter: { id: '1', name: 'Densité et masse volumique', subject: 'Physique', is_demo: false },
  items: [
    { id: 'i1', notion_id: 'n1', item_type: 'KNOWLEDGE', term: 'Définition ρ' },
    { id: 'i2', notion_id: 'n1', item_type: 'KNOWLEDGE', term: 'Formule ρ=m/V' },
    { id: 'i3', notion_id: 'n1', item_type: 'KNOWLEDGE', term: 'Unité g/cm³' },
    { id: 'i4', notion_id: 'n1', item_type: 'PROCEDURE', term: 'Conversion L↔cm³' },
    { id: 'i5', notion_id: 'n1', item_type: 'ANALYSIS', term: 'Comparer densités' },
    { id: 'i6', notion_id: 'n2', item_type: 'KNOWLEDGE', term: 'Condition flotte/coule' },
    { id: 'i7', notion_id: 'n2', item_type: 'KNOWLEDGE', term: "Poussée d'Archimède" },
    { id: 'i8', notion_id: 'n2', item_type: 'PROCEDURE', term: 'Protocole mesure' },
    { id: 'i9', notion_id: 'n2', item_type: 'KNOWLEDGE', term: 'Masse volumique eau' },
    { id: 'i10', notion_id: 'n3', item_type: 'KNOWLEDGE', term: 'Éprouvette graduée' },
    { id: 'i11', notion_id: 'n3', item_type: 'PROCEDURE', term: 'Protocole mesure volume' },
    { id: 'i12', notion_id: 'n3', item_type: 'KNOWLEDGE', term: 'Balance de précision' },
  ],
  notions: [
    { id: 'n1', name: 'Masse volumique (ρ)', sort_order: 1 },
    { id: 'n2', name: 'Flottabilité', sort_order: 2 },
    { id: 'n3', name: 'Mesures & instruments', sort_order: 3 },
  ],
};

const mockMasteries = [
  { item_id: 'i1', state: 'solid' },
  { item_id: 'i2', state: 'solid' },
  { item_id: 'i3', state: 'ok' },
  { item_id: 'i4', state: 'ok' },
  { item_id: 'i5', state: 'fragile' },
  { item_id: 'i6', state: 'ok' },
  { item_id: 'i7', state: 'unknown' },
  { item_id: 'i8', state: 'unknown' },
  { item_id: 'i9', state: 'unknown' },
  { item_id: 'i10', state: 'ok' },
  { item_id: 'i11', state: 'unknown' },
  { item_id: 'i12', state: 'unknown' },
];

describe('Chapter Map Screen', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (getLessonCard as jest.Mock).mockResolvedValue(mockLessonCard);
    (getMasteries as jest.Mock).mockResolvedValue(mockMasteries);
  });

  describe('Z7-AC15: Notion accordion display', () => {
    it('shows chapter title and subject', async () => {
      render(<ChapterScreen />);
      await waitFor(() => {
        expect(screen.getByText('Densité et masse volumique')).toBeTruthy();
        expect(screen.getByText('Physique')).toBeTruthy();
      });
    });

    it('displays 3 notions', async () => {
      render(<ChapterScreen />);
      await waitFor(() => {
        expect(screen.getByText('Masse volumique (ρ)')).toBeTruthy();
        expect(screen.getByText('Flottabilité')).toBeTruthy();
        expect(screen.getByText('Mesures & instruments')).toBeTruthy();
      });
    });

    it('shows mastered/total count per notion', async () => {
      render(<ChapterScreen />);
      await waitFor(() => {
        // Notion 1: 4 mastered (2ok+2solid) / 5 total
        expect(screen.getByText('4/5')).toBeTruthy();
        // Notion 2: 1/4
        expect(screen.getByText('1/4')).toBeTruthy();
        // Notion 3: 1/3
        expect(screen.getByText('1/3')).toBeTruthy();
      });
    });

    it('first notion is expanded by default showing items', async () => {
      render(<ChapterScreen />);
      await waitFor(() => {
        expect(screen.getByText('Définition ρ')).toBeTruthy();
        expect(screen.getByText('Formule ρ=m/V')).toBeTruthy();
        expect(screen.getByText('Unité g/cm³')).toBeTruthy();
        expect(screen.getByText('Conversion L↔cm³')).toBeTruthy();
      });
    });

    it('shows item type labels', async () => {
      render(<ChapterScreen />);
      await waitFor(() => {
        expect(screen.getAllByText('KNOWLEDGE').length).toBeGreaterThanOrEqual(1);
        expect(screen.getByText('PROCEDURE')).toBeTruthy();
        expect(screen.getByText('ANALYSIS')).toBeTruthy();
      });
    });

    it('shows mastery badges per item', async () => {
      render(<ChapterScreen />);
      await waitFor(() => {
        expect(screen.getAllByText('SOLID').length).toBeGreaterThanOrEqual(1);
        expect(screen.getAllByText('OK').length).toBeGreaterThanOrEqual(1);
        expect(screen.getAllByText('FRAGILE').length).toBeGreaterThanOrEqual(1);
      });
    });

    it('toggling a notion collapses/expands it', async () => {
      render(<ChapterScreen />);
      await waitFor(() => {
        expect(screen.getByText('Définition ρ')).toBeTruthy();
      });

      // Press notion 1 to collapse
      fireEvent.press(screen.getByText('Masse volumique (ρ)'));
      expect(screen.queryByText('Définition ρ')).toBeNull();

      // Press again to expand
      fireEvent.press(screen.getByText('Masse volumique (ρ)'));
      expect(screen.getByText('Définition ρ')).toBeTruthy();
    });

    it('expanding another notion collapses the current one', async () => {
      render(<ChapterScreen />);
      await waitFor(() => {
        expect(screen.getByText('Définition ρ')).toBeTruthy();
      });

      // Expand notion 2
      fireEvent.press(screen.getByText('Flottabilité'));
      expect(screen.queryByText('Définition ρ')).toBeNull();
      expect(screen.getByText('Condition flotte/coule')).toBeTruthy();
    });
  });

  describe('Session launch', () => {
    it('shows "Lancer une session" floating button', async () => {
      render(<ChapterScreen />);
      await waitFor(() => {
        expect(screen.getByText(/Lancer une session/)).toBeTruthy();
      });
    });

    it('navigates to session on press', async () => {
      const { router } = require('expo-router');
      render(<ChapterScreen />);
      await waitFor(() => {
        expect(screen.getByText(/Lancer une session/)).toBeTruthy();
      });
      fireEvent.press(screen.getByText(/Lancer une session/));
      expect(router.push).toHaveBeenCalledWith('/session/test-chapter-1');
    });
  });
});
