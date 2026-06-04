import { useState } from "react";
import { postClean } from "../api/client";

const CONFIRM_MESSAGE =
  "Remove Ralph state from this project? This deletes prd.json and .ralph/ run data.";

export default function CleanButton() {
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function handleClick() {
    if (!window.confirm(CONFIRM_MESSAGE)) {
      return;
    }
    setBusy(true);
    setMessage(null);
    setError(null);
    try {
      await postClean();
      setMessage("Ralph state removed.");
    } catch (e) {
      setError(e instanceof Error ? e.message : "clean failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="clean-button-wrap">
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
        <p className="clean-button-message" role="status">
          {message}
        </p>
      ) : null}
      {error ? (
        <p className="form-error clean-button-error" role="alert">
          {error}
        </p>
      ) : null}
    </div>
  );
}
