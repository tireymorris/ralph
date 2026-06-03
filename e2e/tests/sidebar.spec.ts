import { test, expect } from "../helpers/server.ts";

test.describe("sidebar", () => {
  test("created run appears in sidebar runs list", async ({ serverPage: page }) => {
    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("sidebar test");
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page).toHaveURL(/\/runs\/.+/);

    const sidebar = page.locator(".app-sidebar");
    await expect(sidebar.locator(".runs-list a")).toHaveCount(1, {
      timeout: 10_000,
    });
    await expect(sidebar.locator(".runs-list a").first()).toContainText("sidebar test");
  });

  test("sidebar entry has status badge", async ({ serverPage: page }) => {
    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("badge test");
    await page.getByRole("button", { name: /start run/i }).click();

    const sidebar = page.locator(".app-sidebar");
    await expect(sidebar.locator(".runs-list .run-status-badge")).toBeVisible({
      timeout: 10_000,
    });
  });

  test("clicking sidebar entry navigates to detail", async ({ serverPage: page }) => {
    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("nav test");
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page).toHaveURL(/\/runs\/.+/);
    const runURL = page.url();

    await page.goto("/");
    const sidebar = page.locator(".app-sidebar");
    await sidebar.locator(".runs-list a").first().click();

    await expect(page).toHaveURL(runURL);
    await expect(page.locator(".run-detail-title")).toHaveText("nav test");
  });

  test("active run is highlighted in sidebar", async ({ serverPage: page }) => {
    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("highlight test");
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page).toHaveURL(/\/runs\/.+/);

    const link = page.locator(".app-sidebar .runs-list a").first();
    await expect(link).toHaveAttribute("aria-current", "page", {
      timeout: 10_000,
    });
  });
});
