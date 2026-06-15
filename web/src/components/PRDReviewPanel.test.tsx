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
      slices: [
        {
          id: "slice-1",
          behavior: "works",
          red_hint: "make it fail",
          refactor_hint: "extract helper",
          passes: false,
        },
      ],
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

  it("renders slice details and refactor hints", () => {
    render(<PRDReviewPanel runId="run-1" prd={samplePRD} />);
    expect(screen.getAllByText("Behavior:")).toHaveLength(1);
    expect(screen.getAllByText("Red hint:")).toHaveLength(1);
    expect(screen.getAllByText("Refactor hint:")).toHaveLength(1);
    expect(screen.getByText("works")).toBeInTheDocument();
    expect(screen.getByText("make it fail")).toBeInTheDocument();
    expect(screen.getByText("extract helper")).toBeInTheDocument();
  });
});
