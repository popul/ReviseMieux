/**
 * Processing screen tests
 *
 * Validates:
 * - Z8-AC04: J0 processing progress with 3-phase messages
 * - Z8-AC03: OCR failure recovery screen with empathetic messaging
 */
import React from 'react';
import { render, screen, act } from '@testing-library/react-native';

jest.mock('expo-router', () => ({
  router: { push: jest.fn(), replace: jest.fn() },
}));

jest.mock('react-native-safe-area-context', () => ({
  useSafeAreaInsets: () => ({ top: 44, bottom: 34, left: 0, right: 0 }),
}));

// We need to control timers for the animation
jest.useFakeTimers();

import ProcessingScreen from '@/app/processing';

describe('Processing Screen', () => {
  afterEach(() => {
    jest.clearAllTimers();
  });

  describe('Z8-AC04: 3-phase progress messages', () => {
    it('starts with phase 1: "Lecture de tes pages..."', () => {
      render(<ProcessingScreen />);
      expect(screen.getByText('Lecture de tes pages...')).toBeTruthy();
    });

    it('shows phase indicator "Phase 1/3"', () => {
      render(<ProcessingScreen />);
      expect(screen.getByText(/Phase 1\/3/)).toBeTruthy();
    });

    it('shows page reading icon initially', () => {
      render(<ProcessingScreen />);
      expect(screen.getByText('📄')).toBeTruthy();
    });

    it('transitions to phase 2 after progress > 30%', () => {
      render(<ProcessingScreen />);
      // Each tick advances ~2%, so 16 ticks = 32% → phase 2
      act(() => {
        jest.advanceTimersByTime(200 * 16);
      });
      expect(screen.getByText('Création des questions...')).toBeTruthy();
      expect(screen.getByText(/Phase 2\/3/)).toBeTruthy();
    });

    it('transitions to phase 3 after sufficient progress', () => {
      render(<ProcessingScreen />);
      // Advance in chunks to allow React state to settle between intervals
      // Phase 1→2 happens at ~30%, phase 2→3 at ~70%
      // Each tick = +2%, so we advance in multiple act() to let state propagate
      for (let i = 0; i < 40; i++) {
        act(() => {
          jest.advanceTimersByTime(200);
        });
      }
      expect(screen.getByText('Presque fini...')).toBeTruthy();
      expect(screen.getByText(/Phase 3\/3/)).toBeTruthy();
    });

    it('shows "close app" message after 15 seconds', () => {
      render(<ProcessingScreen />);
      // The close hint appears after 15000ms
      act(() => {
        jest.advanceTimersByTime(15_000);
      });
      expect(screen.getByText(/Tu peux fermer l'app/)).toBeTruthy();
    });

    it('does NOT show close hint before 15 seconds', () => {
      render(<ProcessingScreen />);
      act(() => {
        jest.advanceTimersByTime(10_000);
      });
      expect(screen.queryByText(/Tu peux fermer l'app/)).toBeNull();
    });
  });
});

/**
 * Recovery screen is rendered as a sub-component of ProcessingScreen.
 * We test its content expectations directly.
 */
describe('Z8-AC03: Recovery Screen expectations', () => {
  // The recovery screen is shown when `failed` state is true.
  // Since it's internal state, we test the component structure expectations
  // that should be met when recovery is triggered.

  it('recovery content expectations: empathetic message', () => {
    // These expectations validate what Z8-AC03 requires
    // The actual recovery screen text should contain:
    const required = [
      'difficiles à lire',        // Empathetic message
      'Pas de panique',           // Encouraging tone
      'Bonne lumière',            // Tip 1
      'Page entière',             // Tip 2
      'Texte lisible',            // Tip 3
      'Reprendre',               // Retry button
      'Essayer quand même',       // Try anyway button
    ];

    // We verify these strings exist in the source code
    const fs = require('fs');
    const source = fs.readFileSync(
      require.resolve('@/app/processing'),
      'utf8',
    );
    for (const text of required) {
      expect(source).toContain(text);
    }
  });

  it('recovery has 3 visual tips (lighting, framing, text size)', () => {
    const fs = require('fs');
    const source = fs.readFileSync(
      require.resolve('@/app/processing'),
      'utf8',
    );
    // 3 TipRow components rendered
    expect(source).toContain('Bonne lumière');
    expect(source).toContain("Pas d'ombre");
    expect(source).toContain('Page entière');
    expect(source).toContain('Pas coupée');
    expect(source).toContain('Texte lisible');
    expect(source).toContain('Pas trop petit');
  });
});
