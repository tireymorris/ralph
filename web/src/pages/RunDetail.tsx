import { useCallback, useEffect, useRef, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import {
  cancelRun,
  createRun,
  getRun,
  openEventStream,
  postResume,
  submitClarify,
  submitFollowUp,
} from "../api/client";
import {
  LOCAL_PRD_RUN_ID,
  type EventClarifyingQuestionsPayload,
  type EventEnvelope,
} from "../api/types";
import {
  formatStatus,
  isTerminalRunStatus,
  statusBadgeClass,
} from "../lib/format";
import { makeSystemEntry, type TimelineEntry } from "../lib/timeline";
import { useAsyncSubmit } from "../hooks/useAsyncSubmit";
import { usePRDLoader } from "../hooks/usePRDLoader";
import { useRunEventStream } from "../hooks/useRunEventStream";
import { useRunPolling } from "../hooks/useRunPolling";
import { useRunStall } from "../hooks/useRunStall";
import { useTimelineScroll } from "../hooks/useTimelineScroll";
import { FORCE_RESUME_CONFIRM_MESSAGE } from "../lib/stall";
import ClarifyForm from "../components/ClarifyForm";
import FollowUpComposer from "../components/FollowUpComposer";
import GroupedTimeline from "../components/GroupedTimeline";
import ImplementationReviewPanel from "../components/ImplementationReviewPanel";
import PRDReviewPanel from "../components/PRDReviewPanel";
import RunPrompt from "../components/RunPrompt";
import StoryProgressPanel from "../components/StoryProgressPanel";

export default function RunDetail() {
  const { id } = useParams<{ id: string }>();
  const isLocalPRD = id === LOCAL_PRD_RUN_ID;
  const navigate = useNavigate();
  const { run, loadError, setRun, setLoadError } = useRunPolling(id);
  const { prd, prdError } = usePRDLoader(id, run?.status);

  const [entries, setEntries] = useState<TimelineEntry[]>([]);
  const [clarifyQuestions, setClarifyQuestions] = useState<string[]>([]);
  const clarifySubmit = useAsyncSubmit({
    fallback: "failed to submit answers",
    onSuccess: () => setClarifyQuestions([]),
  });
  const followUpSubmit = useAsyncSubmit({
    fallback: "follow-up request failed",
  });
  const cancelSubmit = useAsyncSubmit({
    fallback: "cancel failed",
    onSuccess: () => setLoadError(null),
    onError: setLoadError,
  });
  const retrySubmit = useAsyncSubmit({
    fallback: "retry failed",
    onSuccess: () => setLoadError(null),
    onError: setLoadError,
  });
  const resumeSubmit = useAsyncSubmit({
    fallback: "force resume failed",
    onSuccess: () => setLoadError(null),
    onError: setLoadError,
  });
  const [streamGeneration, setStreamGeneration] = useState(0);

  const scrollRef = useRef<HTMLDivElement>(null);
  useTimelineScroll(scrollRef, entries);

  const appendEntry = useCallback((entry: TimelineEntry) => {
    setEntries((prev) => {
      if (entry.append && prev.length > 0) {
        const last = prev[prev.length - 1];
        if (last.variant === entry.variant && last.append) {
          return [
            ...prev.slice(0, -1),
            { ...last, text: last.text + entry.text, append: true },
          ];
        }
      }
      return [...prev, entry];
    });
  }, []);

  const handleEnvelope = useCallback((envelope: EventEnvelope) => {
    if (envelope.type === "EventClarifyingQuestions") {
      const payload = envelope.payload as EventClarifyingQuestionsPayload;
      setClarifyQuestions(payload.Questions ?? []);
    }
  }, []);

  useEffect(() => {
    setEntries([]);
    setClarifyQuestions([]);
    clarifySubmit.reset();
    followUpSubmit.reset();
    cancelSubmit.reset();
    retrySubmit.reset();
    resumeSubmit.reset();
    setStreamGeneration(0);
  }, [
    id,
    clarifySubmit.reset,
    followUpSubmit.reset,
    cancelSubmit.reset,
    retrySubmit.reset,
    resumeSubmit.reset,
  ]);

  useRunEventStream(
    isLocalPRD ? undefined : id,
    openEventStream,
    appendEntry,
    handleEnvelope,
    streamGeneration,
  );

  async function handleFollowUpSubmit(message: string) {
    if (!id || followUpSubmit.submitting) return;
    await followUpSubmit.run(async () => {
      await submitFollowUp(id, message);
      setEntries((prev) => [...prev, makeSystemEntry("Follow-up accepted")]);
      setStreamGeneration((n) => n + 1);
      try {
        setRun(await getRun(id));
      } catch {
        // polling will refresh
      }
    });
  }

  async function handleCancel() {
    if (!id || cancelSubmit.submitting || !run) return;
    if (isTerminalRunStatus(run.status)) return;
    setLoadError(null);
    await cancelSubmit.run(async () => {
      await cancelRun(id);
      setRun(await getRun(id));
    }).catch(() => {});
  }

  async function handleRetry() {
    if (!run || retrySubmit.submitting) return;
    setLoadError(null);
    await retrySubmit.run(async () => {
      const { id: newId } = await createRun(run.prompt);
      navigate(`/runs/${newId}`);
    }).catch(() => {});
  }

  async function handleClarifySubmit(
    answers: { question: string; answer: string }[],
  ) {
    if (!id || clarifySubmit.submitting) return;
    await clarifySubmit
      .run(async () => {
        await submitClarify(id, answers);
      })
      .catch(() => {});
  }

  const progress = run?.story_progress;
  const isTerminal = run ? isTerminalRunStatus(run.status) : false;
  const lastTimelineId = entries.at(-1)?.id ?? "";
  const activityKey = `${run?.updated_at ?? ""}:${entries.length}:${lastTimelineId}`;
  const stalled = useRunStall(
    run?.status,
    clarifyQuestions.length,
    activityKey,
    !isLocalPRD,
  );

  async function handleForceResume() {
    if (!id || resumeSubmit.submitting || !run) return;
    if (!window.confirm(FORCE_RESUME_CONFIRM_MESSAGE)) {
      return;
    }
    setLoadError(null);
    await resumeSubmit.run(async () => {
      await postResume(id);
      setEntries((prev) => [...prev, makeSystemEntry("Force resume requested")]);
      setStreamGeneration((n) => n + 1);
      try {
        setRun(await getRun(id));
      } catch {
        // polling will refresh
      }
    }).catch(() => {});
  }

  const showStoryProgress =
    prd &&
    prd.stories.length > 0 &&
    run?.status !== "waiting_review" &&
    run?.status !== "waiting_implementation_review";

  return (
    <div
      className={`run-detail${isTerminal ? " run-detail--terminal" : ""}`}
    >
      {loadError && <p className="form-error">{loadError}</p>}
      {run && (
        <header className="run-detail-header">
          <div className="run-detail-toolbar">
            <Link to="/" className="run-detail-home">
              New
            </Link>
            <span className={`run-status-badge ${statusBadgeClass(run.status)}`}>
              {formatStatus(run.status)}
            </span>
            {run.auto_approve && (
              <span className="run-status-badge run-status-badge--default">
                Auto-approve
              </span>
            )}
            {progress && progress.total > 0 && (
              <span className="run-detail-progress-label">
                {progress.completed}/{progress.total}
              </span>
            )}
            {!isTerminal && !isLocalPRD && stalled && (
              <button
                type="button"
                className="btn btn--secondary btn--sm run-detail-force-resume"
                onClick={() => void handleForceResume()}
                disabled={resumeSubmit.submitting}
                title="Stop the stuck step and continue from saved progress"
              >
                {resumeSubmit.submitting ? "Resuming…" : "Force resume"}
              </button>
            )}
            {!isTerminal && !isLocalPRD && (
              <button
                type="button"
                className="btn btn--secondary btn--sm run-detail-cancel"
                onClick={() => void handleCancel()}
                disabled={cancelSubmit.submitting}
              >
                Cancel
              </button>
            )}
            {run.status === "cancelled" && (
              <button
                type="button"
                className="btn btn--primary btn--sm"
                onClick={() => void handleRetry()}
                disabled={retrySubmit.submitting}
              >
                Retry
              </button>
            )}
          </div>
          <RunPrompt prompt={run.prompt} />
        </header>
      )}
      <div ref={scrollRef} className="run-detail-body">
        {showStoryProgress && <StoryProgressPanel prd={prd} />}
        {run?.status === "waiting_clarify" && clarifyQuestions.length > 0 && (
          <>
            {clarifySubmit.error && (
              <p className="form-error">{clarifySubmit.error}</p>
            )}
            <ClarifyForm
              key={clarifyQuestions.join("\0")}
              questions={clarifyQuestions}
              onSubmit={(answers) => void handleClarifySubmit(answers)}
              submitting={clarifySubmit.submitting}
            />
          </>
        )}
        {run?.status === "waiting_review" && prdError && (
          <p className="form-error">{prdError}</p>
        )}
        {run?.status === "waiting_review" && prd && id && (
          <PRDReviewPanel runId={id} prd={prd} />
        )}
        {run?.status === "waiting_implementation_review" && id && (
          <ImplementationReviewPanel
            runId={id}
            iteration={run.review_iteration}
          />
        )}
        {entries.length === 0 && run && !isTerminal && isLocalPRD && (
          <p className="content-body content-muted timeline-empty">
            This run is in progress in the terminal (Ralph CLI or TUI). Continue
            there, or run <code>ralph clean</code> to reset.
          </p>
        )}
        {entries.length === 0 && run && !isTerminal && !isLocalPRD && (
          <p className="content-body content-muted timeline-empty">Waiting for events…</p>
        )}
        {entries.length > 0 && <GroupedTimeline entries={entries} />}
      </div>
      {run && isTerminal && (
        <FollowUpComposer
          onSubmit={(message) => handleFollowUpSubmit(message)}
          submitting={followUpSubmit.submitting}
          error={followUpSubmit.error}
        />
      )}
    </div>
  );
}
