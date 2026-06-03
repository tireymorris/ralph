import { act, cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  ApiError,
  getRun,
  openEventStream,
  submitFollowUp,
} from "../api/client";
import RunDetail from "./RunDetail";
import { resetTimelineEntryIdsForTests } from "../lib/timeline";

type MockEventSource = {
  close: ReturnType<typeof vi.fn>;
  onmessage: ((ev: MessageEvent) => void) | null;
  onerror: ((ev: Event) => void) | null;
};

let mockEventSource: MockEventSource;

vi.mock("../api/client", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api/client")>();
  return {
    ...actual,
    getRun: vi.fn(),
    submitFollowUp: vi.fn().mockResolvedValue(undefined),
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

function emitSSE(data: unknown) {
  act(() => {
    mockEventSource.onmessage?.({
      data: JSON.stringify(data),
    } as MessageEvent);
  });
}

const baseRun = {
  id: "run-1",
  prompt: "build feature",
  status: "running",
  phase: "clarify",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

beforeEach(() => {
  resetTimelineEntryIdsForTests();
});

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
  resetTimelineEntryIdsForTests();
  vi.mocked(openEventStream).mockImplementation(() => {
    mockEventSource = {
      close: vi.fn(),
      onmessage: null,
      onerror: null,
    };
    return mockEventSource as unknown as EventSource;
  });
  vi.mocked(submitFollowUp).mockResolvedValue(undefined);
});

function renderRunDetail(runId = "run-1") {
  return render(
    <MemoryRouter initialEntries={[`/runs/${runId}`]}>
      <Routes>
        <Route path="/runs/:id" element={<RunDetail />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("RunDetail", () => {
  it("renders the run prompt in the header when run data loads", async () => {
    vi.mocked(getRun).mockResolvedValue({
      ...baseRun,
      prompt: "add dark mode support",
    });

    renderRunDetail();

    await waitFor(() => {
      expect(screen.getByText("add dark mode support")).toBeInTheDocument();
    });
  });

  it("renders a color-coded status badge in the header", async () => {
    vi.mocked(getRun).mockResolvedValue({
      ...baseRun,
      status: "running",
    });

    renderRunDetail();

    await waitFor(() => {
      const badge = document.querySelector(
        ".run-detail-toolbar .run-status-badge--running",
      );
      expect(badge).not.toBeNull();
      expect(badge).toHaveTextContent("Running");
    });
  });

  it("renders 3 timeline entries from 3 SSE JSON payloads", async () => {
    vi.mocked(getRun).mockResolvedValue(baseRun);

    renderRunDetail();

    await waitFor(() => {
      expect(openEventStream).toHaveBeenCalledWith("run-1");
    });

    emitSSE({
      type: "EventOutput",
      payload: { Text: "first line", IsErr: false, Verbose: false },
    });
    emitSSE({
      type: "EventStoryStarted",
      payload: { id: "story-1", title: "Add API" },
    });
    emitSSE({
      type: "EventOutput",
      payload: { Text: "third line", IsErr: false, Verbose: false },
    });

    expect(screen.getAllByRole("listitem")).toHaveLength(3);
    expect(screen.getByText("first line")).toBeInTheDocument();
    expect(screen.getByText(/started story/i)).toBeInTheDocument();
    expect(screen.getByText("third line")).toBeInTheDocument();
  });

  describe("EventSource reconnect", () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("retries up to 3 times with 2s backoff on error", async () => {
      vi.mocked(getRun).mockResolvedValue(baseRun);
      const sources: MockEventSource[] = [];
      vi.mocked(openEventStream).mockImplementation(() => {
        const src: MockEventSource = {
          close: vi.fn(),
          onmessage: null,
          onerror: null,
        };
        sources.push(src);
        return src as unknown as EventSource;
      });

      renderRunDetail();

      await act(async () => {
        await Promise.resolve();
      });
      expect(sources).toHaveLength(1);

      for (let attempt = 0; attempt < 3; attempt++) {
        act(() => {
          sources[sources.length - 1].onerror?.({} as Event);
        });
        await act(async () => {
          await vi.advanceTimersByTimeAsync(2000);
        });
      }

      expect(openEventStream).toHaveBeenCalledTimes(4);

      act(() => {
        sources[sources.length - 1].onerror?.({} as Event);
      });
      await act(async () => {
        await vi.advanceTimersByTimeAsync(4000);
      });

      expect(openEventStream).toHaveBeenCalledTimes(4);
    });
  });

  it("truncates long EventOutput with Show more to expand", async () => {
    vi.mocked(getRun).mockResolvedValue(baseRun);
    const longText = "x".repeat(6000);

    renderRunDetail();

    await waitFor(() => {
      expect(openEventStream).toHaveBeenCalled();
    });

    emitSSE({
      type: "EventOutput",
      payload: { Text: longText, IsErr: false, Verbose: false },
    });

    const preview = longText.slice(0, 5000) + "…";
    expect(screen.getByText(preview)).toBeInTheDocument();
    expect(screen.queryByText(longText)).not.toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: /show more/i }));
    expect(screen.getByText(longText)).toBeInTheDocument();
  });

  it("shows story progress in the toolbar", async () => {
    vi.mocked(getRun).mockResolvedValue({
      ...baseRun,
      status: "implementing",
      phase: "implement",
      story_progress: { completed: 2, total: 4 },
    });

    renderRunDetail();

    await waitFor(() => {
      expect(screen.getByText("2/4")).toBeInTheDocument();
    });
  });

  it("renders follow-up composer when status is completed", async () => {
    vi.mocked(getRun).mockResolvedValue({
      ...baseRun,
      status: "completed",
      phase: "complete",
    });

    renderRunDetail();

    await waitFor(() => {
      expect(
        screen.getByRole("textbox", { name: /follow-up/i }),
      ).toBeInTheDocument();
    });
    expect(screen.getByRole("button", { name: /^send$/i })).toBeInTheDocument();
  });

  it("clears follow-up textarea after successful submitFollowUp", async () => {
    vi.mocked(getRun).mockResolvedValue({
      ...baseRun,
      status: "completed",
      phase: "complete",
    });

    renderRunDetail();

    await waitFor(() => {
      expect(
        screen.getByRole("textbox", { name: /follow-up/i }),
      ).toBeInTheDocument();
    });

    const textarea = screen.getByRole("textbox", { name: /follow-up/i });
    await userEvent.type(textarea, "add dark mode");
    await userEvent.click(screen.getByRole("button", { name: /^send$/i }));

    await waitFor(() => {
      expect(submitFollowUp).toHaveBeenCalledWith("run-1", "add dark mode");
    });
    expect(textarea).toHaveValue("");
  });

  it("displays server error message on HTTP 409 from follow-up", async () => {
    vi.mocked(getRun).mockResolvedValue({
      ...baseRun,
      status: "completed",
      phase: "complete",
    });
    vi.mocked(submitFollowUp).mockRejectedValue(
      new ApiError(409, "run is not eligible for follow-up"),
    );

    renderRunDetail();

    await waitFor(() => {
      expect(
        screen.getByRole("textbox", { name: /follow-up/i }),
      ).toBeInTheDocument();
    });

    const textarea = screen.getByRole("textbox", { name: /follow-up/i });
    await userEvent.type(textarea, "retry while running");
    await userEvent.click(screen.getByRole("button", { name: /^send$/i }));

    await waitFor(() => {
      expect(screen.getByRole("alert")).toHaveTextContent(
        "run is not eligible for follow-up",
      );
    });
    expect(textarea).toHaveValue("retry while running");
  });

  it("does not render follow-up UI when status is running", async () => {
    vi.mocked(getRun).mockResolvedValue(baseRun);

    renderRunDetail();

    await waitFor(() => {
      expect(openEventStream).toHaveBeenCalledWith("run-1");
    });

    expect(
      screen.queryByRole("textbox", { name: /follow-up/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /^send$/i }),
    ).not.toBeInTheDocument();
  });
});
