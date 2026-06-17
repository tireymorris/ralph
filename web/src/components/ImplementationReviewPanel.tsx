import { useState } from "react";
import { continueImplementationReview } from "../api/client";

interface ImplementationReviewPanelProps {
  runId: string;
  iteration?: number;
  onContinued?: () => void;
}

export default function ImplementationReviewPanel({
  runId,
  iteration,
  onContinued,
}: ImplementationReviewPanelProps) {
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleContinue() {
    if (submitting) return;
    setSubmitting(true);
    setError(null);
    try {
      await continueImplementationReview(runId);
      onContinued?.();
    } catch (e) {
      setError(e instanceof Error ? e.message : "continue failed");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <section className="panel impl-review-panel" aria-label="Implementation review">
      <h2 className="content-heading">Implementation review</h2>
      <p className="content-body">
        Critical diff review reported findings
        {iteration != null && iteration > 0 ? ` (iteration ${iteration})` : ""}.
        Check the timeline for details, then continue when ready.
      </p>
      {error && (
        <p className="form-error" role="alert">
          {error}
        </p>
      )}
      <button
        type="button"
        className="btn btn--primary"
        onClick={() => void handleContinue()}
        disabled={submitting}
      >
        {submitting ? "Continuing…" : "Continue implementation"}
      </button>
    </section>
  );
}
