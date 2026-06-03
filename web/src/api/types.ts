export interface StoryProgress {
  completed: number;
  total: number;
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
}

export interface EventClarifyingQuestionsPayload {
  Questions: string[];
}

export interface EventErrorPayload {
  error: string;
}

export interface Run {
  id: string;
  prompt: string;
  status: string;
  phase: string;
  created_at: string;
  updated_at: string;
  story_progress?: StoryProgress;
}

export interface PRDStory {
  id: string;
  title: string;
  description: string;
  passes: boolean;
}

export interface PRDDocument {
  version: number;
  project_name: string;
  stories: PRDStory[];
}
