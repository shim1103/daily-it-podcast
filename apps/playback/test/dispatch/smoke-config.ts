import path from "node:path";
import { fileURLToPath } from "node:url";

export const appRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");

/** 疎通は常設 smoke と同じ共有 config を使う。専用 config で実 config の欠陥を隠さない。 */
export const smokeConfigPath = path.join(appRoot, "test/support/wrangler.smoke.jsonc");
