import { describe, expect, it } from "vitest";
import { readIndexCss, ruleBlock } from "./css-test-utils";

describe("index.css app shell", () => {
  it("fills viewport with a centered single-column main", () => {
    const block = ruleBlock(readIndexCss(), ".app-main");
    expect(block).toMatch(/height:\s*100vh/);
    expect(block).toMatch(/max-width:\s*48rem/);
    expect(block).toMatch(/margin:\s*0\s+auto/);
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
});
