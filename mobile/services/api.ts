import Constants from 'expo-constants';

const BASE_URL =
  Constants.expoConfig?.extra?.apiUrl ?? 'http://localhost:8080/api/v1';

type RequestOptions = {
  method?: string;
  body?: unknown;
  headers?: Record<string, string>;
};

let authToken: string | null = null;

export function setAuthToken(token: string | null) {
  authToken = token;
}

async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...opts.headers,
  };
  if (authToken) {
    headers['Authorization'] = `Bearer ${authToken}`;
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    method: opts.method ?? 'GET',
    headers,
    body: opts.body ? JSON.stringify(opts.body) : undefined,
  });

  if (!res.ok) {
    const text = await res.text().catch(() => '');
    throw new ApiError(res.status, text);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export class ApiError extends Error {
  constructor(
    public status: number,
    public body: string,
  ) {
    super(`API ${status}: ${body}`);
  }
}

// --- Chapter ---

export type Chapter = {
  id: string;
  subject: string;
  title: string;
  is_demo: boolean;
  mastery_breakdown: { unknown: number; fragile: number; ok: number; solid: number };
  item_count: number;
  last_revised_at: string | null;
  created_at: string;
};

export type Notion = {
  id: string;
  name: string;
  items: Item[];
  mastery_breakdown: { unknown: number; fragile: number; ok: number; solid: number };
};

export type Item = {
  id: string;
  type: string;
  term: string;
  definition: string;
  mastery_state: string;
  fidelity_flag: string | null;
};

export function listChapters() {
  return request<Chapter[]>('/chapters');
}

export function getChapter(id: string) {
  return request<{ chapter: Chapter; notions: Notion[] }>(`/chapters/${id}`);
}

// --- Session ---

export type Session = {
  id: string;
  type: string;
  status: string;
  chapter_ids: string[];
  total_questions: number;
  answered: number;
  score: number | null;
  created_at: string;
};

export type Question = {
  id: string;
  item_id: string;
  type: string;
  prompt: string;
  choices?: string[];
  unit?: string;
  context?: string;
};

export type AttemptResult = {
  correct: boolean;
  score: number;
  score_class: string;
  expected_answer: string;
  hint: string;
  mastery_transition: { from: string; to: string } | null;
};

export type SessionDebrief = {
  score: number;
  total: number;
  percentage: number;
  duration_seconds: number;
  transitions: { item_term: string; from: string; to: string }[];
  to_consolidate: string[];
};

export function startSession(chapterIds: string[], type: string = 'daily') {
  return request<Session>('/sessions', {
    method: 'POST',
    body: { chapter_ids: chapterIds, type },
  });
}

export function getNextQuestion(sessionId: string) {
  return request<Question | null>(`/sessions/${sessionId}/next`);
}

export function submitAnswer(sessionId: string, questionId: string, answer: string) {
  return request<AttemptResult>(`/sessions/${sessionId}/questions/${questionId}/answer`, {
    method: 'POST',
    body: { answer },
  });
}

export function getSessionDebrief(sessionId: string) {
  return request<SessionDebrief>(`/sessions/${sessionId}/debrief`);
}

// --- Pipeline ---

export type PipelineStatus = {
  status: 'processing' | 'completed' | 'failed';
  phase: number;
  progress: number;
  chapter_id: string | null;
  recovery?: {
    message: string;
    can_retry: boolean;
    items_generated: number;
  };
};

export function uploadPages(chapterId: string | null, subject: string, formData: FormData) {
  return fetch(`${BASE_URL}/pipeline/upload`, {
    method: 'POST',
    headers: authToken ? { Authorization: `Bearer ${authToken}` } : {},
    body: formData,
  }).then((res) => {
    if (!res.ok) throw new ApiError(res.status, 'Upload failed');
    return res.json() as Promise<{ pipeline_id: string }>;
  });
}

export function getPipelineStatus(pipelineId: string) {
  return request<PipelineStatus>(`/pipeline/${pipelineId}/status`);
}

// --- Onboarding ---

export type OnboardingStatus = {
  current_step: string;
  completed_steps: string[];
  has_demo: boolean;
  has_chapters: boolean;
  has_sessions: boolean;
};

export function getOnboardingStatus() {
  return request<OnboardingStatus>('/onboarding/status');
}
