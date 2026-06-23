import {
  test,
  expect,
  buildBinary,
  initGitRepo,
  startServerInWorkDir,
} from "../helpers/server.ts";
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
    seededCleanupTest("shows cleanup panel, status badge, and timeline", async ({
      serverPage: page,
    }) => {
      await page.goto(`/runs/${seededRunId}`);

      await expect(page.locator(".app-main .run-status-badge")).toHaveText("Cleanup");
      await expect(page.getByRole("region", { name: /^cleanup$/i })).toBeVisible();
      await expect(page.getByRole("button", { name: /^continue$/i })).toBeVisible();
      await expect(page.locator(".timeline-entry").filter({ hasText: /Cleanup started/i })).toHaveCount(
        1,
      );
      await expect(page.locator(".timeline-entry").filter({ hasText: /Cleanup findings:/i })).toHaveCount(
        1,
      );
    });

    seededCleanupTest("shows review iteration in panel copy", async ({ browser }) => {
      buildBinary();
      const workDir = mkdtempSync(join(tmpdir(), "ralph-e2e-"));
      initGitRepo(workDir);
      const runId = seedWaitingCleanupReviewRun(workDir, {
        runId: "run-cleanup-iter-2",
        reviewIteration: 2,
      });
      const handle = await startServerInWorkDir(workDir, {});
      const context = await browser.newContext({ baseURL: handle.baseURL });
      const page = await context.newPage();
      try {
        await page.goto(`/runs/${runId}`);
        await expect(page.getByRole("region", { name: /^cleanup$/i })).toContainText(
          "(iteration 2)",
        );
      } finally {
        await context.close();
        handle.stop();
      }
    });

    seededCleanupTest("continue cleanup completes the run without restarting stories", async ({
      serverPage: page,
    }) => {
      await page.goto(`/runs/${seededRunId}`);

      await expect(page.getByRole("button", { name: /^continue$/i })).toBeVisible();

      await page.getByRole("button", { name: /^continue$/i }).click();

      await expect(page.locator(".app-main .run-status-badge")).toHaveText("Completed", {
        timeout: 30_000,
      });
      await expect(page.getByRole("region", { name: /^cleanup$/i })).toHaveCount(0);
      await expect(page.locator(".timeline-entry").filter({ hasText: /Started story:/i })).toHaveCount(
        0,
      );
      await expect(page.locator(".timeline-entry").filter({ hasText: /Cleanup finished/i })).toHaveCount(
        1,
      );
    });

    seededCleanupTest("cancel while waiting updates status and hides panel", async ({
      serverPage: page,
    }) => {
      await page.goto(`/runs/${seededRunId}`);

      await expect(page.getByRole("button", { name: /^continue$/i })).toBeVisible();
      await page.getByRole("button", { name: "Cancel" }).click();

      await expect(page.locator(".app-main .run-status-badge")).toHaveText("Cancelled", {
        timeout: 15_000,
      });
      await expect(page.getByRole("region", { name: /^cleanup$/i })).toHaveCount(0);
    });

    seededCleanupTest("POST /resume from impl_review completes without restarting stories", async ({
      serverPage: page,
    }) => {
      await page.goto(`/runs/${seededRunId}`);

      const res = await page.request.post(`/api/runs/${seededRunId}/resume`, {
        data: {},
      });
      expect(res.status()).toBe(202);

      await expect(page.locator(".app-main .run-status-badge")).toHaveText("Completed", {
        timeout: 30_000,
      });
      await expect(page.locator(".timeline-entry").filter({ hasText: /Started story:/i })).toHaveCount(
        0,
      );
    });

    seededCleanupTest("POST /implementation-review returns 409 when not waiting", async ({
      serverPage: page,
    }) => {
      await page.goto(`/runs/${seededRunId}`);

      await page.getByRole("button", { name: /^continue$/i }).click();
      await expect(page.locator(".app-main .run-status-badge")).toHaveText("Completed", {
        timeout: 30_000,
      });

      const res = await page.request.post(
        `/api/runs/${seededRunId}/implementation-review`,
        { data: {} },
      );
      expect(res.status()).toBe(409);
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
    await expect(page.getByRole("region", { name: /^cleanup$/i })).toHaveCount(0);

    const statusBadge = page.locator(".run-detail-toolbar .run-status-badge").first();
    await expect(statusBadge).toHaveText("Completed", {
      timeout: 60_000,
    });
  });
});
