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
import { usePRDLoader } from "../hooks/usePRDLoader";
import { useRunEventStream } from "../hooks/useRunEventStream";
import { useRunPolling } from "../hooks/useRunPolling";
import { useRunStall } from "../hooks/useRunStall";
import { useTimelineScroll } from "../hooks/useTimelineScroll";
import { FORCE_RESUME_CONFIRM_MESSAGE } from "../lib/stall";
import ClarifyForm from "../components/ClarifyForm";
import FollowUpComposer from "../components/FollowUpComposer";
import GroupedTimeline from "../components/GroupedTimeline";
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
  const [clarifySubmitting, setClarifySubmitting] = useState(false);
  const [clarifyError, setClarifyError] = useState<string | null>(null);
  const [followUpSubmitting, setFollowUpSubmitting] = useState(false);
  const [followUpError, setFollowUpError] = useState<string | null>(null);
  const [cancelSubmitting, setCancelSubmitting] = useState(false);
  const [retrySubmitting, setRetrySubmitting] = useState(false);
  const [resumeSubmitting, setResumeSubmitting] = useState(false);
  const [streamGeneration, setStreamGeneration] = useState(0);

  const scrollRef = useRef<HTMLDivElement>(null);
  useTimelineScroll(scrollRef, entries);

  const appendEntry = useCallback((entry: TimelineEntry) => {
    setEntries((prev) => [...prev, entry]);
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
    setClarifyError(null);
    setFollowUpError(null);
    setStreamGeneration(0);
  }, [id]);

  useRunEventStream(
    isLocalPRD ? undefined : id,
    openEventStream,
    appendEntry,
    handleEnvelope,
    streamGeneration,
  );

  async function handleFollowUpSubmit(message: string) {
    if (!id || followUpSubmitting) return;
    setFollowUpSubmitting(true);
    setFollowUpError(null);
    try {
      await submitFollowUp(id, message);
      setEntries((prev) => [...prev, makeSystemEntry("Follow-up accepted")]);
      setStreamGeneration((n) => n + 1);
      try {
        setRun(await getRun(id));
      } catch {
        // polling will refresh
      }
    } catch (e) {
      setFollowUpError(e instanceof Error ? e.message : "follow-up request failed");
      throw e;
    } finally {
      setFollowUpSubmitting(false);
    }
  }

  async function handleCancel() {
    if (!id || cancelSubmitting || !run) return;
    if (isTerminalRunStatus(run.status)) return;
    setCancelSubmitting(true);
    setLoadError(null);
    try {
      await cancelRun(id);
      setRun(await getRun(id));
    } catch (e) {
      setLoadError(e instanceof Error ? e.message : "cancel failed");
    } finally {
      setCancelSubmitting(false);
    }
  }

  async function handleRetry() {
    if (!run || retrySubmitting) return;
    setRetrySubmitting(true);
    setLoadError(null);
    try {
      const { id: newId } = await createRun(run.prompt);
      navigate(`/runs/${newId}`);
    } catch (e) {
      setLoadError(e instanceof Error ? e.message : "retry failed");
    } finally {
      setRetrySubmitting(false);
    }
  }

  async function handleClarifySubmit(
    answers: { question: string; answer: string }[],
  ) {
    if (!id || clarifySubmitting) return;
    setClarifySubmitting(true);
    setClarifyError(null);
    try {
      await submitClarify(id, answers);
      setClarifyQuestions([]);
    } catch (e) {
      setClarifyError(e instanceof Error ? e.message : "failed to submit answers");
    } finally {
      setClarifySubmitting(false);
    }
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
    if (!id || resumeSubmitting || !run) return;
    if (!window.confirm(FORCE_RESUME_CONFIRM_MESSAGE)) {
      return;
    }
    setResumeSubmitting(true);
    setLoadError(null);
    try {
      await postResume(id);
      setEntries((prev) => [...prev, makeSystemEntry("Force resume requested")]);
      setStreamGeneration((n) => n + 1);
      try {
        setRun(await getRun(id));
      } catch {
        // polling will refresh
      }
    } catch (e) {
      setLoadError(e instanceof Error ? e.message : "force resume failed");
    } finally {
      setResumeSubmitting(false);
    }
  }

  const showStoryProgress =
    prd &&
    prd.stories.length > 0 &&
    run?.status !== "waiting_review";

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
                disabled={resumeSubmitting}
                title="Stop the stuck step and continue from saved progress"
              >
                {resumeSubmitting ? "Resuming…" : "Force resume"}
              </button>
            )}
            {!isTerminal && !isLocalPRD && (
              <button
                type="button"
                className="btn btn--secondary btn--sm run-detail-cancel"
                onClick={() => void handleCancel()}
                disabled={cancelSubmitting}
              >
                Cancel
              </button>
            )}
            {run.status === "cancelled" && (
              <button
                type="button"
                className="btn btn--primary btn--sm"
                onClick={() => void handleRetry()}
                disabled={retrySubmitting}
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
            {clarifyError && <p className="form-error">{clarifyError}</p>}
            <ClarifyForm
              key={clarifyQuestions.join("\0")}
              questions={clarifyQuestions}
              onSubmit={(answers) => void handleClarifySubmit(answers)}
              submitting={clarifySubmitting}
            />
          </>
        )}
        {run?.status === "waiting_review" && prdError && (
          <p className="form-error">{prdError}</p>
        )}
        {run?.status === "waiting_review" && prd && id && (
          <PRDReviewPanel runId={id} prd={prd} />
        )}
        {entries.length === 0 && run && !isTerminal && isLocalPRD && (
          <p className="timeline-empty">
            This run is in progress in the terminal (Ralph CLI or TUI). Continue
            there, or run <code>ralph clean</code> to reset.
          </p>
        )}
        {entries.length === 0 && run && !isTerminal && !isLocalPRD && (
          <p className="timeline-empty">Waiting for events…</p>
        )}
        {entries.length > 0 && <GroupedTimeline entries={entries} />}
      </div>
      {run && isTerminal && (
        <FollowUpComposer
          onSubmit={(message) => handleFollowUpSubmit(message)}
          submitting={followUpSubmitting}
          error={followUpError}
        />
      )}
    </div>
  );
}
