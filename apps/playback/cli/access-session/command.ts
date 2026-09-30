import {
  buildAccessSessionStorageState,
  checkAccessSession,
  judgeAccessSession,
  SECONDS_PER_DAY,
} from "./storage-state.ts";

export const EXIT_OK = 0;
export const EXIT_FAILED = 1;
export const EXIT_USAGE = 2;

const MILLISECONDS_PER_SECOND = 1000;

const UNEXPECTED_FAILURE_MESSAGE = "想定外の失敗";

/** 副作用と非決定な値を、呼び出し側（main.ts）から注入する。cookie の値は build の JSON を除いて出さない。 */
export type AccessSessionCommandIo = {
  readStdin: () => Promise<string>;
  /** env PLAYWRIGHT_STORAGE_STATE_JSON の値。未設定なら undefined。 */
  storageStateEnv: string | undefined;
  nowSec: () => number;
  writeStdout: (line: string) => void;
  writeStderr: (line: string) => void;
};

function formatExpiry(expiresAtSec: number, remainingSec: number): string {
  const expiresAt = new Date(expiresAtSec * MILLISECONDS_PER_SECOND).toISOString();
  const remainingDays = Math.floor(remainingSec / SECONDS_PER_DAY);
  return `期限 ${expiresAt}（残り ${remainingDays} 日）`;
}

async function runBuild(origin: string | undefined, io: AccessSessionCommandIo): Promise<number> {
  if (origin === undefined || origin === "") {
    io.writeStderr("使い方: build <origin>（CF_Authorization を標準入力から渡す）");
    return EXIT_USAGE;
  }
  const storageStateJson = JSON.stringify(
    buildAccessSessionStorageState(origin, await io.readStdin()),
  );
  const status = checkAccessSession(storageStateJson, io.nowSec());
  const expiry = formatExpiry(status.expiresAtSec, status.remainingSec);
  if (judgeAccessSession(status.remainingSec) === "expired") {
    io.writeStderr(`渡された session は既に失効している。${expiry}`);
    return EXIT_FAILED;
  }
  io.writeStderr(`storageState を組み立てた。${expiry}`);
  io.writeStdout(storageStateJson);
  return EXIT_OK;
}

function runCheck(io: AccessSessionCommandIo): number {
  const storageStateJson = io.storageStateEnv?.trim();
  if (storageStateJson === undefined || storageStateJson === "") {
    io.writeStdout(
      "PLAYWRIGHT_STORAGE_STATE_JSON が未設定のため、remote e2e の session 確認は skip する",
    );
    return EXIT_OK;
  }
  const status = checkAccessSession(storageStateJson, io.nowSec());
  const expiry = formatExpiry(status.expiresAtSec, status.remainingSec);
  const verdict = judgeAccessSession(status.remainingSec);
  if (verdict === "expired") {
    io.writeStderr(`Access session が失効している。${expiry}`);
    io.writeStderr(
      "DEPLOY.md「storageState 更新」の手順で、Secret PLAYWRIGHT_STORAGE_STATE_JSON を更新する",
    );
    return EXIT_FAILED;
  }
  io.writeStdout(
    `Access session は有効${verdict === "expiring" ? "だが、失効間近" : ""}。${expiry}`,
  );
  return EXIT_OK;
}

async function dispatch(args: string[], io: AccessSessionCommandIo): Promise<number> {
  const [command, origin] = args;
  if (command === "build") {
    return runBuild(origin, io);
  }
  if (command === "check") {
    return runCheck(io);
  }
  io.writeStderr("使い方: <build <origin> | check>");
  return EXIT_USAGE;
}

/** subcommand（build / check）を実行し、終了 status を返す。失敗は標準エラーへ書いて EXIT_FAILED にする。 */
export async function runAccessSessionCommand(
  args: string[],
  io: AccessSessionCommandIo,
): Promise<number> {
  try {
    return await dispatch(args, io);
  } catch (error) {
    io.writeStderr(error instanceof Error ? error.message : UNEXPECTED_FAILURE_MESSAGE);
    return EXIT_FAILED;
  }
}
