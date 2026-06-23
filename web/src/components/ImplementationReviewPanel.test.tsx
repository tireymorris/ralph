import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { continueImplementationReview } from "../api/client";
import ImplementationReviewPanel from "./ImplementationReviewPanel";

vi.mock("../api/client", () => ({
  continueImplementationReview: vi.fn().mockResolvedValue(undefined),
}));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe("ImplementationReviewPanel", () => {
  it("renders cleanup-oriented heading and continue label", () => {
    render(<ImplementationReviewPanel runId="run-1" />);

    expect(
      screen.getByRole("heading", { name: /^cleanup$/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /^continue$/i }),
    ).toBeInTheDocument();
  });

  it("calls continueImplementationReview when Continue is clicked", async () => {
    const onContinued = vi.fn();
    render(
      <ImplementationReviewPanel
        runId="run-1"
        onContinued={onContinued}
      />,
    );

    await userEvent.click(
      screen.getByRole("button", { name: /^continue$/i }),
    );

    await waitFor(() => {
      expect(continueImplementationReview).toHaveBeenCalledWith("run-1");
      expect(onContinued).toHaveBeenCalledOnce();
    });
  });

  it("shows an error alert when continue fails", async () => {
    vi.mocked(continueImplementationReview).mockRejectedValue(
      new Error("server error"),
    );

    render(<ImplementationReviewPanel runId="run-1" />);

    await userEvent.click(
      screen.getByRole("button", { name: /^continue$/i }),
    );

    await waitFor(() => {
      expect(screen.getByRole("alert")).toHaveTextContent("server error");
    });
    expect(
      screen.getByRole("button", { name: /^continue$/i }),
    ).toBeEnabled();
  });

  it("disables the button and shows Continuing while submit is in flight", async () => {
    let resolveContinue!: () => void;
    vi.mocked(continueImplementationReview).mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          resolveContinue = resolve;
        }),
    );

    render(<ImplementationReviewPanel runId="run-1" />);

    await userEvent.click(
      screen.getByRole("button", { name: /^continue$/i }),
    );

    expect(
      screen.getByRole("button", { name: /continuing/i }),
    ).toBeDisabled();

    resolveContinue();
    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /^continue$/i }),
      ).toBeEnabled();
    });
  });
});
