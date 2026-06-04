import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { postClean } from "../api/client";
import CleanButton from "./CleanButton";

vi.mock("../api/client", () => ({
  postClean: vi.fn(),
}));

describe("CleanButton", () => {
  it("renders a Clean button in the header", () => {
    render(<CleanButton />);
    expect(screen.getByRole("button", { name: "Clean" })).toBeInTheDocument();
  });
});
