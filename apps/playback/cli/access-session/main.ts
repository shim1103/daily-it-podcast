/**
 * Access session（Playwright の storageState）の期限確認の入口。
 * process の env・時刻・出力を command へ結線するだけで、判断も分岐も持たない。
 * 出力の実体は worker/src/routes/logger.ts の 1 点（console を呼ぶ唯一の file）に任せる。
 *
 *   check  env PLAYWRIGHT_STORAGE_STATE_JSON の期限を確認し、失効なら終了 status 1 を返す
 *
 * storageState の JSON は人が組み立てて登録する（構造は storage-state.example.json、手順は DEPLOY.md）。
 *
 * @require Node（apps/playback/.nvmrc）。`--experimental-strip-types` で実行する。
 * @invariant cookie の値を出力しない。
 */
import { writeErrorLine, writeLine } from "../../worker/src/routes/logger.ts";
import { runAccessSessionCommand } from "./command.ts";

const MILLISECONDS_PER_SECOND = 1000;

process.exitCode = await runAccessSessionCommand(process.argv.slice(2), {
  storageStateEnv: process.env.PLAYWRIGHT_STORAGE_STATE_JSON,
  nowSec: () => Math.floor(Date.now() / MILLISECONDS_PER_SECOND),
  writeStdout: writeLine,
  writeStderr: writeErrorLine,
});
