import { test, expect } from "../helpers/server.ts";

test.describe("navigation", () => {
  test("home page renders branding and CTA", async ({ serverPage: page }) => {
    await page.goto("/");
    await expect(page.locator(".app-brand")).toHaveText("Ralph");
    await expect(page.locator(".home-hero h1")).toHaveText("Ralph");
    await expect(page.getByRole("link", { name: "Start a new run" })).toBeVisible();
  });

  test("clicking CTA navigates to /new", async ({ serverPage: page }) => {
    await page.goto("/");
    await page.getByRole("link", { name: "Start a new run" }).click();
    await expect(page).toHaveURL(/\/new$/);
    await expect(page.getByRole("textbox", { name: "Goal prompt" })).toBeVisible();
  });

  test("sidebar has Home and New run nav links", async ({ serverPage: page }) => {
    await page.goto("/");
    const nav = page.getByRole("navigation", { name: "Main" });
    await expect(nav.getByRole("link", { name: "Home" })).toBeVisible();
    await expect(nav.getByRole("link", { name: "New run" })).toBeVisible();
  });

  test("new run page has goal textarea and submit button", async ({ serverPage: page }) => {
    await page.goto("/new");
    await expect(page.getByRole("textbox", { name: "Goal prompt" })).toBeVisible();
    await expect(page.getByRole("button", { name: /start run/i })).toBeVisible();
  });
});
