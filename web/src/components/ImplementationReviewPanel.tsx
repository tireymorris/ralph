import { continueImplementationReview } from "../api/client";
import { useAsyncSubmit } from "../hooks/useAsyncSubmit";

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
  const { submitting, error, run } = useAsyncSubmit({
    fallback: "continue failed",
    onSuccess: onContinued,
  });

  async function handleContinue() {
    if (submitting) return;
    await run(async () => {
      await continueImplementationReview(runId);
    }).catch(() => {});
  }

  return (
    <section className="panel impl-review-panel" aria-label="Cleanup">
      <h2 className="content-heading">Cleanup</h2>
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
        {submitting ? "Continuing…" : "Continue"}
      </button>
    </section>
  );
}
