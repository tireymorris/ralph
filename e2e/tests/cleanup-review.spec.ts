import { test, expect, buildBinary, initGitRepo, startServerInWorkDir } from "../helpers/server.ts";
import { seedWaitingCleanupReviewRun } from "../helpers/cleanup-review.ts";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const seededRunId = "run-cleanup-review";

const seededCleanupTest = test.extend({
  server: async ({ serverEnv }, use) => {
    buildBinary();
    const workDir = mkdtempSync(join(tmpdir(), "ralph-e2e-"));
    initGitRepo(workDir);
    seedWaitingCleanupReviewRun(workDir, seededRunId);
    const handle = await startServerInWorkDir(workDir, serverEnv);
    await use(handle);
    handle.stop();
  },
});

test.describe("cleanup review", () => {
  seededCleanupTest.describe("seeded waiting run", () => {
    seededCleanupTest("shows cleanup review panel and status badge", async ({ serverPage: page }) => {
      await page.goto(`/runs/${seededRunId}`);

      await expect(page.locator(".app-main .run-status-badge")).toHaveText("Cleanup Review");
      await expect(page.getByRole("region", { name: /cleanup review/i })).toBeVisible();
      await expect(page.getByRole("button", { name: /continue cleanup/i })).toBeVisible();
    });

    seededCleanupTest("continue cleanup completes the run without restarting stories", async ({
      serverPage: page,
    }) => {
      await page.goto(`/runs/${seededRunId}`);

      await expect(page.getByRole("button", { name: /continue cleanup/i })).toBeVisible();

      await page.getByRole("button", { name: /continue cleanup/i }).click();

      await expect(page.locator(".app-main .run-status-badge")).toHaveText("Completed", {
        timeout: 30_000,
      });
      await expect(page.getByRole("region", { name: /cleanup review/i })).toHaveCount(0);
      await expect(page.locator(".timeline-entry").filter({ hasText: /Started story:/i })).toHaveCount(
        0,
      );
    });
  });

  test("auto-approve run completes through cleanup without manual gates", async ({
    serverPage: page,
  }) => {
    await page.goto("/new");
    await page.getByRole("textbox", { name: "Goal prompt" }).fill("auto cleanup path");
    await page.getByRole("checkbox", { name: /auto-approve/i }).check();
    await page.getByRole("button", { name: /start run/i }).click();

    await expect(page).toHaveURL(/\/runs\/[^/]+/);
    await expect(page.locator(".clarify-form")).toHaveCount(0);
    await expect(page.getByRole("button", { name: /approve/i })).toHaveCount(0);
    await expect(page.getByRole("region", { name: /cleanup review/i })).toHaveCount(0);

    const statusBadge = page.locator(".run-detail-toolbar .run-status-badge").first();
    await expect(statusBadge).toHaveText("Completed", {
      timeout: 60_000,
    });
  });
});
