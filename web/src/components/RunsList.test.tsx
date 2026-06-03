import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { listRuns } from "../api/client";
import RunsList from "./RunsList";

vi.mock("../api/client", () => ({
  listRuns: vi.fn(),
}));

describe("RunsList", () => {
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
