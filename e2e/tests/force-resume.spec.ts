import { test, expect } from "../helpers/server.ts";

test.describe("force resume", () => {
  test("POST /api/runs/{id}/resume accepts an active run", async ({
    serverPage: page,
  }) => {
    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("stall test goal");
    await page.getByRole("button", { name: /start run/i }).click();
    await expect(page).toHaveURL(/\/runs\/[^/]+/);

    const runId = page.url().match(/\/runs\/([^/?#]+)/)?.[1];
    expect(runId).toBeTruthy();

    const res = await page.request.post(`/api/runs/${runId}/resume`, {
      data: {},
    });
    expect(res.status()).toBe(202);
  });
});
