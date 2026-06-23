import { test, expect } from "../helpers/server.ts";

test.use({
  serverEnv: {
    RALPH_MOCK_IMPL_DELAY_MS: "30000",
  },
});

test.describe("force resume UI", () => {
  test("shows force resume after stall and submits resume", async ({ serverPage: page }) => {
    await page.clock.install();
    page.on("dialog", (dialog) => dialog.accept());

    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("force resume ui goal");
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page).toHaveURL(/\/runs\/[^/]+/);
    await expect(page.locator(".app-main .run-status-badge")).toHaveText("Needs Review", {
      timeout: 30_000,
    });
    await page.getByRole("button", { name: /approve/i }).click();

    await expect(page.locator(".app-main .run-status-badge")).toHaveText("Implementing", {
      timeout: 30_000,
    });

    await page.clock.fastForward(6_000);

    const forceResumeBtn = page.getByRole("button", { name: /force resume/i });
    await expect(forceResumeBtn).toBeVisible();
    await forceResumeBtn.click();

    await expect(page.locator(".timeline-entry").filter({ hasText: /Force resume requested/i })).toHaveCount(
      1,
    );
    await expect(page.locator(".app-main .run-status-badge")).toHaveText("Implementing");
  });
});
