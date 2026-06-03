import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { createRun } from "../api/client";

import NewRunPage from "./NewRunPage";

vi.mock("../api/client", () => ({
  createRun: vi.fn(),
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
  it("calls createRun exactly once when submitting a non-empty prompt", async () => {
    const user = userEvent.setup();
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
    renderComposer();
    const button = screen.getByRole("button", { name: /start run/i });
    expect(button).toHaveClass("btn", "btn--primary");
  });
});
