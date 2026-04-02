/**
 * API client tests
 *
 * Validates:
 * - Request construction (headers, auth, body)
 * - Error handling (ApiError)
 * - All endpoint functions exist and call correct paths
 */
import {
  setAuthToken,
  ApiError,
  listChapters,
  getLessonCard,
  startDailySession,
  getQuestions,
  submitAnswer,
  getDebrief,
  getRevisionProgress,
  getOnboardingStatus,
  seedDemo,
} from '@/services/api';

// Mock global fetch
const mockFetch = jest.fn();
(global as any).fetch = mockFetch;

beforeEach(() => {
  mockFetch.mockReset();
  setAuthToken(null);
});

function mockJsonResponse(data: unknown, status = 200) {
  mockFetch.mockResolvedValueOnce({
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(data),
    text: () => Promise.resolve(JSON.stringify(data)),
  });
}

function mockErrorResponse(status: number, body = 'error') {
  mockFetch.mockResolvedValueOnce({
    ok: false,
    status,
    json: () => Promise.reject(new Error('not json')),
    text: () => Promise.resolve(body),
  });
}

describe('API client', () => {
  describe('Authentication', () => {
    it('sends no Authorization header when token is null', async () => {
      mockJsonResponse([]);
      await listChapters();
      const [, opts] = mockFetch.mock.calls[0];
      expect(opts.headers['Authorization']).toBeUndefined();
    });

    it('sends Bearer token when set', async () => {
      setAuthToken('test-jwt-token');
      mockJsonResponse([]);
      await listChapters();
      const [, opts] = mockFetch.mock.calls[0];
      expect(opts.headers['Authorization']).toBe('Bearer test-jwt-token');
      setAuthToken(null);
    });
  });

  describe('Error handling', () => {
    it('throws ApiError on non-OK response', async () => {
      mockErrorResponse(404, 'not found');
      await expect(listChapters()).rejects.toThrow(ApiError);
    });

    it('ApiError contains status code and body', async () => {
      mockErrorResponse(422, 'validation failed');
      try {
        await listChapters();
        throw new Error('should have thrown');
      } catch (e) {
        expect(e).toBeInstanceOf(ApiError);
        expect((e as ApiError).status).toBe(422);
        expect((e as ApiError).body).toBe('validation failed');
      }
    });
  });

  describe('Chapter endpoints', () => {
    it('listChapters calls GET /chapters', async () => {
      const chapters = [{ id: '1', name: 'Test' }];
      mockJsonResponse(chapters);
      const result = await listChapters();
      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, opts] = mockFetch.mock.calls[0];
      expect(url).toContain('/chapters');
      expect(opts.method).toBe('GET');
      expect(result).toEqual(chapters);
    });

    it('getLessonCard calls GET /chapters/:id/lesson-card', async () => {
      const data = { chapter: { id: '123' }, items: [], notions: [] };
      mockJsonResponse(data);
      const result = await getLessonCard('123');
      const [url] = mockFetch.mock.calls[0];
      expect(url).toContain('/chapters/123/lesson-card');
      expect(result).toEqual(data);
    });
  });

  describe('Session endpoints', () => {
    it('startDailySession calls POST /sessions/daily with chapter_id', async () => {
      mockJsonResponse({ id: 's1', status: 'COMPOSING' });
      await startDailySession('c1');
      const [url, opts] = mockFetch.mock.calls[0];
      expect(url).toContain('/sessions/daily');
      expect(opts.method).toBe('POST');
      const body = JSON.parse(opts.body);
      expect(body.chapter_id).toBe('c1');
    });

    it('getQuestions calls GET /sessions/:id/questions', async () => {
      mockJsonResponse([{ id: 'q1', rendered_prompt: 'What is...' }]);
      await getQuestions('s1');
      const [url] = mockFetch.mock.calls[0];
      expect(url).toContain('/sessions/s1/questions');
    });

    it('submitAnswer calls POST /sessions/:id/answer with question_id, answer and score', async () => {
      mockJsonResponse({ attempt_id: 'a1', score: 1.0 });
      await submitAnswer('s1', 'q1', 'my answer', 1.0);
      const [url, opts] = mockFetch.mock.calls[0];
      expect(url).toContain('/sessions/s1/answer');
      expect(opts.method).toBe('POST');
      const body = JSON.parse(opts.body);
      expect(body.question_id).toBe('q1');
      expect(body.answer).toBe('my answer');
      expect(body.score).toBe(1.0);
    });

    it('getDebrief calls GET /sessions/:id/debrief', async () => {
      mockJsonResponse({ score: 7, total: 10, percentage: 70 });
      await getDebrief('s1');
      const [url] = mockFetch.mock.calls[0];
      expect(url).toContain('/sessions/s1/debrief');
    });
  });

  describe('Pipeline endpoints', () => {
    it('getRevisionProgress calls GET /revisions/:id/progress', async () => {
      mockJsonResponse({ revision_id: 'r1', status: 'PROCESSING', total_pages: 3, processed_pages: 1, failed_pages: 0, total_items: 0, phase: 'ocr', phase_message: 'Processing pages' });
      await getRevisionProgress('r1');
      const [url] = mockFetch.mock.calls[0];
      expect(url).toContain('/revisions/r1/progress');
    });
  });

  describe('Onboarding endpoints', () => {
    it('getOnboardingStatus calls GET /onboarding/status', async () => {
      mockJsonResponse({ account_created: true, has_demo_chapter: false });
      await getOnboardingStatus();
      const [url] = mockFetch.mock.calls[0];
      expect(url).toContain('/onboarding/status');
    });

    it('seedDemo calls POST /onboarding/seed-demo', async () => {
      mockJsonResponse({ chapter_id: 'c1', item_count: 8 });
      await seedDemo();
      const [url, opts] = mockFetch.mock.calls[0];
      expect(url).toContain('/onboarding/seed-demo');
      expect(opts.method).toBe('POST');
    });
  });
});
