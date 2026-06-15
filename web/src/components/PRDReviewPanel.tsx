import { useState } from "react";
import { submitReview } from "../api/client";
import type { PRDDocument } from "../api/types";

interface PRDReviewPanelProps {
  runId: string;
  prd: PRDDocument;
  onApproved?: () => void;
}

export default function PRDReviewPanel({
  runId,
  prd,
  onApproved,
}: PRDReviewPanelProps) {
  const [critique, setCritique] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleApprove() {
    if (submitting) return;
    setSubmitting(true);
    setError(null);
    try {
      await submitReview(runId, "approve");
      onApproved?.();
    } catch (e) {
      setError(e instanceof Error ? e.message : "approve failed");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleRevise(e: React.FormEvent) {
    e.preventDefault();
    const text = critique.trim();
    if (!text || submitting) return;
    setSubmitting(true);
    setError(null);
    try {
      await submitReview(runId, "revise", text);
      setCritique("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "revise failed");
    } finally {
      setSubmitting(false);
    }
  }

  function handleCritiqueKeyDown(e: React.KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
      e.preventDefault();
      const form = (e.target as HTMLElement).closest("form");
      form?.requestSubmit();
    }
  }

  return (
    <section className="prd-review-panel" aria-label="PRD review">
      <div className="prd-review-header">
        <h2>{prd.project_name}</h2>
        <span className="prd-review-version">v{prd.version}</span>
      </div>

      <ul className="prd-review-stories">
        {prd.stories.map((story, i) => (
          <li key={story.id}>
            <span className="prd-story-number">{i + 1}</span>
            <div>
              <strong>{story.title}</strong>
              <p>{story.description}</p>
              <ul className="prd-review-slices">
                {story.slices.map((slice) => (
                  <li key={slice.id}>
                    <p>
                      <strong>Behavior:</strong> {slice.behavior}
                    </p>
                    <p>
                      <strong>Red hint:</strong> {slice.red_hint}
                    </p>
                    {slice.refactor_hint ? (
                      <p>
                        <strong>Refactor hint:</strong> {slice.refactor_hint}
                      </p>
                    ) : null}
                  </li>
                ))}
              </ul>
            </div>
          </li>
        ))}
      </ul>

      {error && (
        <p className="form-error" role="alert">
          {error}
        </p>
      )}

      <div className="prd-review-actions">
        <button
          type="button"
          className="btn btn--primary"
          onClick={() => void handleApprove()}
          disabled={submitting}
        >
          Approve &amp; implement
        </button>
      </div>

      <form
        className="prd-review-revise-form"
        onSubmit={(e) => void handleRevise(e)}
      >
        <label className="field">
          <span className="field-label">Request changes</span>
          <textarea
            className="composer-input"
            placeholder="Describe what should change…"
            value={critique}
            onChange={(e) => setCritique(e.target.value)}
            onKeyDown={handleCritiqueKeyDown}
            disabled={submitting}
            rows={3}
          />
        </label>
        <button
          type="submit"
          className="btn btn--secondary"
          disabled={submitting || !critique.trim()}
        >
          Send revision
        </button>
      </form>
    </section>
  );
}
