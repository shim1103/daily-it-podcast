/**
 * Access session（Playwright の storageState）の登録と期限確認の入口。
 * 組み立てと判定は access-session-storage-state.ts が持つ。ここは入出力だけを担う。
 *
 *   build <origin>  標準入力の CF_Authorization（JWT）から storageState の JSON を標準出力へ書く
 *   check           env PLAYWRIGHT_STORAGE_STATE_JSON の期限を確認し、失効なら終了 status 1 を返す
 *
 * @require Node（apps/playback/.nvmrc）。`--experimental-strip-types` で実行する。
 * @invariant cookie の値を標準出力（build の JSON を除く）・標準エラー・log へ出さない。
 */
import {
  buildAccessSessionStorageState,
  checkAccessSession,
  judgeAccessSession,
  SECONDS_PER_DAY,
} from "./access-session-storage-state.ts";

const EXIT_OK = 0;
const EXIT_FAILED = 1;
const EXIT_USAGE = 2;

const MILLISECONDS_PER_SECOND = 1000;

// why: 標準出力と標準エラーが、この CLI の出力契約（build の JSON は標準出力だけ）。console を使わず直接書く
function writeStdout(line: string): void {
  process.stdout.write(`${line}\n`);
}

function writeStderr(line: string): void {
  process.stderr.write(`${line}\n`);
}

function nowSec(): number {
  return Math.floor(Date.now() / MILLISECONDS_PER_SECOND);
}

function formatExpiry(expiresAtSec: number, remainingSec: number): string {
  const expiresAt = new Date(expiresAtSec * MILLISECONDS_PER_SECOND).toISOString();
  const remainingDays = Math.floor(remainingSec / SECONDS_PER_DAY);
  return `期限 ${expiresAt}（残り ${remainingDays} 日）`;
}

async function readStdin(): Promise<string> {
  const chunks: Buffer[] = [];
  for await (const chunk of process.stdin) {
    chunks.push(chunk);
  }
  return Buffer.concat(chunks).toString("utf8").trim();
}

async function runBuild(origin: string | undefined): Promise<number> {
  if (origin === undefined || origin === "") {
    writeStderr("使い方: build <origin>（CF_Authorization を標準入力から渡す）");
    return EXIT_USAGE;
  }
  const storageStateJson = JSON.stringify(
    buildAccessSessionStorageState(origin, await readStdin()),
  );
  const status = checkAccessSession(storageStateJson, nowSec());
  const expiry = formatExpiry(status.expiresAtSec, status.remainingSec);
  if (judgeAccessSession(status.remainingSec) === "expired") {
    writeStderr(`渡された session は既に失効している。${expiry}`);
    return EXIT_FAILED;
  }
  writeStderr(`storageState を組み立てた。${expiry}`);
  writeStdout(storageStateJson);
  return EXIT_OK;
}

function runCheck(): number {
  const storageStateJson = process.env.PLAYWRIGHT_STORAGE_STATE_JSON?.trim();
  if (storageStateJson === undefined || storageStateJson === "") {
    writeStdout(
      "PLAYWRIGHT_STORAGE_STATE_JSON が未設定のため、remote e2e の session 確認は skip する",
    );
    return EXIT_OK;
  }
  const status = checkAccessSession(storageStateJson, nowSec());
  const expiry = formatExpiry(status.expiresAtSec, status.remainingSec);
  const verdict = judgeAccessSession(status.remainingSec);
  if (verdict === "expired") {
    writeStderr(`Access session が失効している。${expiry}`);
    writeStderr(
      "DEPLOY.md「storageState 更新」の手順で、Secret PLAYWRIGHT_STORAGE_STATE_JSON を更新する",
    );
    return EXIT_FAILED;
  }
  writeStdout(`Access session は有効${verdict === "expiring" ? "だが、失効間近" : ""}。${expiry}`);
  return EXIT_OK;
}

async function main(args: string[]): Promise<number> {
  const [command, origin] = args;
  if (command === "build") {
    return runBuild(origin);
  }
  if (command === "check") {
    return runCheck();
  }
  writeStderr("使い方: access-session-cli.ts <build <origin> | check>");
  return EXIT_USAGE;
}

try {
  process.exitCode = await main(process.argv.slice(2));
} catch (error) {
  writeStderr(error instanceof Error ? error.message : "想定外の失敗");
  process.exitCode = EXIT_FAILED;
}
