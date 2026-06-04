import { beforeEach, describe, expect, it, vi } from "vitest";
import { ApiError, createRun, postClean } from "../api/client";
import {
  CONFLICT_CLEAN_CONFIRM_MESSAGE,
  isRunConflict,
  retryRunAfterClean,
} from "./clean";

vi.mock("../api/client", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api/client")>();
  return {
    ...actual,
    createRun: vi.fn(),
    postClean: vi.fn(),
  };
});

describe("isRunConflict", () => {
  it("returns true for 409 with run_conflict code", () => {
    expect(
      isRunConflict(
        new ApiError(409, 'active run "x" in progress', "run_conflict"),
      ),
    ).toBe(true);
  });

  it("returns false for 409 without run_conflict code", () => {
    expect(
      isRunConflict(new ApiError(409, "run is not eligible for follow-up")),
    ).toBe(false);
  });
});

describe("retryRunAfterClean", () => {
  beforeEach(() => {
    vi.mocked(createRun).mockReset();
    vi.mocked(postClean).mockReset();
  });

  it("returns the conflict message when confirm is declined", async () => {
    vi.stubGlobal("confirm", vi.fn(() => false));

    await expect(
      retryRunAfterClean("goal", 'active run "abc" in progress'),
    ).resolves.toEqual({
      ok: false,
      error: 'active run "abc" in progress',
    });
    expect(postClean).not.toHaveBeenCalled();
    vi.unstubAllGlobals();
  });

  it("cleans and creates a run when confirm is accepted", async () => {
    vi.stubGlobal("confirm", vi.fn(() => true));
    vi.mocked(postClean).mockResolvedValue(undefined);
    vi.mocked(createRun).mockResolvedValue({ id: "new-run" });

    await expect(
      retryRunAfterClean("goal", 'active run "abc" in progress'),
    ).resolves.toEqual({ ok: true, id: "new-run" });
    expect(vi.mocked(globalThis.confirm)).toHaveBeenCalledWith(
      CONFLICT_CLEAN_CONFIRM_MESSAGE,
    );
    expect(postClean).toHaveBeenCalledTimes(1);
    expect(createRun).toHaveBeenCalledWith("goal");
    vi.unstubAllGlobals();
  });

  it("returns postClean errors without retrying createRun", async () => {
    vi.stubGlobal("confirm", vi.fn(() => true));
    vi.mocked(postClean).mockRejectedValue(new ApiError(500, "clean failed"));

    await expect(
      retryRunAfterClean("goal", 'active run "abc" in progress'),
    ).resolves.toEqual({ ok: false, error: "clean failed" });
    expect(createRun).not.toHaveBeenCalled();
    vi.unstubAllGlobals();
  });
});
