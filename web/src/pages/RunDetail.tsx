import { useCallback, useEffect, useRef, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import {
  cancelRun,
  createRun,
  getRun,
  openEventStream,
  submitClarify,
  submitFollowUp,
} from "../api/client";
import type { EventClarifyingQuestionsPayload, EventEnvelope } from "../api/types";
import { formatStatus, statusBadgeClass } from "../lib/format";
import { makeSystemEntry, type TimelineEntry } from "../lib/timeline";
import { usePRDLoader } from "../hooks/usePRDLoader";
import { useRunEventStream } from "../hooks/useRunEventStream";
import { useRunPolling } from "../hooks/useRunPolling";
import { useTimelineScroll } from "../hooks/useTimelineScroll";
import ClarifyForm from "../components/ClarifyForm";
import FollowUpComposer from "../components/FollowUpComposer";
import PRDReviewPanel from "../components/PRDReviewPanel";
import TimelineEntryBubble from "../components/TimelineEntry";

const TERMINAL_RUN_STATUSES = new Set(["completed", "failed", "cancelled"]);

export default function RunDetail() {
  const { id } = useParams<{ id: string }>();
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

  useRunEventStream(id, openEventStream, appendEntry, handleEnvelope, streamGeneration);

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
    if (TERMINAL_RUN_STATUSES.has(run.status)) return;
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
  const isTerminal = run ? TERMINAL_RUN_STATUSES.has(run.status) : false;

  return (
    <div className="run-detail">
      {loadError && <p className="form-error">{loadError}</p>}
      {run && (
        <header className="run-detail-toolbar">
          <Link to="/" className="run-detail-home">
            New
          </Link>
          <p className="run-detail-prompt" title={run.prompt}>
            {run.prompt}
          </p>
          <span className={`run-status-badge ${statusBadgeClass(run.status)}`}>
            {formatStatus(run.status)}
          </span>
          {progress && progress.total > 0 && (
            <span className="run-detail-progress-label">
              {progress.completed}/{progress.total}
            </span>
          )}
          {!isTerminal && (
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
        </header>
      )}
      <div ref={scrollRef} className="run-detail-body">
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
        <ul className="run-timeline" aria-live="polite">
          {entries.length === 0 && run && !isTerminal && (
            <li className="timeline-empty">Waiting for events…</li>
          )}
          {entries.map((entry) => (
            <TimelineEntryBubble
              key={entry.id}
              variant={entry.variant}
              text={entry.text}
            />
          ))}
        </ul>
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
