import { execSync, spawn, type ChildProcess } from "node:child_process";
import { mkdtempSync, existsSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test as base, type Page } from "@playwright/test";

const ROOT = join(import.meta.dirname, "..", "..");
const BIN_PATH = join(ROOT, ".ralph-e2e-bin");

export function ralphBinaryPath() {
  return BIN_PATH;
}

let built = false;

export function buildBinary() {
  if (built && existsSync(BIN_PATH)) return;
  execSync("./scripts/build.sh -o .ralph-e2e-bin", { cwd: ROOT, stdio: "pipe" });
  built = true;
}

export function buildFrontend(env: Record<string, string> = {}) {
  execSync("npm run build", {
    cwd: join(ROOT, "web"),
    stdio: "pipe",
    env: { ...process.env, ...env },
  });
}

export interface ServerHandle {
  baseURL: string;
  workDir: string;
  stop: () => void;
}

export async function startServer(
  env: Record<string, string> = {},
): Promise<ServerHandle> {
  const workDir = mkdtempSync(join(tmpdir(), "ralph-e2e-"));
  initGitRepo(workDir);

  return startServerInWorkDir(workDir, env);
}

export function initGitRepo(workDir: string) {
  execSync("git init", { cwd: workDir, stdio: "pipe" });
  execSync('git config user.email "ralph-e2e@example.com"', { cwd: workDir, stdio: "pipe" });
  execSync('git config user.name "Ralph E2E"', { cwd: workDir, stdio: "pipe" });
}

export async function startServerInWorkDir(
  workDir: string,
  env: Record<string, string> = {},
): Promise<ServerHandle> {
  const child: ChildProcess = spawn(BIN_PATH, ["web", "--port", "0"], {
    cwd: workDir,
    env: { ...process.env, RALPH_RUNNER: "mock", ...env },
    stdio: ["ignore", "pipe", "pipe"],
  });

  const baseURL = await waitForListenURL(child);
  await waitForHealth(baseURL);

  return {
    baseURL,
    workDir,
    stop: () => {
      child.kill("SIGTERM");
    },
  };
}

function waitForListenURL(child: ChildProcess): Promise<string> {
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      reject(new Error("timed out waiting for server listen URL"));
    }, 15_000);

    let buffer = "";
    child.stdout!.on("data", (chunk: Buffer) => {
      buffer += chunk.toString();
      const prefix = "ralph web listening on ";
      const idx = buffer.indexOf(prefix);
      if (idx !== -1) {
        const rest = buffer.slice(idx + prefix.length);
        const newline = rest.indexOf("\n");
        if (newline !== -1) {
          clearTimeout(timeout);
          resolve(rest.slice(0, newline).trim());
        }
      }
    });

    child.on("exit", (code) => {
      clearTimeout(timeout);
      reject(new Error(`server exited with code ${code} before printing URL`));
    });
  });
}

async function waitForHealth(baseURL: string, timeoutMs = 10_000) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    try {
      const res = await fetch(`${baseURL}/health`);
      if (res.ok) return;
    } catch {
      // not ready yet
    }
    await new Promise((r) => setTimeout(r, 100));
  }
  throw new Error(`/health did not return 200 within ${timeoutMs}ms`);
}

interface ServerFixtures {
  serverEnv: Record<string, string>;
  server: ServerHandle;
  serverPage: Page;
}

export const test = base.extend<ServerFixtures>({
  serverEnv: [{}, { option: true }],

  server: async ({ serverEnv }, use) => {
    buildBinary();
    const handle = await startServer(serverEnv);
    await use(handle);
    handle.stop();
  },

  serverPage: async ({ server, browser }, use) => {
    const context = await browser.newContext({ baseURL: server.baseURL });
    const page = await context.newPage();
    await use(page);
    await context.close();
  },
});

export { expect } from "@playwright/test";
