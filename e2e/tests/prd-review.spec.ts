import { test, expect } from "../helpers/server.ts";

async function createRunAndWaitForReview(page: import("@playwright/test").Page) {
  await page.goto("/new");
  await page.getByRole("textbox", { name: "Goal prompt" }).fill("test prd review");
  await page.getByRole("button", { name: /start run/i }).click();
  await expect(page.locator(".app-main .run-status-badge")).toHaveText("Needs Review", {
    timeout: 30_000,
  });
}

test.describe("PRD review", () => {
  test("review panel shows project name, version, and stories", async ({ serverPage: page }) => {
    await createRunAndWaitForReview(page);

    await expect(page.locator(".prd-review-panel")).toBeVisible();
    await expect(page.locator(".prd-review-panel h2")).toHaveText("Mock");
    await expect(page.locator(".prd-review-version")).toContainText("v1");
    await expect(page.locator(".prd-review-stories > li")).toHaveCount(1);
    await expect(page.locator(".prd-review-stories > li").first()).toContainText("Mock story");
    await expect(page.locator(".prd-review-slices li")).toHaveCount(1);
    await expect(page.locator(".prd-review-slices li").first()).toContainText("ok");
  });

  test("approving PRD completes the run", async ({ serverPage: page }) => {
    await createRunAndWaitForReview(page);

    await page.getByRole("button", { name: /approve/i }).click();

    await expect(page.locator(".app-main .run-status-badge")).toHaveText("Completed", {
      timeout: 30_000,
    });
  });

  test("revising PRD re-shows the review panel", async ({ serverPage: page }) => {
    await createRunAndWaitForReview(page);

    await page.getByPlaceholder("Describe what should change").fill("add more stories");
    await page.getByRole("button", { name: /send revision/i }).click();

    await expect(page.locator(".prd-review-panel")).toBeVisible({ timeout: 30_000 });
    await expect(page.locator(".app-main .run-status-badge")).toHaveText("Needs Review", {
      timeout: 30_000,
    });
  });
});
