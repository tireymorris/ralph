import { ApiError, createRun, postClean } from "../api/client";
import { errorMessage } from "./errors";

export const CLEAN_STATE_ARTIFACTS = "prd.json and .ralph/ run data";

export const CLEAN_CONFIRM_MESSAGE =
  `Remove Ralph state from this project? This deletes ${CLEAN_STATE_ARTIFACTS}.`;

export const CONFLICT_CLEAN_CONFIRM_MESSAGE =
  `An active run is blocking a new start. Remove Ralph state (${CLEAN_STATE_ARTIFACTS}) and try again?`;

export const CLEAN_SUCCESS_MESSAGE = "Ralph state removed.";

export type RunConflictRetryResult =
  | { ok: true; id: string }
  | { ok: false; error: string };

export function isRunConflict(err: unknown): err is ApiError {
  return err instanceof ApiError && err.status === 409;
}

export async function retryRunAfterClean(
  prompt: string,
  conflictMessage: string,
): Promise<RunConflictRetryResult> {
  if (!globalThis.confirm(CONFLICT_CLEAN_CONFIRM_MESSAGE)) {
    return { ok: false, error: conflictMessage };
  }
  try {
    await postClean();
    const { id } = await createRun(prompt);
    return { ok: true, id };
  } catch (err) {
    return { ok: false, error: errorMessage(err, "Failed to start run") };
  }
}
