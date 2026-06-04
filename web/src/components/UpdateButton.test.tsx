import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getVersion, postUpdate } from "../api/client";
import UpdateButton from "./UpdateButton";

vi.mock("../api/client", () => ({
  getVersion: vi.fn(),
  postUpdate: vi.fn(),
}));

describe("UpdateButton", () => {
  beforeEach(() => {
    vi.mocked(getVersion).mockReset();
    vi.mocked(postUpdate).mockReset();
  });

  it("shows green up-to-date label when current", async () => {
    vi.mocked(getVersion).mockResolvedValue({
      version: "1.0",
      commit: "abc",
      ref: "main",
      status: "current",
    });

    render(<UpdateButton />);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Up to date" })).toBeDisabled();
    });
    expect(screen.getByRole("button")).toHaveClass("btn--update-current");
  });

  it("shows amber update-available label when out of date", async () => {
    vi.mocked(getVersion).mockResolvedValue({
      version: "1.0",
      commit: "abc",
      ref: "main",
      status: "available",
    });

    render(<UpdateButton />);

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: "Update available" }),
      ).toBeEnabled();
    });
    expect(screen.getByRole("button")).toHaveClass("btn--update-available");
  });

  it("runs update when available and confirmed", async () => {
    vi.mocked(getVersion).mockResolvedValue({
      version: "1.0",
      commit: "abc",
      ref: "main",
      status: "available",
    });
    vi.mocked(postUpdate).mockResolvedValue({
      status: "updated",
      message: "updated; restart ralph web",
    });
    vi.stubGlobal("confirm", vi.fn(() => true));

    const user = userEvent.setup();
    render(<UpdateButton />);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Update available" })).toBeEnabled();
    });
    await user.click(screen.getByRole("button", { name: "Update available" }));

    await waitFor(() => {
      expect(postUpdate).toHaveBeenCalled();
    });
    vi.unstubAllGlobals();
  });
});
