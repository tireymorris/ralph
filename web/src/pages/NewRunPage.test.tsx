import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ApiError, createRun, listRuns, postClean } from "../api/client";
import { stubConfirm } from "../test/helpers";
import NewRunPage from "./NewRunPage";

vi.mock("../api/client", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api/client")>();
  return {
    ...actual,
    createRun: vi.fn(),
    listRuns: vi.fn(),
    postClean: vi.fn(),
  };
});

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

async function submitGoal(user: ReturnType<typeof userEvent.setup>, goal: string) {
  const textarea = screen.getByRole("textbox", { name: "Goal prompt" });
  await user.type(textarea, goal);
  await user.click(screen.getByRole("button", { name: /start run/i }));
}

describe("NewRunPage", () => {
  beforeEach(() => {
    vi.mocked(listRuns).mockReset();
    vi.mocked(createRun).mockReset();
    vi.mocked(postClean).mockReset();
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
    await submitGoal(user, "build the web ui");

    await waitFor(() => {
      expect(createRun).toHaveBeenCalledTimes(1);
    });
    expect(createRun).toHaveBeenCalledWith("build the web ui");
    await waitFor(() => {
      expect(screen.getByTestId("run-detail")).toBeInTheDocument();
    });
  });

  it("cleans state and retries when conflict confirm is accepted", async () => {
    stubConfirm(true);
    vi.mocked(listRuns).mockResolvedValue([]);
    vi.mocked(createRun)
      .mockRejectedValueOnce(
        new ApiError(409, 'active run "abc" in progress', "run_conflict"),
      )
      .mockResolvedValueOnce({ id: "new-run" });
    vi.mocked(postClean).mockResolvedValue(undefined);

    const user = userEvent.setup();
    renderComposer();
    await submitGoal(user, "my goal");

    await waitFor(() => {
      expect(postClean).toHaveBeenCalledTimes(1);
    });
    expect(createRun).toHaveBeenCalledTimes(2);
    expect(createRun).toHaveBeenNthCalledWith(1, "my goal");
    expect(createRun).toHaveBeenNthCalledWith(2, "my goal");
    await waitFor(() => {
      expect(screen.getByTestId("run-detail")).toBeInTheDocument();
    });
    vi.unstubAllGlobals();
  });

  it("shows postClean error when conflict clean fails", async () => {
    stubConfirm(true);
    vi.mocked(listRuns).mockResolvedValue([]);
    vi.mocked(createRun).mockRejectedValueOnce(
      new ApiError(409, 'active run "abc" in progress', "run_conflict"),
    );
    vi.mocked(postClean).mockRejectedValue(new ApiError(500, "clean failed"));

    const user = userEvent.setup();
    renderComposer();
    await submitGoal(user, "my goal");

    await waitFor(() => {
      expect(screen.getByRole("alert")).toHaveTextContent("clean failed");
    });
    expect(createRun).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("button", { name: /start run/i })).toBeEnabled();
    vi.unstubAllGlobals();
  });

  it("shows 409 error without calling postClean when conflict confirm is declined", async () => {
    const confirm = stubConfirm(false);
    vi.mocked(listRuns).mockResolvedValue([]);
    vi.mocked(createRun).mockRejectedValueOnce(
      new ApiError(409, 'active run "abc" in progress', "run_conflict"),
    );

    const user = userEvent.setup();
    renderComposer();
    await submitGoal(user, "my goal");

    await waitFor(() => {
      expect(confirm).toHaveBeenCalledTimes(1);
    });
    expect(postClean).not.toHaveBeenCalled();
    expect(createRun).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("alert")).toHaveTextContent(
      'active run "abc" in progress',
    );
    vi.unstubAllGlobals();
  });

  it("submit button uses btn and btn--primary classes", () => {
    vi.mocked(listRuns).mockResolvedValue([]);
    renderComposer();
    const button = screen.getByRole("button", { name: /start run/i });
    expect(button).toHaveClass("btn", "btn--primary");
  });
});
