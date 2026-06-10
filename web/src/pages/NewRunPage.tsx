import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { createRun } from "../api/client";
import RunsList from "../components/RunsList";
import { isRunConflict, retryRunAfterClean } from "../lib/clean";
import { errorMessage } from "../lib/errors";

export default function NewRunPage() {
  const navigate = useNavigate();
  const [prompt, setPrompt] = useState("");
  const [autoApprove, setAutoApprove] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    const trimmed = prompt.trim();
    if (!trimmed || submitting) {
      return;
    }
    setSubmitting(true);
    setError(null);
    const runOptions = { autoApprove };
    try {
      const { id } = await createRun(trimmed, runOptions);
      navigate(`/runs/${id}`);
    } catch (err) {
      if (isRunConflict(err)) {
        const result = await retryRunAfterClean(trimmed, err.message, runOptions);
        if (result.ok) {
          navigate(`/runs/${result.id}`);
          return;
        }
        setError(result.error);
        setSubmitting(false);
        return;
      }
      setError(errorMessage(err, "Failed to start run"));
      setSubmitting(false);
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
      e.preventDefault();
      const form = (e.target as HTMLElement).closest("form");
      form?.requestSubmit();
    }
  }

  return (
    <section className="new-run-page">
      <RunsList activeOnly heading="Active chats" hideWhenEmpty />
      <div className="new-run-main">
        <header className="new-run-header">
          <h1 className="app-wordmark">Ralph</h1>
          <p className="app-tagline">Describe a goal. Ralph plans and implements it.</p>
        </header>
        <form className="new-run-composer" onSubmit={handleSubmit}>
          <textarea
            id="goal-prompt"
            aria-label="Goal prompt"
            className="composer-input"
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            onKeyDown={handleKeyDown}
            rows={4}
            disabled={submitting}
            placeholder="What do you want to build?"
          />
          <label className="new-run-auto-approve">
            <input
              type="checkbox"
              checked={autoApprove}
              onChange={(e) => setAutoApprove(e.target.checked)}
              disabled={submitting}
            />
            Auto-approve (skip clarify and PRD review gates)
          </label>
          {error ? (
            <p className="form-error" role="alert">
              {error}
            </p>
          ) : null}
          <div className="new-run-actions">
            <button
              type="submit"
              className="btn btn--primary"
              disabled={submitting || !prompt.trim()}
            >
              Start run
            </button>
            <kbd className="kbd-hint">
              {navigator.platform?.includes("Mac") ? "⌘" : "Ctrl"} Enter
            </kbd>
          </div>
        </form>
      </div>
    </section>
  );
}
