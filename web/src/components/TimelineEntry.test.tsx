import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import TimelineEntryBubble from "./TimelineEntry";

afterEach(() => {
  cleanup();
});

describe("TimelineEntryBubble", () => {
  it("labels assistant entries for assistive tech", () => {
    render(<TimelineEntryBubble variant="assistant" text="Hello" />);
    expect(screen.getByLabelText("Assistant message")).toBeInTheDocument();
  });

  it("labels system entries for assistive tech", () => {
    render(<TimelineEntryBubble variant="system" text="Phase started" />);
    expect(screen.getByLabelText("System message")).toBeInTheDocument();
  });

  it("labels error entries for assistive tech", () => {
    render(<TimelineEntryBubble variant="error" text="Something failed" />);
    expect(screen.getByLabelText("Error message")).toBeInTheDocument();
  });
});
