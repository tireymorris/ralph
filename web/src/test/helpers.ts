import { vi } from "vitest";

export function stubConfirm(accepted: boolean) {
  const confirm = vi.fn(() => accepted);
  vi.stubGlobal("confirm", confirm);
  return confirm;
}
