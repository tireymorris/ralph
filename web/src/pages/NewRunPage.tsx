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
          <kbd className="kbd-hint">{navigator.platform?.includes("Mac") ? "⌘" : "Ctrl"} Enter</kbd>
        </div>
      </form>
    </section>
  );
}
