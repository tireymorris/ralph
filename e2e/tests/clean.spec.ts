import { test, expect } from "../helpers/server.ts";

test.describe("clean", () => {
  test("clean button shows success after accepting confirm", async ({ serverPage: page }) => {
    page.on("dialog", (dialog) => dialog.accept());

    await page.goto("/new");
    await page.getByRole("button", { name: "Clean" }).click();

    await expect(page.getByText("Ralph state removed")).toBeVisible({
      timeout: 5_000,
    });
  });

  test("conflict clean-and-retry starts a new run with a different id", async ({
    serverPage: page,
  }) => {
    page.on("dialog", (dialog) => dialog.accept());

    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("first goal");
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page).toHaveURL(/\/runs\/[^/]+/);
    const firstRunId = page.url().match(/\/runs\/([^/?#]+)/)?.[1];
    expect(firstRunId).toBeTruthy();

    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("second goal");
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page).toHaveURL(/\/runs\/[^/]+/);
    const secondRunId = page.url().match(/\/runs\/([^/?#]+)/)?.[1];
    expect(secondRunId).toBeTruthy();
    expect(secondRunId).not.toBe(firstRunId);
  });
});
