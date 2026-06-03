import { describe, expect, it } from "vitest";
import {
  focusVisibleBlock,
  mediaBlock,
  readIndexCss,
  ruleBlock,
} from "./css-test-utils";

describe("index.css focus-visible rings", () => {
  it("styles button:focus-visible with accent outline", () => {
    const block = focusVisibleBlock(readIndexCss(), "button");
    expect(block).toMatch(/outline:\s*2px solid var\(--accent\)/);
    expect(block).toMatch(/outline-offset:\s*2px/);
  });

  it("styles a:focus-visible with accent outline", () => {
    const block = focusVisibleBlock(readIndexCss(), "a");
    expect(block).toMatch(/outline:\s*2px solid var\(--accent\)/);
  });

  it("styles .composer-input:focus-visible with accent outline", () => {
    const block = focusVisibleBlock(readIndexCss(), ".composer-input");
    expect(block).toMatch(/outline:\s*2px solid var\(--accent\)/);
  });

  it("does not set outline on .composer-input:focus", () => {
    const block = ruleBlock(readIndexCss(), ".composer-input:focus");
    if (block === null) return;
    expect(block).not.toMatch(/outline:/);
  });
});

describe("index.css reduced motion", () => {
  it("zeros transition and animation durations when motion is reduced", () => {
    const block = mediaBlock(readIndexCss(), "(prefers-reduced-motion: reduce)");
    expect(block).not.toBeNull();
    expect(block).toMatch(/transition-duration:\s*0\.01ms\s*!important/);
    expect(block).toMatch(/animation-duration:\s*0\.01ms\s*!important/);
  });
});
