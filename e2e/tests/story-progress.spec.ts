import { test, expect } from "../helpers/server.ts";

test.use({
  serverEnv: {
    RALPH_MOCK_IMPL_DELAY_MS: "5000",
  },
});

test.describe("story progress", () => {
  test("shows the stories panel during implementation", async ({ serverPage: page }) => {
    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("story progress goal");
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page.locator(".app-main .run-status-badge")).toHaveText("Needs Review", {
      timeout: 30_000,
    });
    await page.getByRole("button", { name: /approve/i }).click();

    await expect(page.locator(".story-progress-panel")).toBeVisible({
      timeout: 30_000,
    });
    await expect(page.getByText("Stories")).toBeVisible();
    await expect(page.locator(".run-detail-progress-label")).toBeVisible();
  });
});
