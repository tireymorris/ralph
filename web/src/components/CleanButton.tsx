import { useState } from "react";
import { postClean } from "../api/client";
import {
  CLEAN_CONFIRM_MESSAGE,
  CLEAN_SUCCESS_MESSAGE,
} from "../lib/clean";
import { errorMessage } from "../lib/errors";

export default function CleanButton() {
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function handleClick() {
    if (!window.confirm(CLEAN_CONFIRM_MESSAGE)) {
      return;
    }
    setBusy(true);
    setMessage(null);
    setError(null);
    try {
      await postClean();
      setMessage(CLEAN_SUCCESS_MESSAGE);
    } catch (e) {
      setError(errorMessage(e, "clean failed"));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="topbar-action-wrap">
      <button
        type="button"
        className="btn btn--sm btn--secondary"
        aria-label="Clean"
        onClick={() => void handleClick()}
        disabled={busy}
      >
        {busy ? "Cleaning…" : "Clean"}
      </button>
      {message ? (
        <p className="topbar-action-message" role="status">
          {message}
        </p>
      ) : null}
      {error ? (
        <p className="form-error topbar-action-error" role="alert">
          {error}
        </p>
      ) : null}
    </div>
  );
}
