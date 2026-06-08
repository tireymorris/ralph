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
      acceptance_criteria: ["shows as done"],
      priority: 1,
      depends_on: ["setup"],
      passes: true,
    },
    {
      id: "story-2",
      title: "Pending story",
      description: "Still pending",
      acceptance_criteria: ["shows as pending"],
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
});
