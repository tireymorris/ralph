import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { listRuns } from "../api/client";
import RunsList from "./RunsList";

vi.mock("../api/client", () => ({
  listRuns: vi.fn(),
}));

describe("RunsList", () => {
  beforeEach(() => {
    vi.mocked(listRuns).mockReset();
  });

  it("hides active-only list when all runs are terminal", async () => {
    vi.mocked(listRuns).mockResolvedValue([
      {
        id: "run-done",
        prompt: "Old goal",
        status: "completed",
        phase: "complete",
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      },
    ]);

    const { container } = render(
      <MemoryRouter>
        <RunsList activeOnly hideWhenEmpty />
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(listRuns).toHaveBeenCalled();
    });
    expect(container).toBeEmptyDOMElement();
  });

  it("shows local_prd CLI tag on sidebar entry", async () => {
    vi.mocked(listRuns).mockResolvedValue([
      {
        id: "prd-local",
        prompt: "CLI project",
        status: "implementing",
        phase: "implement",
        source: "local_prd",
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      },
    ]);

    render(
      <MemoryRouter>
        <RunsList activeOnly heading="Active chats" />
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(screen.getByText("CLI")).toHaveClass("run-source-tag");
    });
    expect(screen.getByText("CLI project")).toBeInTheDocument();
  });

  it("shows only non-terminal runs when activeOnly", async () => {
    vi.mocked(listRuns).mockResolvedValue([
      {
        id: "run-active",
        prompt: "In progress",
        status: "running",
        phase: "implement",
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      },
      {
        id: "run-done",
        prompt: "Finished",
        status: "completed",
        phase: "complete",
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      },
    ]);

    render(
      <MemoryRouter>
        <RunsList activeOnly heading="Active chats" />
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(screen.getByText("In progress")).toBeInTheDocument();
    });
    expect(screen.getByRole("heading", { name: "Active chats" })).toBeInTheDocument();
    expect(screen.queryByText("Finished")).not.toBeInTheDocument();
    expect(screen.queryAllByRole("listitem")).toHaveLength(1);
  });

  it("renders 0 rows when API returns []", async () => {
    vi.mocked(listRuns).mockResolvedValue([]);

    render(
      <MemoryRouter>
        <RunsList />
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(screen.queryAllByRole("listitem")).toHaveLength(0);
    });
    expect(screen.getByText(/no runs yet/i)).toBeInTheDocument();
  });

  it("renders a running status badge with run-status-badge--running", async () => {
    vi.mocked(listRuns).mockResolvedValue([
      {
        id: "run-1",
        prompt: "Build a feature",
        status: "running",
        phase: "implement",
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      },
    ]);

    render(
      <MemoryRouter>
        <RunsList />
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(screen.getByText("Running")).toBeInTheDocument();
    });

    const badge = screen.getByText("Running");
    expect(badge).toHaveClass("run-status-badge");
    expect(badge).toHaveClass("run-status-badge--running");
  });

  it.each([
    ["waiting_clarify", "run-status-badge--waiting", "Needs Answers"],
    ["waiting_review", "run-status-badge--waiting", "Needs Review"],
    ["completed", "run-status-badge--completed", "Completed"],
    ["failed", "run-status-badge--failed", "Failed"],
    ["cancelled", "run-status-badge--cancelled", "Cancelled"],
    ["unknown_status", "run-status-badge--default", "unknown_status"],
  ])(
    "maps status %s to badge class %s",
    async (status, expectedClass, expectedLabel) => {
      vi.mocked(listRuns).mockResolvedValue([
        {
          id: "run-1",
          prompt: "Build a feature",
          status,
          phase: "implement",
          created_at: "2026-01-01T00:00:00Z",
          updated_at: "2026-01-01T00:00:00Z",
        },
      ]);

      render(
        <MemoryRouter>
          <RunsList />
        </MemoryRouter>,
      );

      await waitFor(() => {
        expect(screen.getByText(expectedLabel)).toBeInTheDocument();
      });

      const badge = screen.getByText(expectedLabel);
      expect(badge).toHaveClass("run-status-badge");
      expect(badge).toHaveClass(expectedClass);
    },
  );
});
