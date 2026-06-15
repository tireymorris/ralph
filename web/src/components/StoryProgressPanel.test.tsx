import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import StoryProgressPanel from "./StoryProgressPanel";

const expandedPRD = {
  version: 1,
  project_name: "Test",
  context: "Build a dashboard",
  test_spec: "Render the story list",
  branch_name: "story-types",
  test_command: "npm test",
  stories: [
    {
      id: "story-1",
      title: "Completed story",
      description: "Done already",
      slices: [
        {
          id: "slice-1",
          behavior: "shows as done",
          red_hint: "make it fail",
          passes: true,
        },
      ],
      priority: 1,
      depends_on: ["setup"],
      passes: true,
    },
    {
      id: "story-2",
      title: "Pending story",
      description: "Still pending",
      slices: [
        {
          id: "slice-1",
          behavior: "shows as pending",
          red_hint: "make it fail",
          refactor_hint: "extract helper",
          passes: false,
        },
      ],
      priority: 2,
      passes: false,
    },
  ],
};

const mixedSlicePRD = {
  version: 1,
  project_name: "Test",
  stories: [
    {
      id: "story-1",
      title: "Slice labels story",
      description: "Show slice states",
      slices: [
        {
          id: "slice-1",
          behavior: "passed slice behavior",
          red_hint: "passed slice red hint",
          passes: true,
        },
        {
          id: "slice-2",
          behavior: "current slice behavior",
          red_hint: "current slice red hint",
          passes: false,
        },
        {
          id: "slice-3",
          behavior: "pending slice behavior",
          red_hint: "pending slice red hint",
          passes: false,
        },
      ],
      priority: 1,
      passes: false,
    },
  ],
};

afterEach(() => {
  cleanup();
});

describe("StoryProgressPanel", () => {
  it("renders story title and pass state from the expanded PRD story shape", () => {
    render(<StoryProgressPanel prd={expandedPRD} />);

    expect(screen.getByText("Completed story")).toBeInTheDocument();
    expect(screen.getByText("Pending story")).toBeInTheDocument();
    expect(screen.getByText("1/2 done")).toBeInTheDocument();
  });

  it("renders slice progress and refactor hints", () => {
    render(<StoryProgressPanel prd={expandedPRD} />);

    expect(screen.getAllByText("Behavior:")).toHaveLength(2);
    expect(screen.getAllByText("Red hint:")).toHaveLength(2);
    expect(screen.getAllByText("Refactor hint:")).toHaveLength(1);
    expect(screen.getByText("shows as done")).toBeInTheDocument();
    expect(screen.getByText("extract helper")).toBeInTheDocument();
    expect(screen.getByText("1/1 slices done")).toBeInTheDocument();
  });

  it("renders labels for completed, in progress, and pending slices", () => {
    render(<StoryProgressPanel prd={mixedSlicePRD} />);

    expect(screen.getByText("completed")).toBeInTheDocument();
    expect(screen.getByText("in progress")).toBeInTheDocument();
    expect(screen.getByText("pending")).toBeInTheDocument();
    expect(screen.getByText("passed slice behavior")).toBeInTheDocument();
    expect(screen.getByText("current slice red hint")).toBeInTheDocument();
  });
});
