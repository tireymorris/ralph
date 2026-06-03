import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import ClarifyForm from "./ClarifyForm";

afterEach(() => {
  cleanup();
});

describe("ClarifyForm", () => {
  it("disables Submit when no answers are filled", () => {
    const onSubmit = vi.fn();
    render(
      <ClarifyForm
        questions={["What is the goal?", "What stack?"]}
        onSubmit={onSubmit}
      />,
    );

    expect(screen.getByRole("button", { name: /submit/i })).toBeDisabled();
  });

  it("submit button uses btn and btn--primary classes", () => {
    render(<ClarifyForm questions={["What is the goal?"]} onSubmit={vi.fn()} />);
    expect(screen.getByRole("button", { name: /submit/i })).toHaveClass(
      "btn",
      "btn--primary",
    );
  });

  it("enables Submit when all questions have answers", async () => {
    const onSubmit = vi.fn();
    render(
      <ClarifyForm
        questions={["What is the goal?"]}
        onSubmit={onSubmit}
      />,
    );

    await userEvent.type(
      screen.getByLabelText("What is the goal?"),
      "build api",
    );
    expect(screen.getByRole("button", { name: /submit/i })).toBeEnabled();
  });
});
