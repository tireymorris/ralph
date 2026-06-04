import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import RunPrompt from "./RunPrompt";

describe("RunPrompt", () => {
  it("shows toggle for long prompts and expands on click", async () => {
    const prompt = "a".repeat(200);
    render(<RunPrompt prompt={prompt} />);

    expect(screen.getByText(prompt)).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /show full prompt/i }),
    ).toBeInTheDocument();

    await userEvent.click(
      screen.getByRole("button", { name: /show full prompt/i }),
    );
    expect(
      screen.getByRole("button", { name: /show less/i }),
    ).toBeInTheDocument();
  });
});
