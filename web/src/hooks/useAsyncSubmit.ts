import { useCallback, useState } from "react";

import { errorMessage } from "../lib/errors";

export interface UseAsyncSubmitOptions {
  fallback?: string;
  onSuccess?: () => void;
  onError?: (message: string) => void;
}

export function useAsyncSubmit(options: UseAsyncSubmitOptions = {}) {
  const { fallback = "request failed", onSuccess, onError } = options;
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const reset = useCallback(() => {
    setError(null);
  }, []);

  const run = useCallback(
    async (fn: () => Promise<void>) => {
      setSubmitting(true);
      setError(null);
      try {
        await fn();
        onSuccess?.();
      } catch (err) {
        const message = errorMessage(err, fallback);
        setError(message);
        onError?.(message);
        throw err;
      } finally {
        setSubmitting(false);
      }
    },
    [fallback, onSuccess, onError],
  );

  return { submitting, error, run, reset };
}
