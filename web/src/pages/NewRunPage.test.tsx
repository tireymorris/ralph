import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createRun, listRuns } from "../api/client";

import NewRunPage from "./NewRunPage";

vi.mock("../api/client", () => ({
  createRun: vi.fn(),
  listRuns: vi.fn(),
}));

function renderComposer() {
  return render(
    <MemoryRouter initialEntries={["/"]}>
      <Routes>
        <Route path="/" element={<NewRunPage />} />
        <Route path="/runs/:id" element={<div data-testid="run-detail" />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("NewRunPage", () => {
  beforeEach(() => {
    vi.mocked(listRuns).mockReset();
    vi.mocked(createRun).mockReset();
  });

  it("shows active chats when non-terminal runs exist", async () => {
    vi.mocked(listRuns).mockResolvedValue([
      {
        id: "run-active",
        prompt: "Ship the API",
        status: "waiting_review",
        phase: "review",
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      },
    ]);

    renderComposer();

    await waitFor(() => {
      expect(screen.getByRole("heading", { name: "Active chats" })).toBeInTheDocument();
    });
    expect(screen.getByText("Ship the API")).toBeInTheDocument();
  });

  it("calls createRun exactly once when submitting a non-empty prompt", async () => {
    const user = userEvent.setup();
    vi.mocked(listRuns).mockResolvedValue([]);
    vi.mocked(createRun).mockResolvedValue({ id: "run-1" });

    renderComposer();

    const textarea = screen.getByRole("textbox", { name: "Goal prompt" });
    await user.type(textarea, "build the web ui");
    await user.click(screen.getByRole("button", { name: /start run/i }));

    await waitFor(() => {
      expect(createRun).toHaveBeenCalledTimes(1);
    });
    expect(createRun).toHaveBeenCalledWith("build the web ui");
    await waitFor(() => {
      expect(screen.getByTestId("run-detail")).toBeInTheDocument();
    });
  });

  it("submit button uses btn and btn--primary classes", () => {
    vi.mocked(listRuns).mockResolvedValue([]);
    renderComposer();
    const button = screen.getByRole("button", { name: /start run/i });
    expect(button).toHaveClass("btn", "btn--primary");
  });
});
