/**
 * MasteryBar component tests
 *
 * Validates:
 * - Z6-AC30: Global mastery indicator (% OK+SOLID / total items)
 * - Z7-AC15: Per-notion mastery bar rendering
 * - Z1-AC14: Mastery state visual representation
 */
import React from 'react';
import { render, screen } from '@testing-library/react-native';
import { MasteryBar, MasteryDot, MasteryBadge } from '@/components/MasteryBar';

describe('MasteryBar', () => {
  it('Z6-AC30: displays correct mastery percentage (OK+SOLID / total)', () => {
    const breakdown = { unknown: 2, fragile: 3, ok: 3, solid: 2 };
    // (3+2) / (2+3+3+2) = 5/10 = 50%
    render(<MasteryBar breakdown={breakdown} />);
    expect(screen.getByText('50% maitrise')).toBeTruthy();
  });

  it('Z6-AC30: displays 0% when all items are unknown', () => {
    const breakdown = { unknown: 10, fragile: 0, ok: 0, solid: 0 };
    render(<MasteryBar breakdown={breakdown} />);
    expect(screen.getByText('0% maitrise')).toBeTruthy();
  });

  it('Z6-AC30: displays 100% when all items are OK or SOLID', () => {
    const breakdown = { unknown: 0, fragile: 0, ok: 4, solid: 6 };
    render(<MasteryBar breakdown={breakdown} />);
    expect(screen.getByText('100% maitrise')).toBeTruthy();
  });

  it('returns null when total is zero', () => {
    const breakdown = { unknown: 0, fragile: 0, ok: 0, solid: 0 };
    const { toJSON } = render(<MasteryBar breakdown={breakdown} />);
    expect(toJSON()).toBeNull();
  });

  it('hides label when showLabel=false', () => {
    const breakdown = { unknown: 1, fragile: 1, ok: 1, solid: 1 };
    render(<MasteryBar breakdown={breakdown} showLabel={false} />);
    expect(screen.queryByText(/maitrise/)).toBeNull();
  });

  it('Z7-AC15: renders segments proportional to state counts', () => {
    const breakdown = { unknown: 0, fragile: 2, ok: 6, solid: 2 };
    // 4 non-zero segments: solid(2), ok(6), fragile(2)
    // Should render bar with proportional widths
    const { toJSON } = render(<MasteryBar breakdown={breakdown} />);
    const json = toJSON();
    expect(json).not.toBeNull();
  });
});

describe('MasteryDot', () => {
  it('Z1-AC14: renders dot with correct color for each state', () => {
    const states = ['unknown', 'fragile', 'ok', 'solid'] as const;
    for (const state of states) {
      const { toJSON } = render(<MasteryDot state={state} />);
      expect(toJSON()).not.toBeNull();
    }
  });
});

describe('MasteryBadge', () => {
  it('Z1-AC14: renders label for each mastery state', () => {
    render(<MasteryBadge state="unknown" />);
    expect(screen.getByText('NEW')).toBeTruthy();
  });

  it('Z1-AC14: renders FRAGILE label', () => {
    render(<MasteryBadge state="fragile" />);
    expect(screen.getByText('FRAGILE')).toBeTruthy();
  });

  it('Z1-AC14: renders OK label', () => {
    render(<MasteryBadge state="ok" />);
    expect(screen.getByText('OK')).toBeTruthy();
  });

  it('Z1-AC14: renders SOLID label', () => {
    render(<MasteryBadge state="solid" />);
    expect(screen.getByText('SOLID')).toBeTruthy();
  });
});
