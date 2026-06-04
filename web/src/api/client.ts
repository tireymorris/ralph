import type {
  CreateRunResponse,
  PRDDocument,
  QuestionAnswer,
  Run,
  UpdateResult,
  VersionInfo,
} from "./types";

export class ApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

async function apiFetch<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, init);
  if (!res.ok) {
    let message = res.statusText;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) {
        message = body.error;
      }
    } catch {
      // ignore non-JSON error bodies
    }
    throw new ApiError(res.status, message);
  }
  return res.json() as Promise<T>;
}

async function apiPost(url: string, body: unknown): Promise<void> {
  const res = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    let message = res.statusText;
    try {
      const errBody = (await res.json()) as { error?: string };
      if (errBody.error) {
        message = errBody.error;
      }
    } catch {
      // ignore non-JSON error bodies
    }
    throw new ApiError(res.status, message);
  }
}

export async function listRuns(): Promise<Run[]> {
  return apiFetch<Run[]>("/api/runs");
}

export async function getRun(id: string): Promise<Run> {
  return apiFetch<Run>(`/api/runs/${id}`);
}

export async function getRunPRD(id: string): Promise<PRDDocument> {
  return apiFetch<PRDDocument>(`/api/runs/${id}/prd`);
}

export async function createRun(
  prompt: string,
): Promise<CreateRunResponse> {
  return apiFetch<CreateRunResponse>("/api/runs", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ prompt }),
  });
}

export function openEventStream(id: string): EventSource {
  return new EventSource(`/api/runs/${id}/events`);
}

export async function submitClarify(
  id: string,
  answers: QuestionAnswer[],
): Promise<void> {
  await apiPost(`/api/runs/${id}/clarify`, { answers });
}

export async function submitReview(
  id: string,
  action: "approve" | "revise",
  critique?: string,
): Promise<void> {
  const body: { action: string; critique?: string } = { action };
  if (critique !== undefined) {
    body.critique = critique;
  }
  await apiPost(`/api/runs/${id}/review`, body);
}

export async function submitFollowUp(
  id: string,
  message: string,
): Promise<void> {
  await apiPost(`/api/runs/${id}/followup`, { message });
}

export async function cancelRun(id: string): Promise<void> {
  await apiPost(`/api/runs/${id}/cancel`, {});
}

export async function getVersion(): Promise<VersionInfo> {
  return apiFetch<VersionInfo>("/api/version");
}

export async function postUpdate(): Promise<UpdateResult> {
  return apiFetch<UpdateResult>("/api/update", { method: "POST" });
}
