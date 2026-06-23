import { execSync } from "node:child_process";
import { mkdirSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const findingsTranscript = `===ralph-findings===
[{"category":"bug","path":"feature.go","summary":"add test"}]
===/ralph-findings===`;

export function seedWaitingCleanupReviewRun(workDir: string, runId = "run-cleanup-review") {
  const prd = {
    version: 1,
    project_name: "Mock",
    branch_name: "main",
    stories: [
      {
        id: "story-1",
        title: "Mock story",
        description: "d",
        slices: [
          {
            id: "slice-1",
            behavior: "first behavior",
            red_hint: "write failing test for first behavior",
            passes: true,
          },
          {
            id: "slice-2",
            behavior: "second behavior",
            red_hint: "write failing test for second behavior",
            passes: true,
          },
        ],
        priority: 1,
        passes: true,
      },
    ],
  };

  writeFileSync(join(workDir, "prd.json"), JSON.stringify(prd));
  writeFileSync(join(workDir, "main.go"), "package main\n");
  execSync('git add main.go prd.json && git commit -m "seed completed prd"', {
    cwd: workDir,
    stdio: "pipe",
  });
  writeFileSync(join(workDir, "feature.go"), "package main\n// feature\n");

  const runDir = join(workDir, ".ralph", "runs", runId);
  mkdirSync(runDir, { recursive: true });
  writeFileSync(join(runDir, "review-1.txt"), findingsTranscript);

  const now = new Date().toISOString();
  const meta = {
    id: runId,
    work_dir: workDir,
    prompt: "cleanup review e2e",
    status: "waiting_implementation_review",
    phase: "cleanup",
    checkpoint: "impl_review",
    review_iteration: 1,
    last_review_transcript_path: "review-1.txt",
    created_at: now,
    updated_at: now,
    prd_path: "prd.json",
  };
  writeFileSync(join(runDir, "meta.json"), JSON.stringify(meta, null, 2));

  return runId;
}
