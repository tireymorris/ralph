import { test, expect } from "../helpers/server.ts";

test.use({
  serverEnv: {
    RALPH_MOCK_IMPL_DELAY_MS: "30000",
  },
});

test.describe("retry cancelled run", () => {
  test("retry starts a new run with the same prompt", async ({ serverPage: page }) => {
    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("retry me");
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page).toHaveURL(/\/runs\/([^/?#]+)/);
    const firstRunId = page.url().match(/\/runs\/([^/?#]+)/)?.[1];
    expect(firstRunId).toBeTruthy();

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

    const retryBtn = page.getByRole("button", { name: /^retry$/i });
    await expect(retryBtn).toBeEnabled();

    const createRunResponse = page.waitForResponse(
      (res) =>
        res.url().includes("/api/runs") &&
        res.request().method() === "POST" &&
        res.status() === 201,
    );
    await retryBtn.click();
    const res = await createRunResponse;
    const { id: secondRunId } = (await res.json()) as { id: string };

    expect(secondRunId).toBeTruthy();
    expect(secondRunId).not.toBe(firstRunId);
    await expect(page).toHaveURL(new RegExp(`/runs/${secondRunId}$`));
    await expect(page.getByText("retry me")).toBeVisible();
  });
});
