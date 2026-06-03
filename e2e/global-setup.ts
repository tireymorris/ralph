import { buildBinary, buildFrontend } from "./helpers/server.ts";

export default function globalSetup() {
  if (process.env.SKIP_BUILD !== "1") {
    buildFrontend();
    buildBinary();
  }
}
