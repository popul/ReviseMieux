import Constants from 'expo-constants';

// Use Mac's local IP for iPhone testing, fallback to localhost
const BASE_URL =
  Constants.expoConfig?.extra?.apiUrl ?? 'http://192.168.1.19:8080/api/v1';

const DEV_TOKEN_URL =
  Constants.expoConfig?.extra?.apiUrl
    ? Constants.expoConfig.extra.apiUrl.replace('/api/v1', '/dev/token')
    : 'http://192.168.1.19:8080/dev/token';

type RequestOptions = {
  method?: string;
  body?: unknown;
  headers?: Record<string, string>;
};

let authToken: string | null = null;

export function setAuthToken(token: string | null) {
  authToken = token;
}

export function getAuthToken(): string | null {
  return authToken;
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

// --- Dev Auth ---

export type DevTokenResponse = {
  token: string;
  user_id: string;
};

export async function fetchDevToken(): Promise<DevTokenResponse> {
  const res = await fetch(DEV_TOKEN_URL);
  if (!res.ok) throw new ApiError(res.status, 'Failed to get dev token');
  return res.json();
}

// --- Chapter ---

export type MasteryBreakdown = {
  unknown: number;
  fragile: number;
  ok: number;
  solid: number;
};

export type Chapter = {
  id: string;
  user_id: string;
  subject: string;
  class_level: string;
  name: string;
  current_revision_id?: string;
  archived: boolean;
  is_demo: boolean;
  item_count: number;
  mastery_breakdown?: MasteryBreakdown;
};

export type ItemResponse = {
  id: string;
  chapter_id: string;
  notion_id?: string;
  item_type: string;
  term?: string;
  confidence: number;
  validation_required: boolean;
  archived: boolean;
  keywords?: string[];
};

export type NotionResponse = {
  id: string;
  chapter_id: string;
  name: string;
  sort_order: number;
};

export type LessonCardResponse = {
  chapter: Chapter;
  items: ItemResponse[];
  notions: NotionResponse[];
};

export function listChapters() {
  return request<Chapter[]>('/chapters');
}

export function getLessonCard(chapterId: string) {
  return request<LessonCardResponse>(`/chapters/${chapterId}/lesson-card`);
}

// --- Mastery ---

export type MasteryResponse = {
  id: string;
  user_id: string;
  item_id: string;
  state: string;
  consecutive_successes: number;
  next_due_at?: string;
};

export function getMasteries(state?: string) {
  const query = state ? `?state=${state}` : '';
  return request<MasteryResponse[]>(`/masteries${query}`);
}

export function getMasteryForItem(itemId: string) {
  return request<MasteryResponse>(`/masteries/${itemId}`);
}

// --- Session ---

export type Session = {
  id: string;
  user_id: string;
  session_type: string;
  status: string;
  trigger_type: string;
  started_at?: string;
  completed_at?: string;
  current_question_index: number;
  chapter_ids?: string[];
};

export type Question = {
  id: string;
  template_id: string;
  item_id: string;
  question_type: string;
  rendered_prompt: string;
  choices?: string[];
  visual_url?: string;
};

export type SubmitAnswerResponse = {
  attempt_id: string;
  score: number;
  feedback?: {
    correct_answer: string;
    what_was_missing: string;
    hint: string;
  };
};

export type DebriefResponse = {
  score: number;
  total: number;
  percentage: number;
  transitions: { item_id: string; item_term: string; from: string; to: string }[];
};

export function startDailySession(chapterId: string) {
  return request<Session>('/sessions/daily', {
    method: 'POST',
    body: { chapter_id: chapterId },
  });
}

export function getSession(sessionId: string) {
  return request<Session>(`/sessions/${sessionId}`);
}

export function getQuestions(sessionId: string) {
  return request<Question[]>(`/sessions/${sessionId}/questions`);
}

export function submitAnswer(sessionId: string, questionId: string, answer: string, score: number) {
  return request<SubmitAnswerResponse>(`/sessions/${sessionId}/answer`, {
    method: 'POST',
    body: { question_id: questionId, answer, score },
  });
}

export function getDebrief(sessionId: string) {
  return request<DebriefResponse>(`/sessions/${sessionId}/debrief`);
}

// --- Onboarding ---

export type OnboardingStatus = {
  account_created: boolean;
  demo_session_done: boolean;
  first_chapter_ready: boolean;
  has_demo_chapter: boolean;
  demo_chapter_id?: string;
  step1_label: string;
  step2_label: string;
  step3_label: string;
};

export type SeedDemoResponse = {
  chapter_id: string;
  item_count: number;
  message: string;
};

export function getOnboardingStatus() {
  return request<OnboardingStatus>('/onboarding/status');
}

export function seedDemo() {
  return request<SeedDemoResponse>('/onboarding/seed-demo', { method: 'POST' });
}

// --- Pipeline ---

export type PipelineProgress = {
  revision_id: string;
  status: 'PROCESSING' | 'READY' | 'PARTIAL' | 'FAILED';
  total_pages: number;
  processed_pages: number;
  failed_pages: number;
  total_items: number;
  phase: string;
  phase_message: string;
};

export function uploadPages(chapterId: string | null, subject: string, formData: FormData) {
  return fetch(`${BASE_URL}/chapters/${chapterId}/upload`, {
    method: 'POST',
    headers: authToken ? { Authorization: `Bearer ${authToken}` } : {},
    body: formData,
  }).then((res) => {
    if (!res.ok) throw new ApiError(res.status, 'Upload failed');
    return res.json() as Promise<{ revision_id: string }>;
  });
}

export function getRevisionProgress(revisionId: string) {
  return request<PipelineProgress>(`/revisions/${revisionId}/progress`);
}
