const configuredStallThreshold = Number(
  import.meta.env.VITE_RUN_STALL_THRESHOLD_MS,
);

export const RUN_STALL_THRESHOLD_MS =
  Number.isFinite(configuredStallThreshold) && configuredStallThreshold > 0
    ? configuredStallThreshold
    : 120_000;
export const RUN_STALL_CHECK_INTERVAL_MS = 5_000;

export const FORCE_RESUME_CONFIRM_MESSAGE =
  "Resume from saved progress? This stops the current stuck step and continues the run.";

export function isRunAwaitingUser(
  status: string,
  clarifyQuestionCount: number,
): boolean {
  if (status === "waiting_review" || status === "waiting_implementation_review") {
    return true;
  }
  if (status === "waiting_clarify" && clarifyQuestionCount > 0) {
    return true;
  }
  return false;
}

export function isRunStalled(
  lastActivityMs: number,
  now = Date.now(),
  thresholdMs = RUN_STALL_THRESHOLD_MS,
): boolean {
  return now - lastActivityMs >= thresholdMs;
}
