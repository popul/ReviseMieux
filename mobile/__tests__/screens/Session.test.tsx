/**
 * Session screen tests
 *
 * Validates:
 * - Question display with prompt text
 * - Text input for answers
 * - Self-score flow (validate -> self-assess -> feedback)
 * - Feedback display (correct answer, hint)
 * - Debrief display (score, transitions)
 * - Navigation (quitter, voir cours)
 */
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react-native';

jest.mock('expo-router', () => ({
  router: { push: jest.fn(), replace: jest.fn(), back: jest.fn() },
  useLocalSearchParams: () => ({ id: 'chapter-1' }),
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

import { startDailySession, getQuestions, submitAnswer, getDebrief } from '@/services/api';
import SessionScreen from '@/app/session/[id]';

const mockQuestions = [
  {
    id: 'q1', template_id: 'T1', item_id: 'i1',
    rendered_prompt: 'Quelle est la formule de la masse volumique ?',
  },
  {
    id: 'q2', template_id: 'T2', item_id: 'i2',
    rendered_prompt: 'Quelle unité utilise-t-on pour la masse volumique ?',
  },
];

const mockFeedback = {
  attempt_id: 'a1', score: 1.0,
  feedback: {
    correct_answer: 'ρ = m / V',
    what_was_missing: '',
    hint: 'Pense à "rho" comme "ratio masse/volume"',
  },
};

const mockDebrief = {
  score: 2, total: 2, percentage: 100,
  transitions: [
    { item_id: 'i1', item_term: 'Formule ρ=m/V', from: 'fragile', to: 'ok' },
  ],
};

describe('Session Screen - Question View', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (startDailySession as jest.Mock).mockResolvedValue({ id: 's1', status: 'COMPOSING' });
    (getQuestions as jest.Mock).mockResolvedValue(mockQuestions);
    (submitAnswer as jest.Mock).mockResolvedValue(mockFeedback);
    (getDebrief as jest.Mock).mockResolvedValue(mockDebrief);
  });

  describe('Question display', () => {
    it('displays question prompt after loading', async () => {
      render(<SessionScreen />);
      await waitFor(() => {
        expect(screen.getByText('Quelle est la formule de la masse volumique ?')).toBeTruthy();
      });
    });

    it('shows question progress (Q 1/2)', async () => {
      render(<SessionScreen />);
      await waitFor(() => {
        expect(screen.getByText('Q 1/2')).toBeTruthy();
      });
    });

    it('shows text input for answer', async () => {
      render(<SessionScreen />);
      await waitFor(() => {
        expect(screen.getByPlaceholderText('Ta reponse...')).toBeTruthy();
      });
    });
  });

  describe('Navigation', () => {
    it('shows "Quitter" back button', async () => {
      render(<SessionScreen />);
      await waitFor(() => {
        expect(screen.getByText('← Quitter')).toBeTruthy();
      });
    });

    it('quitter navigates back', async () => {
      const { router } = require('expo-router');
      render(<SessionScreen />);
      await waitFor(() => {
        expect(screen.getByText('← Quitter')).toBeTruthy();
      });
      fireEvent.press(screen.getByText('← Quitter'));
      expect(router.back).toHaveBeenCalled();
    });

    it('shows "Voir cours" link in question view', async () => {
      render(<SessionScreen />);
      await waitFor(() => {
        expect(screen.getByText('Voir cours')).toBeTruthy();
      });
    });
  });

  describe('Self-score flow', () => {
    it('validate button leads to self-score view', async () => {
      render(<SessionScreen />);
      await waitFor(() => {
        expect(screen.getByText('Quelle est la formule de la masse volumique ?')).toBeTruthy();
      });

      // Type an answer
      fireEvent.changeText(screen.getByPlaceholderText('Ta reponse...'), 'rho = m/V');
      // Press Valider
      fireEvent.press(screen.getByText('Valider'));

      // Should show self-score buttons
      await waitFor(() => {
        expect(screen.getByText('Je savais')).toBeTruthy();
        expect(screen.getByText('Je ne savais pas')).toBeTruthy();
      });
    });

    it('pressing "Je savais" submits answer and shows feedback', async () => {
      render(<SessionScreen />);
      await waitFor(() => {
        expect(screen.getByPlaceholderText('Ta reponse...')).toBeTruthy();
      });

      // Type answer and validate
      fireEvent.changeText(screen.getByPlaceholderText('Ta reponse...'), 'rho = m/V');
      fireEvent.press(screen.getByText('Valider'));

      await waitFor(() => {
        expect(screen.getByText('Je savais')).toBeTruthy();
      });

      fireEvent.press(screen.getByText('Je savais'));

      // Should show feedback with correct answer
      await waitFor(() => {
        expect(screen.getByText('Correct !')).toBeTruthy();
        expect(screen.getByText('ρ = m / V')).toBeTruthy();
      });
    });
  });
});

describe('Session Screen - Feedback View', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (startDailySession as jest.Mock).mockResolvedValue({ id: 's1', status: 'COMPOSING' });
    (getQuestions as jest.Mock).mockResolvedValue(mockQuestions);
    (submitAnswer as jest.Mock).mockResolvedValue(mockFeedback);
    (getDebrief as jest.Mock).mockResolvedValue(mockDebrief);
  });

  async function advanceToFeedback() {
    render(<SessionScreen />);
    await waitFor(() => {
      expect(screen.getByPlaceholderText('Ta reponse...')).toBeTruthy();
    });
    fireEvent.changeText(screen.getByPlaceholderText('Ta reponse...'), 'rho = m/V');
    fireEvent.press(screen.getByText('Valider'));
    await waitFor(() => {
      expect(screen.getByText('Je savais')).toBeTruthy();
    });
    fireEvent.press(screen.getByText('Je savais'));
    await waitFor(() => {
      expect(screen.getByText('Correct !')).toBeTruthy();
    });
  }

  it('shows correct answer in feedback', async () => {
    await advanceToFeedback();
    expect(screen.getByText('ρ = m / V')).toBeTruthy();
  });

  it('shows hint with lightbulb', async () => {
    await advanceToFeedback();
    expect(
      screen.getByText(/Pense à "rho" comme "ratio masse\/volume"/)
    ).toBeTruthy();
  });

  it('shows "Suivant" button to advance', async () => {
    await advanceToFeedback();
    expect(screen.getByText('Suivant →')).toBeTruthy();
  });

  it('shows user answer section', async () => {
    await advanceToFeedback();
    expect(screen.getByText(/Ta reponse/)).toBeTruthy();
  });
});

describe('Session Screen - Debrief View', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (startDailySession as jest.Mock).mockResolvedValue({ id: 's1', status: 'COMPOSING' });
    (getQuestions as jest.Mock).mockResolvedValue([mockQuestions[0]]);
    (submitAnswer as jest.Mock).mockResolvedValue(mockFeedback);
    (getDebrief as jest.Mock).mockResolvedValue(mockDebrief);
  });

  async function advanceToDebrief() {
    render(<SessionScreen />);
    await waitFor(() => {
      expect(screen.getByPlaceholderText('Ta reponse...')).toBeTruthy();
    });
    // Answer the single question
    fireEvent.changeText(screen.getByPlaceholderText('Ta reponse...'), 'rho = m/V');
    fireEvent.press(screen.getByText('Valider'));
    await waitFor(() => {
      expect(screen.getByText('Je savais')).toBeTruthy();
    });
    fireEvent.press(screen.getByText('Je savais'));
    await waitFor(() => {
      expect(screen.getByText('Correct !')).toBeTruthy();
    });
    // Last question -> "Voir le bilan" instead of "Suivant"
    fireEvent.press(screen.getByText('Voir le bilan'));
    await waitFor(() => {
      expect(screen.getByText('🎉')).toBeTruthy();
    });
  }

  it('shows celebration emoji', async () => {
    await advanceToDebrief();
    expect(screen.getByText('🎉')).toBeTruthy();
  });

  it('shows score', async () => {
    await advanceToDebrief();
    expect(screen.getByText(/2 \/ 2/)).toBeTruthy();
  });

  it('shows percentage', async () => {
    await advanceToDebrief();
    expect(screen.getByText(/100%/)).toBeTruthy();
  });

  it('shows time elapsed', async () => {
    await advanceToDebrief();
    expect(screen.getByText(/min/)).toBeTruthy();
  });

  it('shows PROGRESSIONS section', async () => {
    await advanceToDebrief();
    expect(screen.getByText('PROGRESSIONS')).toBeTruthy();
  });

  it('shows transition item name', async () => {
    await advanceToDebrief();
    expect(screen.getByText('Formule ρ=m/V')).toBeTruthy();
  });

  it('shows "Retour au chapitre" button', async () => {
    await advanceToDebrief();
    expect(screen.getByText('Retour au chapitre')).toBeTruthy();
  });

  it('shows "Encore une session" button', async () => {
    await advanceToDebrief();
    expect(screen.getByText('Encore une session')).toBeTruthy();
  });
});
