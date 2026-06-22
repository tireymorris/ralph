import { useState } from "react";
import { submitReview } from "../api/client";
import type { PRDDocument } from "../api/types";
import { useAsyncSubmit } from "../hooks/useAsyncSubmit";

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
  const approveSubmit = useAsyncSubmit({
    fallback: "approve failed",
    onSuccess: onApproved,
  });
  const reviseSubmit = useAsyncSubmit({ fallback: "revise failed" });
  const submitting = approveSubmit.submitting || reviseSubmit.submitting;
  const error = approveSubmit.error ?? reviseSubmit.error;

  async function handleApprove() {
    if (submitting) return;
    await approveSubmit
      .run(async () => {
        await submitReview(runId, "approve");
      })
      .catch(() => {});
  }

  async function handleRevise(e: React.FormEvent) {
    e.preventDefault();
    const text = critique.trim();
    if (!text || submitting) return;
    await reviseSubmit
      .run(async () => {
        await submitReview(runId, "revise", text);
        setCritique("");
      })
      .catch(() => {});
  }

  function handleCritiqueKeyDown(e: React.KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
      e.preventDefault();
      const form = (e.target as HTMLElement).closest("form");
      form?.requestSubmit();
    }
  }

  return (
    <section className="panel prd-review-panel" aria-label="PRD review">
      <div className="prd-review-header">
        <h2 className="content-heading">{prd.project_name}</h2>
        <span className="content-meta prd-review-version">v{prd.version}</span>
      </div>

      <ul className="prd-review-stories">
        {prd.stories.map((story, i) => (
          <li key={story.id}>
            <span className="prd-story-number">{i + 1}</span>
            <div>
              <strong className="content-subheading prd-review-story-title">
                {story.title}
              </strong>
              <p className="content-body">{story.description}</p>
              <ul className="slice-list prd-review-slices">
                {story.slices.map((slice) => (
                  <li key={slice.id} className="slice-item">
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
