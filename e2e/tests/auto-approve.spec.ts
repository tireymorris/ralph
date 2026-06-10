import { spawn, type ChildProcess } from "node:child_process";
import { existsSync, mkdtempSync, readFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test, expect, buildBinary, initGitRepo, ralphBinaryPath, startServerInWorkDir } from "../helpers/server.ts";

interface StoryStatus {
  id: string;
  passes?: boolean;
}

interface PRDStatus {
  stories?: StoryStatus[];
}

test.describe("auto-approve CLI run", () => {
  test("--yolo completes without rendering manual gates", async ({ page }) => {
    buildBinary();
    const workDir = mkdtempSync(join(tmpdir(), "ralph-yolo-e2e-"));
    initGitRepo(workDir);

    const cli = spawnRalphYolo(workDir);
    const server = await startServerInWorkDir(workDir, { RALPH_RUNNER: "mock" });

    try {
      await waitForPRD(workDir);
      await page.goto(`${server.baseURL}/runs/prd-local`);

      await expect(page.locator(".clarify-form")).toHaveCount(0);
      await expect(page.getByRole("button", { name: /approve/i })).toHaveCount(0);

      await waitForAllStoriesPassing(workDir, async () => {
        await page.reload();
        await expect(page.locator(".clarify-form")).toHaveCount(0);
        await expect(page.getByRole("button", { name: /approve/i })).toHaveCount(0);
      });

      const prd = readPRD(workDir);
      expect(prd.stories?.length).toBeGreaterThan(0);
      expect(prd.stories?.every((story) => story.passes === true)).toBe(true);
    } finally {
      server.stop();
      cli.kill("SIGTERM");
    }
  });
});

function spawnRalphYolo(workDir: string): ChildProcess {
  const bin = ralphBinaryPath();
  const args = process.platform === "darwin"
    ? ["-q", "/dev/null", bin, "--yolo", "build a widget"]
    : ["-q", "-c", `${shellQuote(bin)} --yolo ${shellQuote("build a widget")}`, "/dev/null"];

  return spawn("script", args, {
    cwd: workDir,
    env: {
      ...process.env,
      RALPH_RUNNER: "mock",
      RALPH_MOCK_IMPL_DELAY_MS: "2000",
    },
    stdio: "ignore",
  });
}

function shellQuote(value: string): string {
  return `'${value.replaceAll("'", `'\\''`)}'`;
}

async function execInWorkDir(command: string, args: string[], cwd: string): Promise<void> {
  await new Promise<void>((resolve, reject) => {
    const child = spawn(command, args, { cwd, stdio: "ignore" });
    child.on("error", reject);
    child.on("exit", (code) => {
      if (code === 0) resolve();
      else reject(new Error(`${command} exited with code ${code}`));
    });
  });
}

async function waitForPRD(workDir: string): Promise<void> {
  await waitUntil(() => existsSync(join(workDir, "prd.json")), "prd.json to exist");
}

async function waitForAllStoriesPassing(workDir: string, tick: () => Promise<void>): Promise<void> {
  await waitUntil(async () => {
    await tick();
    const prd = readPRD(workDir);
    return Boolean(prd.stories?.length) && prd.stories.every((story) => story.passes === true);
  }, "all stories to pass");
}

function readPRD(workDir: string): PRDStatus {
  return JSON.parse(readFileSync(join(workDir, "prd.json"), "utf8")) as PRDStatus;
}

async function waitUntil(check: () => boolean | Promise<boolean>, label: string): Promise<void> {
  const deadline = Date.now() + 90_000;
  let lastError: unknown;
  while (Date.now() < deadline) {
    try {
      if (await check()) return;
    } catch (error) {
      lastError = error;
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`timed out waiting for ${label}${lastError ? `: ${String(lastError)}` : ""}`);
}
