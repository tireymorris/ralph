import { test, expect } from "../helpers/server.ts";

test.use({
  serverEnv: {
    RALPH_MOCK_IMPL_DELAY_MS: "30000",
  },
});

test.describe("cancel run", () => {
  test("cancelling during implementation updates status", async ({ serverPage: page }) => {
    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("test cancel");
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page.locator(".app-main .run-status-badge")).toHaveText("Needs Review", {
      timeout: 30_000,
    });
    await page.getByRole("button", { name: /approve/i }).click();

    const cancelBtn = page.getByRole("button", { name: "Cancel" });
    await expect(cancelBtn).toBeVisible({ timeout: 15_000 });
    await cancelBtn.click();

    await expect(page.locator(".app-main .run-status-badge")).toHaveText("Cancelled", {
      timeout: 15_000,
    });
    await expect(cancelBtn).toBeHidden();
  });
});
