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

    expect(screen.getByText("shows as done")).toBeInTheDocument();
    expect(screen.getByText("extract helper")).toBeInTheDocument();
    expect(screen.getByText("1/1 slices done")).toBeInTheDocument();
  });
});
