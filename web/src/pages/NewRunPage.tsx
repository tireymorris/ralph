import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { createRun } from "../api/client";

export default function NewRunPage() {
  const navigate = useNavigate();
  const [prompt, setPrompt] = useState("");
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
    try {
      const { id } = await createRun(trimmed);
      navigate(`/runs/${id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to start run");
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
      <h1 className="new-run-heading">New run</h1>
      <p className="new-run-hint">
        Describe what you want to build. Ralph will generate a PRD, ask
        clarifying questions if needed, then implement story by story.
      </p>
      <form className="new-run-composer" onSubmit={handleSubmit}>
        <label htmlFor="goal-prompt" className="field-label">
          Goal
        </label>
        <textarea
          id="goal-prompt"
          aria-label="Goal prompt"
          className="composer-input"
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          onKeyDown={handleKeyDown}
          rows={6}
          disabled={submitting}
          placeholder="Describe your goal…"
        />
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
          <kbd className="kbd-hint">⌘ Enter</kbd>
        </div>
      </form>
    </section>
  );
}
