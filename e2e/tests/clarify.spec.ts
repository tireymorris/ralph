import { test, expect } from "../helpers/server.ts";

test.use({
  serverEnv: {
    RALPH_MOCK_QUESTIONS: '["What is the scope?","Preferred language?"]',
  },
});

test.describe("clarify flow", () => {
  test("questions appear and submitting answers proceeds to review", async ({ serverPage: page }) => {
    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("test clarify");
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page.locator(".app-main .run-status-badge")).toHaveText("Needs Answers", {
      timeout: 30_000,
    });

    const textareas = page.locator(".clarify-form textarea");
    await expect(textareas).toHaveCount(2, { timeout: 10_000 });
    await textareas.nth(0).fill("Full application");
    await textareas.nth(1).fill("TypeScript");

    await page.getByRole("button", { name: /submit answers/i }).click();

    await expect(page.locator(".app-main .run-status-badge")).toHaveText("Needs Review", {
      timeout: 30_000,
    });
  });
});
