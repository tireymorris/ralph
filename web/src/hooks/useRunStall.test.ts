/** @vitest-environment jsdom */

import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { RUN_STALL_THRESHOLD_MS } from "../lib/stall";
import { useRunStall } from "./useRunStall";

describe("useRunStall", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("returns false until inactivity exceeds the stall threshold", () => {
    const { result } = renderHook(() =>
      useRunStall("implementing", 0, "activity-1", true),
    );

    expect(result.current).toBe(false);

    act(() => {
      vi.advanceTimersByTime(RUN_STALL_THRESHOLD_MS - 1);
    });
    expect(result.current).toBe(false);

    act(() => {
      vi.advanceTimersByTime(1);
    });
    expect(result.current).toBe(true);
  });

  it("resets the stall timer when activity changes", () => {
    const { result, rerender } = renderHook(
      ({ key }: { key: string }) => useRunStall("implementing", 0, key, true),
      { initialProps: { key: "a" } },
    );

    act(() => {
      vi.advanceTimersByTime(RUN_STALL_THRESHOLD_MS - 1_000);
    });
    expect(result.current).toBe(false);

    rerender({ key: "b" });

    act(() => {
      vi.advanceTimersByTime(RUN_STALL_THRESHOLD_MS - 1_000);
    });
    expect(result.current).toBe(false);
  });

  it("returns false while waiting for review", () => {
    const { result } = renderHook(() =>
      useRunStall("waiting_review", 0, "activity", true),
    );

    act(() => {
      vi.advanceTimersByTime(5_000);
    });
    expect(result.current).toBe(false);
  });

  it("returns false when disabled", () => {
    const { result } = renderHook(() =>
      useRunStall("implementing", 0, "activity", false),
    );

    act(() => {
      vi.advanceTimersByTime(5_000);
    });
    expect(result.current).toBe(false);
  });
});
