import { test, expect } from "../helpers/server.ts";

test.use({
  serverEnv: {
    RALPH_MOCK_IMPL_DELAY_MS: "30000",
  },
});

test.describe("active runs list", () => {
  test("shows an in-progress run on home and navigates back to it", async ({ serverPage: page }) => {
    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("active run list goal");
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page).toHaveURL(/\/runs\/([^/?#]+)/);
    const runId = page.url().match(/\/runs\/([^/?#]+)/)?.[1];
    expect(runId).toBeTruthy();

    await page.getByRole("link", { name: "New" }).click();
    await expect(page).toHaveURL(/\/$/);

    const activeSection = page.getByRole("heading", { name: "Active chats" });
    await expect(activeSection).toBeVisible({ timeout: 15_000 });

    const activeRunLink = page.getByRole("link", { name: /active run list goal/i });
    await expect(activeRunLink).toBeVisible();
    await activeRunLink.click();

    await expect(page).toHaveURL(new RegExp(`/runs/${runId}$`));
    await expect(page.getByText("active run list goal")).toBeVisible();
  });
});
