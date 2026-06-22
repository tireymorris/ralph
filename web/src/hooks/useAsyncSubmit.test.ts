/** @vitest-environment jsdom */

import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { useAsyncSubmit } from "./useAsyncSubmit";

describe("useAsyncSubmit", () => {
  it("resets submitting in finally after rejection", async () => {
    const { result } = renderHook(() => useAsyncSubmit());

    await act(async () => {
      await expect(
        result.current.run(async () => {
          throw new Error("fail");
        }),
      ).rejects.toThrow("fail");
    });

    expect(result.current.submitting).toBe(false);
  });

  it("sets submitting true while the async fn is in flight", async () => {
    let resolvePending!: () => void;
    const pending = new Promise<void>((resolve) => {
      resolvePending = resolve;
    });

    const { result } = renderHook(() => useAsyncSubmit());

    let runPromise!: Promise<void>;
    act(() => {
      runPromise = result.current.run(async () => pending);
    });

    expect(result.current.submitting).toBe(true);

    await act(async () => {
      resolvePending();
      await runPromise;
    });

    expect(result.current.submitting).toBe(false);
  });

  it("clears error on start and captures failures via errorMessage", async () => {
    const { result } = renderHook(() =>
      useAsyncSubmit({ fallback: "submit failed" }),
    );

    await act(async () => {
      await expect(
        result.current.run(async () => {
          throw new Error("first");
        }),
      ).rejects.toThrow("first");
    });
    expect(result.current.error).toBe("first");

    await act(async () => {
      await expect(
        result.current.run(async () => {
          throw "plain";
        }),
      ).rejects.toBe("plain");
    });
    expect(result.current.error).toBe("submit failed");

    await act(async () => {
      await result.current.run(async () => {});
    });
    expect(result.current.error).toBeNull();
  });

  it("reset clears the error", async () => {
    const { result } = renderHook(() => useAsyncSubmit());

    await act(async () => {
      await expect(
        result.current.run(async () => {
          throw new Error("oops");
        }),
      ).rejects.toThrow("oops");
    });
    expect(result.current.error).toBe("oops");

    act(() => {
      result.current.reset();
    });

    expect(result.current.error).toBeNull();
  });

  it("calls optional onSuccess and onError callbacks", async () => {
    const onSuccess = vi.fn();
    const onError = vi.fn();
    const { result } = renderHook(() =>
      useAsyncSubmit({ onSuccess, onError, fallback: "failed" }),
    );

    await act(async () => {
      await result.current.run(async () => {});
    });
    expect(onSuccess).toHaveBeenCalledOnce();
    expect(onError).not.toHaveBeenCalled();

    await act(async () => {
      await expect(
        result.current.run(async () => {
          throw new Error("bad");
        }),
      ).rejects.toThrow("bad");
    });
    expect(onError).toHaveBeenCalledWith("bad");
  });
});
