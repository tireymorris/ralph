import { useCallback, useEffect, useState } from "react";
import { getVersion, postUpdate } from "../api/client";
import type { VersionInfo } from "../api/types";

const POLL_MS = 5 * 60 * 1000;

type UpdateUIState = "loading" | VersionInfo["status"] | "updating";

export default function UpdateButton() {
  const [state, setState] = useState<UpdateUIState>("loading");
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    try {
      const info = await getVersion();
      setState(info.status);
      setError(null);
    } catch (e) {
      setState("unknown");
      setError(e instanceof Error ? e.message : "failed to check version");
    }
  }, []);

  useEffect(() => {
    void refresh();
    const timer = setInterval(() => void refresh(), POLL_MS);
    return () => clearInterval(timer);
  }, [refresh]);

  async function handleClick() {
    if (state !== "available") {
      return;
    }
    if (
      !window.confirm(
        "Update Ralph from GitHub? This may take a minute. Restart ralph web afterward.",
      )
    ) {
      return;
    }
    setState("updating");
    setMessage(null);
    setError(null);
    try {
      const result = await postUpdate();
      setMessage(result.message);
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "update failed");
      setState("available");
    }
  }

  const label =
    state === "loading"
      ? "Update"
      : state === "updating"
        ? "Updating…"
        : state === "available"
          ? "Update available"
          : state === "current"
            ? "Up to date"
            : "Update";

  const className = [
    "btn",
    "btn--sm",
    "btn--update",
    state === "available" && "btn--update-available",
    state === "current" && "btn--update-current",
    (state === "unknown" || state === "loading") && "btn--update-unknown",
    state === "updating" && "btn--update-updating",
  ]
    .filter(Boolean)
    .join(" ");

  const title =
    message ??
    error ??
    (state === "available"
      ? "A newer Ralph release is available"
      : state === "current"
        ? "Ralph is up to date"
        : state === "unknown"
          ? "Cannot check for updates from this build"
          : undefined);

  return (
    <div className="topbar-action-wrap">
      <button
        type="button"
        className={className}
        onClick={() => void handleClick()}
        disabled={state !== "available"}
        title={title}
        aria-label={label}
      >
        {label}
      </button>
      {message ? (
        <p className="topbar-action-message" role="status">
          {message}
        </p>
      ) : null}
      {error && state === "available" ? (
        <p className="form-error topbar-action-error" role="alert">
          {error}
        </p>
      ) : null}
    </div>
  );
}
