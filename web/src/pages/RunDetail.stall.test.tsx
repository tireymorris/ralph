/** @vitest-environment jsdom */

import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { getRun, openEventStream, postResume } from "../api/client";
import { FORCE_RESUME_CONFIRM_MESSAGE } from "../lib/stall";
import { resetTimelineEntryIdsForTests } from "../lib/timeline";
import RunDetail from "./RunDetail";

vi.mock("../hooks/useRunStall", () => ({
  useRunStall: vi.fn(() => false),
}));

vi.mock("../api/client", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api/client")>();
  return {
    ...actual,
    getRun: vi.fn(),
    postResume: vi.fn().mockResolvedValue(undefined),
    openEventStream: vi.fn(() => {
      mockEventSource = {
        close: vi.fn(),
        onmessage: null,
        onerror: null,
      };
      return mockEventSource as unknown as EventSource;
    }),
  };
});

type MockEventSource = {
  close: ReturnType<typeof vi.fn>;
  onmessage: ((ev: MessageEvent) => void) | null;
  onerror: ((ev: Event) => void) | null;
};

let mockEventSource: MockEventSource;

const baseRun = {
  id: "run-1",
  prompt: "build feature",
  status: "implementing",
  phase: "implement",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

function renderRunDetail() {
  return render(
    <MemoryRouter initialEntries={["/runs/run-1"]}>
      <Routes>
        <Route path="/runs/:id" element={<RunDetail />} />
      </Routes>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  resetTimelineEntryIdsForTests();
});

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
  resetTimelineEntryIdsForTests();
  vi.mocked(postResume).mockResolvedValue(undefined);
});

describe("RunDetail stall recovery", () => {
  it("does not show force resume while the run is not stalled", async () => {
    const { useRunStall } = await import("../hooks/useRunStall");
    vi.mocked(useRunStall).mockReturnValue(false);
    vi.mocked(getRun).mockResolvedValue(baseRun);

    renderRunDetail();

    await waitFor(() => {
      expect(screen.getByText("Implementing")).toBeInTheDocument();
    });

    expect(
      screen.queryByRole("button", { name: /force resume/i }),
    ).not.toBeInTheDocument();
  });

  it("calls postResume when force resume is confirmed", async () => {
    const { useRunStall } = await import("../hooks/useRunStall");
    vi.mocked(useRunStall).mockReturnValue(true);
    vi.stubGlobal("confirm", vi.fn(() => true));
    vi.mocked(getRun).mockResolvedValue(baseRun);

    renderRunDetail();

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /force resume/i })).toBeEnabled();
    });

    await userEvent.click(screen.getByRole("button", { name: /force resume/i }));

    await waitFor(() => {
      expect(postResume).toHaveBeenCalledWith("run-1");
    });
    expect(globalThis.confirm).toHaveBeenCalledWith(
      FORCE_RESUME_CONFIRM_MESSAGE,
    );
    vi.unstubAllGlobals();
  });
});
