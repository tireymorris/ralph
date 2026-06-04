import { afterEach, describe, expect, it, vi } from "vitest";
import {
  ApiError,
  createRun,
  getRun,
  listRuns,
  openEventStream,
  postClean,
} from "./client";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("getRun", () => {
  it("throws on 404 with server error message", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
        statusText: "Not Found",
        json: async () => ({ error: "run not found" }),
      }),
    );

    await expect(getRun("missing")).rejects.toMatchObject({
      message: "run not found",
      status: 404,
    });
    await expect(getRun("missing")).rejects.toBeInstanceOf(ApiError);
  });

  it("parses error code from JSON body", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 409,
        statusText: "Conflict",
        json: async () => ({
          error: 'active run "x" in progress',
          code: "run_conflict",
        }),
      }),
    );

    await expect(createRun("goal")).rejects.toMatchObject({
      message: 'active run "x" in progress',
      status: 409,
      code: "run_conflict",
    });
  });
});

describe("listRuns", () => {
  it("returns runs from GET /api/runs", async () => {
    const runs = [
      {
        id: "a",
        prompt: "goal",
        status: "running",
        phase: "clarify",
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      },
    ];
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => runs,
    });
    vi.stubGlobal("fetch", fetchMock);

    await expect(listRuns()).resolves.toEqual(runs);
    expect(fetchMock).toHaveBeenCalledWith("/api/runs", undefined);
  });
});

describe("createRun", () => {
  it("POSTs JSON to /api/runs with Content-Type application/json", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ id: "new-run" }),
    });
    vi.stubGlobal("fetch", fetchMock);

    await expect(createRun("build feature")).resolves.toEqual({
      id: "new-run",
    });
    expect(fetchMock).toHaveBeenCalledWith("/api/runs", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ prompt: "build feature" }),
    });
  });
});

describe("postClean", () => {
  it("POSTs empty JSON to /api/clean with Content-Type application/json", async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true });
    vi.stubGlobal("fetch", fetchMock);

    await postClean();

    expect(fetchMock).toHaveBeenCalledWith("/api/clean", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({}),
    });
  });

  it("resolves without throwing on HTTP 200", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({ ok: true, status: 200 }),
    );

    await expect(postClean()).resolves.toBeUndefined();
  });

  it("throws ApiError with server status on HTTP 4xx/5xx", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        statusText: "Internal Server Error",
        json: async () => ({ error: "clean failed" }),
      }),
    );

    await expect(postClean()).rejects.toMatchObject({
      message: "clean failed",
      status: 500,
    });
    await expect(postClean()).rejects.toBeInstanceOf(ApiError);
  });
});

describe("openEventStream", () => {
  it("uses EventSource at /api/runs/{id}/events", () => {
    const eventSource = { close: vi.fn() };
    const EventSourceMock = vi.fn(() => eventSource);
    vi.stubGlobal("EventSource", EventSourceMock);

    const stream = openEventStream("run-abc");
    expect(EventSourceMock).toHaveBeenCalledWith("/api/runs/run-abc/events");
    expect(stream).toBe(eventSource);
  });
});
