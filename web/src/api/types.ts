export interface StoryProgress {
  completed: number;
  total: number;
  stories?: StoryProgressStory[];
}

export interface PRDSlice {
  id: string;
  behavior: string;
  red_hint: string;
  refactor_hint?: string;
  passes: boolean;
}

export interface StoryProgressStory {
  id: string;
  title: string;
  passes: boolean;
  completed_slices: number;
  total_slices: number;
  slices?: PRDSlice[];
}

export interface CreateRunRequestOptions {
  autoApprove?: boolean;
}

export interface CreateRunResponse {
  id: string;
}

export interface QuestionAnswer {
  question: string;
  answer: string;
}

export interface EventEnvelope<T = unknown> {
  type: string;
  payload: T;
}

export interface EventOutputPayload {
  Text: string;
  IsErr: boolean;
  Verbose: boolean;
  Append?: boolean;
}

export interface EventClarifyingQuestionsPayload {
  Questions: string[];
}

export interface EventErrorPayload {
  error: string;
}

export const LOCAL_PRD_RUN_ID = "prd-local";

export type VersionStatus = "current" | "available" | "unknown";

export interface VersionInfo {
  version: string;
  commit: string;
  ref: string;
  status: VersionStatus;
  local_commit?: string;
  remote_commit?: string;
  check_error?: string;
}

export interface UpdateResult {
  status: "current" | "updated";
  message: string;
  binary?: string;
  local_commit?: string;
  remote_commit?: string;
}

export interface Run {
  id: string;
  prompt: string;
  status: string;
  phase: string;
  created_at: string;
  updated_at: string;
  source?: string;
  story_progress?: StoryProgress;
  checkpoint?: string;
  review_iteration?: number;
  review_fingerprint?: string;
  review_elapsed_ms?: number;
  stop_reason?: string;
  auto_approve?: boolean;
}

export interface PRDStory {
  id: string;
  title: string;
  description: string;
  slices: PRDSlice[];
  priority: number;
  depends_on?: string[];
  passes: boolean;
}

export interface PRDDocument {
  version: number;
  project_name: string;
  branch_name?: string;
  context?: string;
  test_spec?: string;
  test_command?: string;
  stories: PRDStory[];
}
