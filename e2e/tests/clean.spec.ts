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
});
