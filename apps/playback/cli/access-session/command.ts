import {
  checkAccessSession,
  judgeAccessSession,
  normalizeAccessSession,
  SECONDS_PER_DAY,
} from "./storage-state.ts";

export const EXIT_OK = 0;
export const EXIT_FAILED = 1;
export const EXIT_USAGE = 2;

const MILLISECONDS_PER_SECOND = 1000;

const UNEXPECTED_FAILURE_MESSAGE = "想定外の失敗";

/** 副作用と非決定な値を、呼び出し側（main.ts）から注入する。cookie の値は出力しない。 */
export type AccessSessionCommandIo = {
  /** env PLAYWRIGHT_STORAGE_STATE_JSON の値。未設定なら undefined。 */
  storageStateEnv: string | undefined;
  nowSec: () => number;
  writeFile: (filePath: string, content: string) => void;
  writeStdout: (line: string) => void;
  writeStderr: (line: string) => void;
};

function formatExpiry(expiresAtSec: number, remainingSec: number): string {
  const expiresAt = new Date(expiresAtSec * MILLISECONDS_PER_SECOND).toISOString();
  const remainingDays = Math.floor(remainingSec / SECONDS_PER_DAY);
  return `期限 ${expiresAt}（残り ${remainingDays} 日）`;
}

function readStorageStateEnv(io: AccessSessionCommandIo): string | undefined {
  const storageStateJson = io.storageStateEnv?.trim();
  return storageStateJson === "" ? undefined : storageStateJson;
}

function runCheck(io: AccessSessionCommandIo): number {
  const storageStateJson = readStorageStateEnv(io);
  if (storageStateJson === undefined) {
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
    `Access session の期限は有効${verdict === "expiring" ? "だが、失効間近" : ""}。${expiry}。期限だけの確認で、Access が受理するかは E2E が確かめる`,
  );
  return EXIT_OK;
}

function runNormalize(outputPath: string | undefined, io: AccessSessionCommandIo): number {
  if (outputPath === undefined || outputPath === "") {
    io.writeStderr("使い方: normalize <出力先の path>");
    return EXIT_USAGE;
  }
  const storageStateJson = readStorageStateEnv(io);
  if (storageStateJson === undefined) {
    io.writeStderr("PLAYWRIGHT_STORAGE_STATE_JSON が未設定のため、storageState を書けない");
    return EXIT_FAILED;
  }
  io.writeFile(outputPath, normalizeAccessSession(storageStateJson));
  io.writeStdout(`storageState を ${outputPath} へ書いた`);
  return EXIT_OK;
}

function dispatch(args: string[], io: AccessSessionCommandIo): number {
  const [command, outputPath] = args;
  if (command === "check") {
    return runCheck(io);
  }
  if (command === "normalize") {
    return runNormalize(outputPath, io);
  }
  io.writeStderr("使い方: <check | normalize <出力先の path>>");
  return EXIT_USAGE;
}

/** subcommand（check / normalize）を実行し、終了 status を返す。失敗は標準エラーへ書いて EXIT_FAILED にする。 */
export async function runAccessSessionCommand(
  args: string[],
  io: AccessSessionCommandIo,
): Promise<number> {
  try {
    return dispatch(args, io);
  } catch (error) {
    io.writeStderr(error instanceof Error ? error.message : UNEXPECTED_FAILURE_MESSAGE);
    return EXIT_FAILED;
  }
}
