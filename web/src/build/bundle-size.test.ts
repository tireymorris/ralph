import { gzipSync } from "node:zlib";
import { existsSync, readdirSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const distAssets = resolve(
  import.meta.dirname,
  "../../../internal/web/static/dist/assets",
);

const MAX_GZIP_BYTES = 500 * 1024;

describe("production bundle size", () => {
  it("main js chunk is at most 500kb gzipped", () => {
    if (!existsSync(distAssets)) {
      throw new Error(
        "dist/assets missing; run `cd web && npm run build` before this test",
      );
    }

    const jsFiles = readdirSync(distAssets).filter((name) => name.endsWith(".js"));
    expect(jsFiles.length).toBeGreaterThan(0);

    const mainFile =
      jsFiles.find((name) => /^index-/.test(name)) ?? jsFiles[0];
    const raw = readFileSync(resolve(distAssets, mainFile));
    const gzipped = gzipSync(raw);

    expect(gzipped.length).toBeLessThanOrEqual(MAX_GZIP_BYTES);
  });
});
