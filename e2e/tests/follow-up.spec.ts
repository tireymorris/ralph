import { test, expect } from "../helpers/server.ts";

test.describe("follow-up", () => {
  test("sending follow-up on completed run shows acceptance in timeline", async ({ serverPage: page }) => {
    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("test follow-up");
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page.locator(".app-main .run-status-badge")).toHaveText("Needs Review", {
      timeout: 30_000,
    });
    await page.getByRole("button", { name: /approve/i }).click();
    await expect(page.locator(".app-main .run-status-badge")).toHaveText("Completed", {
      timeout: 30_000,
    });

    const composer = page.locator(".follow-up-composer");
    await expect(composer).toBeVisible();
    await composer.getByRole("textbox").fill("add unit tests");
    await composer.getByRole("button", { name: /send/i }).click();

    await expect(
      page.locator(".timeline-entry", { hasText: "Follow-up accepted" }),
    ).toBeVisible({ timeout: 15_000 });
  });
});
