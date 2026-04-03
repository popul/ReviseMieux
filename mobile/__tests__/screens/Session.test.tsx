/**
 * Session screen tests
 *
 * Validates:
 * - Question display with prompt text
 * - Text input for answers
 * - Auto-scoring flow (validate -> feedback directly, no self-score)
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
  getLessonCard: jest.fn().mockResolvedValue({ items: [] }),
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
    rendered_prompt: 'Quelle unite utilise-t-on pour la masse volumique ?',
  },
];

const mockFeedback = {
  attempt_id: 'a1', score: 1.0,
  feedback: {
    correct_answer: 'rho = m / V',
    what_was_missing: '',
    hint: 'Pense a rho comme ratio masse/volume',
  },
};

const mockDebrief = {
  score: 2, total: 2, percentage: 100,
  transitions: [
    { item_id: 'i1', item_term: 'Formule rho=m/V', from: 'fragile', to: 'ok' },
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
    it('shows Quitter back button', async () => {
      render(<SessionScreen />);
      await waitFor(() => {
        expect(screen.getByText(/Quitter/)).toBeTruthy();
      });
    });

    it('quitter navigates back', async () => {
      const { router } = require('expo-router');
      render(<SessionScreen />);
      await waitFor(() => {
        expect(screen.getByText(/Quitter/)).toBeTruthy();
      });
      fireEvent.press(screen.getByText(/Quitter/));
      expect(router.back).toHaveBeenCalled();
    });
  });

  describe('Auto-scoring flow', () => {
    it('validate button submits answer and shows feedback', async () => {
      render(<SessionScreen />);
      await waitFor(() => {
        expect(screen.getByPlaceholderText('Ta reponse...')).toBeTruthy();
      });

      fireEvent.changeText(screen.getByPlaceholderText('Ta reponse...'), 'rho = m/V');
      fireEvent.press(screen.getByText('Valider'));

      // Should go directly to feedback (auto-scoring, no self-score step)
      await waitFor(() => {
        expect(submitAnswer).toHaveBeenCalled();
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
      expect(submitAnswer).toHaveBeenCalled();
    });
    // Wait for feedback view to appear
    await waitFor(() => {
      expect(screen.getByText(/rho = m \/ V/)).toBeTruthy();
    });
  }

  it('shows correct answer in feedback', async () => {
    await advanceToFeedback();
    expect(screen.getByText(/rho = m \/ V/)).toBeTruthy();
  });

  it('shows hint', async () => {
    await advanceToFeedback();
    expect(screen.getByText(/ratio masse\/volume/)).toBeTruthy();
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
    fireEvent.changeText(screen.getByPlaceholderText('Ta reponse...'), 'rho = m/V');
    fireEvent.press(screen.getByText('Valider'));
    await waitFor(() => {
      expect(submitAnswer).toHaveBeenCalled();
    });
    // Single question -> should show "Voir le bilan" after feedback
    await waitFor(() => {
      expect(screen.getByText('Voir le bilan')).toBeTruthy();
    });
    fireEvent.press(screen.getByText('Voir le bilan'));
    await waitFor(() => {
      expect(getDebrief).toHaveBeenCalled();
    });
  }

  it('shows score after debrief', async () => {
    await advanceToDebrief();
    await waitFor(() => {
      expect(screen.getByText(/2 \/ 2/)).toBeTruthy();
    });
  });

  it('shows percentage', async () => {
    await advanceToDebrief();
    await waitFor(() => {
      expect(screen.getByText(/100%/)).toBeTruthy();
    });
  });

  it('shows transition item name', async () => {
    await advanceToDebrief();
    await waitFor(() => {
      expect(screen.getByText(/Formule/)).toBeTruthy();
    });
  });

  it('shows Retour au chapitre button', async () => {
    await advanceToDebrief();
    await waitFor(() => {
      expect(screen.getByText('Retour au chapitre')).toBeTruthy();
    });
  });
});
