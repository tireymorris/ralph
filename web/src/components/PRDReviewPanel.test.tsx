import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { submitReview } from "../api/client";
import PRDReviewPanel from "./PRDReviewPanel";

vi.mock("../api/client", () => ({
  submitReview: vi.fn().mockResolvedValue(undefined),
}));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

const samplePRD = {
  version: 1,
  project_name: "Test",
  stories: [
    {
      id: "story-1",
      title: "First story",
      description: "Do thing",
      acceptance_criteria: ["works"],
      priority: 1,
      passes: false,
    },
  ],
};

describe("PRDReviewPanel", () => {
  it("calls submitReview with approve once when Approve is clicked", async () => {
    const onApproved = vi.fn();
    render(
      <PRDReviewPanel runId="run-1" prd={samplePRD} onApproved={onApproved} />,
    );

    await userEvent.click(screen.getByRole("button", { name: /approve/i }));

    await waitFor(() => {
      expect(submitReview).toHaveBeenCalledTimes(1);
    });
    expect(submitReview).toHaveBeenCalledWith("run-1", "approve");
  });

  it("approve button uses btn and btn--primary classes", () => {
    render(<PRDReviewPanel runId="run-1" prd={samplePRD} />);
    expect(screen.getByRole("button", { name: /approve/i })).toHaveClass(
      "btn",
      "btn--primary",
    );
  });

  it("revise button uses btn and btn--secondary classes", () => {
    render(<PRDReviewPanel runId="run-1" prd={samplePRD} />);
    expect(
      screen.getByRole("button", { name: /send revision/i }),
    ).toHaveClass("btn", "btn--secondary");
  });

  it("critique textarea uses composer-input class", () => {
    render(<PRDReviewPanel runId="run-1" prd={samplePRD} />);
    expect(
      screen.getByLabelText(/request changes/i, { selector: "textarea" }),
    ).toHaveClass("composer-input");
  });
});
