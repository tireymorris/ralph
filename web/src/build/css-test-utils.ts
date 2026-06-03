import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const indexCssPath = resolve(import.meta.dirname, "../index.css");

export function readIndexCss(): string {
  return readFileSync(indexCssPath, "utf8");
}

export function ruleBlock(css: string, selector: string): string | null {
  const escaped = selector.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const match = css.match(new RegExp(`${escaped}\\s*\\{([^}]*)\\}`, "s"));
  return match ? match[1] : null;
}

export function mediaBlock(css: string, condition: string): string | null {
  const needle = `@media ${condition}`;
  const start = css.indexOf(needle);
  if (start === -1) return null;
  const open = css.indexOf("{", start);
  if (open === -1) return null;
  let depth = 1;
  for (let i = open + 1; i < css.length && depth > 0; i++) {
    if (css[i] === "{") depth++;
    else if (css[i] === "}") depth--;
    if (depth === 0) return css.slice(open + 1, i);
  }
  return null;
}

export function focusVisibleBlock(
  css: string,
  pseudoHost: string,
): string | null {
  const escaped = pseudoHost.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const match = css.match(
    new RegExp(`${escaped}:focus-visible[^{]*\\{([^}]*)\\}`, "s"),
  );
  return match ? match[1] : null;
}
