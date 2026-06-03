import { test, expect } from "../helpers/server.ts";

test.describe("create run", () => {
  test("submitting a goal creates a run and shows detail page", async ({ serverPage: page }) => {
    await page.goto("/new");

    await page.getByRole("textbox", { name: "Goal prompt" }).fill("build a widget");
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page).toHaveURL(/\/runs\/.+/);
    await expect(page.locator(".run-detail-prompt")).toHaveText("build a widget");
  });

  test("status badge appears after run is created", async ({ serverPage: page }) => {
    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("test status");
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page.locator(".app-main .run-status-badge")).toBeVisible();
  });

  test("timeline populates with at least one entry via SSE", async ({ serverPage: page }) => {
    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("test timeline");
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page.locator(".timeline-entry").first()).toBeVisible({
      timeout: 30_000,
    });
  });

  test("run reaches waiting_review status", async ({ serverPage: page }) => {
    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("test review gate");
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page.locator(".app-main .run-status-badge")).toHaveText("Needs Review", {
      timeout: 30_000,
    });
  });
});
