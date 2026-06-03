import { useState } from "react";

interface FollowUpComposerProps {
  onSubmit: (message: string) => void | Promise<void>;
  submitting?: boolean;
  error?: string | null;
}

export default function FollowUpComposer({
  onSubmit,
  submitting = false,
  error = null,
}: FollowUpComposerProps) {
  const [message, setMessage] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const text = message.trim();
    if (!text || submitting) return;
    try {
      await onSubmit(text);
      setMessage("");
    } catch {
      // parent surfaces errors via error prop
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
    <form className="follow-up-composer" onSubmit={handleSubmit}>
      {error && (
        <p className="form-error" role="alert">
          {error}
        </p>
      )}
      <textarea
        className="composer-input"
        aria-label="Follow-up message"
        placeholder="Send a follow-up…"
        value={message}
        onChange={(e) => setMessage(e.target.value)}
        onKeyDown={handleKeyDown}
        disabled={submitting}
        rows={2}
      />
      <div className="follow-up-composer-actions">
        <button
          type="submit"
          className="btn btn--primary"
          disabled={submitting || !message.trim()}
        >
          Send
        </button>
        <kbd className="kbd-hint">⌘ Enter</kbd>
      </div>
    </form>
  );
}
