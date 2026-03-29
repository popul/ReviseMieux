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
  getChapter,
  startSession,
  getNextQuestion,
  submitAnswer,
  getSessionDebrief,
  getPipelineStatus,
  getOnboardingStatus,
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
      const chapters = [{ id: '1', title: 'Test' }];
      mockJsonResponse(chapters);
      const result = await listChapters();
      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, opts] = mockFetch.mock.calls[0];
      expect(url).toContain('/chapters');
      expect(opts.method).toBe('GET');
      expect(result).toEqual(chapters);
    });

    it('getChapter calls GET /chapters/:id', async () => {
      const data = { chapter: { id: '123' }, notions: [] };
      mockJsonResponse(data);
      const result = await getChapter('123');
      const [url] = mockFetch.mock.calls[0];
      expect(url).toContain('/chapters/123');
      expect(result).toEqual(data);
    });
  });

  describe('Session endpoints', () => {
    it('startSession calls POST /sessions with chapter_ids and type', async () => {
      mockJsonResponse({ id: 's1', status: 'in_progress' });
      await startSession(['c1', 'c2'], 'daily');
      const [url, opts] = mockFetch.mock.calls[0];
      expect(url).toContain('/sessions');
      expect(opts.method).toBe('POST');
      const body = JSON.parse(opts.body);
      expect(body.chapter_ids).toEqual(['c1', 'c2']);
      expect(body.type).toBe('daily');
    });

    it('getNextQuestion calls GET /sessions/:id/next', async () => {
      mockJsonResponse({ id: 'q1', prompt: 'What is...' });
      await getNextQuestion('s1');
      const [url] = mockFetch.mock.calls[0];
      expect(url).toContain('/sessions/s1/next');
    });

    it('submitAnswer calls POST with answer body', async () => {
      mockJsonResponse({ correct: true, score: 1.0 });
      await submitAnswer('s1', 'q1', 'my answer');
      const [url, opts] = mockFetch.mock.calls[0];
      expect(url).toContain('/sessions/s1/questions/q1/answer');
      expect(opts.method).toBe('POST');
      const body = JSON.parse(opts.body);
      expect(body.answer).toBe('my answer');
    });

    it('getSessionDebrief calls GET /sessions/:id/debrief', async () => {
      mockJsonResponse({ score: 7, total: 10 });
      await getSessionDebrief('s1');
      const [url] = mockFetch.mock.calls[0];
      expect(url).toContain('/sessions/s1/debrief');
    });
  });

  describe('Pipeline endpoints', () => {
    it('getPipelineStatus calls GET /pipeline/:id/status', async () => {
      mockJsonResponse({ status: 'processing', phase: 2, progress: 0.5 });
      await getPipelineStatus('p1');
      const [url] = mockFetch.mock.calls[0];
      expect(url).toContain('/pipeline/p1/status');
    });
  });

  describe('Onboarding endpoint', () => {
    it('getOnboardingStatus calls GET /onboarding/status', async () => {
      mockJsonResponse({ current_step: 'demo_available', completed_steps: ['account_created'] });
      await getOnboardingStatus();
      const [url] = mockFetch.mock.calls[0];
      expect(url).toContain('/onboarding/status');
    });
  });
});
