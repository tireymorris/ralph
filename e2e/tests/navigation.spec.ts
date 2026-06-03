import { test, expect } from "../helpers/server.ts";

test.describe("navigation", () => {
  test("home page shows goal composer", async ({ serverPage: page }) => {
    await page.goto("/");
    await expect(page.getByRole("textbox", { name: "Goal prompt" })).toBeVisible();
    await expect(page.getByRole("button", { name: /start run/i })).toBeVisible();
  });

  test("/new redirects to home", async ({ serverPage: page }) => {
    await page.goto("/new");
    await expect(page).toHaveURL(/\/$/);
    await expect(page.getByRole("textbox", { name: "Goal prompt" })).toBeVisible();
  });
});
