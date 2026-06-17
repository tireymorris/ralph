import { describe, expect, it } from "vitest";
import { mediaBlock, readIndexCss, ruleBlock } from "./css-test-utils";

describe("index.css color scheme", () => {
  it("supports light and dark via prefers-color-scheme", () => {
    const css = readIndexCss();
    expect(css).toMatch(/color-scheme:\s*light\s+dark/);
    const dark = mediaBlock(css, "(prefers-color-scheme: dark)") ?? "";
    expect(dark).toMatch(/--bg:\s*#09090b/);
    expect(ruleBlock(css, ":root")).toMatch(/--bg:\s*#f8f8fa/);
  });
});

describe("index.css app shell", () => {
  it("fills viewport with a full-width single-column main", () => {
    const block = ruleBlock(readIndexCss(), ".app-main");
    expect(block).toMatch(/height:\s*100vh/);
    expect(block).toMatch(/width:\s*100%/);
    expect(block).not.toMatch(/max-width:/);
  });

  it("anchors follow-up composer to full viewport width", () => {
    const block = ruleBlock(readIndexCss(), ".follow-up-composer");
    expect(block).toMatch(/left:\s*0/);
    expect(block).toMatch(/right:\s*0/);
  });

  it("styles run toolbar as a single compact row", () => {
    const block = ruleBlock(readIndexCss(), ".run-detail-toolbar");
    expect(block).toMatch(/display:\s*flex/);
    expect(block).toMatch(/align-items:\s*center/);
  });

  it("scrolls run content below the toolbar", () => {
    const block = ruleBlock(readIndexCss(), ".run-detail-body");
    expect(block).toMatch(/overflow-y:\s*auto/);
    expect(block).toMatch(/flex:\s*1/);
  });
});
