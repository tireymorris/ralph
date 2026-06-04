import { describe, expect, it } from "vitest";
import { formatStatus, statusBadgeClass } from "./format";

describe("formatStatus", () => {
  it("maps implementing to Implementing", () => {
    expect(formatStatus("implementing")).toBe("Implementing");
  });

  it("preserves existing labels", () => {
    expect(formatStatus("running")).toBe("Running");
    expect(formatStatus("completed")).toBe("Completed");
    expect(formatStatus("waiting_implementation_review")).toBe("Review Findings");
  });
});

describe("statusBadgeClass", () => {
  it("maps implementing to running badge", () => {
    expect(statusBadgeClass("implementing")).toBe("run-status-badge--running");
  });
});
