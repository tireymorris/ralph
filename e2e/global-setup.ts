import { buildBinary, buildFrontend } from "./helpers/server.ts";

export default function globalSetup() {
  if (process.env.SKIP_BUILD !== "1") {
    buildFrontend({ VITE_RUN_STALL_THRESHOLD_MS: "3000" });
    buildBinary();
  }
}
