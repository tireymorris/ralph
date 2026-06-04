import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { postClean } from "../api/client";
import CleanButton from "./CleanButton";

vi.mock("../api/client", () => ({
  postClean: vi.fn(),
}));

describe("CleanButton", () => {
  beforeEach(() => {
    vi.mocked(postClean).mockReset();
  });

  it("renders a Clean button in the header", () => {
    render(<CleanButton />);
    expect(screen.getByRole("button", { name: "Clean" })).toBeInTheDocument();
  });

  it("asks to confirm removal of prd.json and .ralph/ before cleaning", async () => {
    const confirm = vi.fn(() => false);
    vi.stubGlobal("confirm", confirm);
    const user = userEvent.setup();
    render(<CleanButton />);

    await user.click(screen.getByRole("button", { name: "Clean" }));

    expect(confirm).toHaveBeenCalledWith(
      expect.stringMatching(/prd\.json/i),
    );
    expect(confirm.mock.calls[0]?.[0]).toMatch(/\.ralph\//);
    vi.unstubAllGlobals();
  });

  it("does not call postClean when confirm is declined", async () => {
    vi.stubGlobal("confirm", vi.fn(() => false));
    const user = userEvent.setup();
    render(<CleanButton />);

    await user.click(screen.getByRole("button", { name: "Clean" }));

    expect(postClean).not.toHaveBeenCalled();
    vi.unstubAllGlobals();
  });
});
