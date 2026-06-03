import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";
import HomePage from "./HomePage";

afterEach(() => {
  cleanup();
});

function renderHomePage() {
  return render(
    <MemoryRouter>
      <HomePage />
    </MemoryRouter>,
  );
}

describe("HomePage", () => {
  it("renders a welcome hero section", () => {
    renderHomePage();
    expect(document.querySelector(".home-hero")).toBeInTheDocument();
  });

  it("renders a primary CTA link to start a new run", () => {
    renderHomePage();
    const cta = screen.getByRole("link", { name: "Start a new run" });
    expect(cta).toHaveAttribute("href", "/new");
    expect(cta).toHaveClass("btn", "btn--primary");
  });
});
