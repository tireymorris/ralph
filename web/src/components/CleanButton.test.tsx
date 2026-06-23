import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { MemoryRouter } from "react-router-dom";
import { ApiError, postClean } from "../api/client";
import { CLEAN_SUCCESS_MESSAGE } from "../lib/clean";
import { stubConfirm } from "../test/helpers";
import CleanButton from "./CleanButton";

function renderCleanButton() {
  return render(
    <MemoryRouter>
      <CleanButton />
    </MemoryRouter>,
  );
}

vi.mock("../api/client", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api/client")>();
  return {
    ...actual,
    postClean: vi.fn(),
  };
});

describe("CleanButton", () => {
  beforeEach(() => {
    vi.mocked(postClean).mockReset();
  });

  it("renders a Clean button in the header", () => {
    renderCleanButton();
    expect(screen.getByRole("button", { name: "Clean" })).toBeInTheDocument();
  });

  it("asks to confirm removal of prd.json and .ralph/ before cleaning", async () => {
    const confirm = stubConfirm(false);
    const user = userEvent.setup();
    renderCleanButton();

    await user.click(screen.getByRole("button", { name: "Clean" }));

    expect(confirm).toHaveBeenCalledWith(
      expect.stringMatching(/prd\.json/i),
    );
    expect(confirm).toHaveBeenCalledWith(
      expect.stringMatching(/\.ralph\//),
    );
    vi.unstubAllGlobals();
  });

  it("does not call postClean when confirm is declined", async () => {
    stubConfirm(false);
    const user = userEvent.setup();
    renderCleanButton();

    await user.click(screen.getByRole("button", { name: "Clean" }));

    expect(postClean).not.toHaveBeenCalled();
    vi.unstubAllGlobals();
  });

  it("calls postClean and shows success when confirm is accepted", async () => {
    vi.mocked(postClean).mockResolvedValue(undefined);
    stubConfirm(true);
    const user = userEvent.setup();
    renderCleanButton();

    await user.click(screen.getByRole("button", { name: "Clean" }));

    await waitFor(() => {
      expect(postClean).toHaveBeenCalledTimes(1);
      expect(screen.getByRole("status")).toHaveTextContent(
        CLEAN_SUCCESS_MESSAGE,
      );
    });
    vi.unstubAllGlobals();
  });

  it("shows alert with ApiError message when postClean fails", async () => {
    vi.mocked(postClean).mockRejectedValue(
      new ApiError(500, "server blew up"),
    );
    stubConfirm(true);
    const user = userEvent.setup();
    renderCleanButton();

    await user.click(screen.getByRole("button", { name: "Clean" }));

    await waitFor(() => {
      expect(screen.getByRole("alert")).toHaveTextContent("server blew up");
    });
    vi.unstubAllGlobals();
  });
});
