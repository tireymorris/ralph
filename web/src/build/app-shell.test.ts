import { describe, expect, it } from "vitest";
import { mediaBlock, readIndexCss, ruleBlock } from "./css-test-utils";

describe("index.css app shell", () => {
  it("sets .app-shell grid to --sidebar-width token", () => {
    const block = ruleBlock(readIndexCss(), ".app-shell");
    expect(block).toMatch(/grid-template-columns:\s*var\(--sidebar-width\)\s+1fr/);
  });

  it("adds 150ms background transition on nav links", () => {
    const block = ruleBlock(readIndexCss(), ".app-nav a");
    expect(block).toMatch(/transition:[^;]*background[^;]*150ms/);
  });

  it("adds accent color on current nav link", () => {
    const block = ruleBlock(readIndexCss(), '.app-nav a[aria-current="page"]');
    expect(block).toMatch(/color:\s*var\(--accent\)/);
  });

  it("offsets .follow-up-composer left to match sidebar width", () => {
    const block = ruleBlock(readIndexCss(), ".follow-up-composer");
    expect(block).toMatch(/left:\s*var\(--sidebar-width\)/);
  });
});

describe("index.css app shell at 768px", () => {
  const narrowCss = () =>
    mediaBlock(readIndexCss(), "(max-width: 768px)") ?? "";

  it("stacks .app-shell to a single column", () => {
    const block = ruleBlock(narrowCss(), ".app-shell");
    expect(block).toMatch(/grid-template-columns:\s*1fr/);
  });

  it("resets .follow-up-composer left offset for full width", () => {
    const block = ruleBlock(narrowCss(), ".follow-up-composer");
    expect(block).toMatch(/left:\s*0/);
  });

  it("caps .runs-list height with scroll", () => {
    const block = ruleBlock(narrowCss(), ".runs-list");
    expect(block).toMatch(/max-height:\s*200px/);
    expect(block).toMatch(/overflow-y:\s*auto/);
  });
});
