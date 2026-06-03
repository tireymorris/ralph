import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import FollowUpComposer from "./FollowUpComposer";

describe("FollowUpComposer", () => {
  it("send button uses btn and btn--primary classes", () => {
    render(<FollowUpComposer onSubmit={vi.fn()} />);
    const button = screen.getByRole("button", { name: /send/i });
    expect(button).toHaveClass("btn", "btn--primary");
  });
});
