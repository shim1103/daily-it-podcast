/** Cloudflare Access が入場後に付ける認可 cookie の名前。 */
export const ACCESS_COOKIE_NAME = "CF_Authorization";

export const SECONDS_PER_DAY = 24 * 60 * 60;

/** 残りがこれより短い間は、失効間近として警告する。 */
export const ACCESS_SESSION_WARNING_SEC = 7 * SECONDS_PER_DAY;

const MILLISECONDS_PER_SECOND = 1000;

/** DevTools の Expires 列の日時（例 `2026-10-30T16:44:30.650Z`）の見分け方。 */
const ISO_DATE_PATTERN = /^\d{4}-\d{2}-\d{2}T/;

export type AccessSessionStatus = { expiresAtSec: number; remainingSec: number };

export type AccessSessionVerdict = "alive" | "expiring" | "expired";

type CookieLike = { name?: unknown; expires?: unknown };

type StorageStateLike = { cookies: CookieLike[] };

type AccessSession = { state: StorageStateLike; expiresAtSec: number };

function parseStorageState(storageStateJson: string): unknown {
  try {
    return JSON.parse(storageStateJson);
  } catch {
    throw new Error("storageState を JSON として読めない");
  }
}

function hasCookies(state: unknown): state is StorageStateLike {
  return (
    typeof state === "object" &&
    state !== null &&
    "cookies" in state &&
    Array.isArray(state.cookies)
  );
}

/** expires を Unix 秒で返す。数値はそのまま、日時の文字列は Unix 秒（小数秒は切り捨て）に直す。 */
function readExpirySec(expires: unknown): number | undefined {
  if (typeof expires === "number") {
    return expires;
  }
  if (typeof expires === "string" && ISO_DATE_PATTERN.test(expires)) {
    const expiresAtMs = Date.parse(expires);
    return Number.isNaN(expiresAtMs)
      ? undefined
      : Math.floor(expiresAtMs / MILLISECONDS_PER_SECOND);
  }
  return undefined;
}

function readAccessSession(storageStateJson: string): AccessSession {
  const state = parseStorageState(storageStateJson);
  if (hasCookies(state)) {
    const accessCookie = state.cookies.find((cookie) => cookie.name === ACCESS_COOKIE_NAME);
    const expiresAtSec = readExpirySec(accessCookie?.expires);
    if (expiresAtSec !== undefined) {
      return { state, expiresAtSec };
    }
  }
  throw new Error(`storageState に ${ACCESS_COOKIE_NAME} cookie の期限が無い`);
}

/** storageState の JSON から、Access cookie の期限と、現在時刻からの残り秒を求める。 */
export function checkAccessSession(storageStateJson: string, nowSec: number): AccessSessionStatus {
  const { expiresAtSec } = readAccessSession(storageStateJson);
  return { expiresAtSec, remainingSec: expiresAtSec - nowSec };
}

/** Access cookie の expires を Unix 秒（数値）へ直した、Playwright が読める storageState の JSON を返す。 */
export function normalizeAccessSession(storageStateJson: string): string {
  const { state, expiresAtSec } = readAccessSession(storageStateJson);
  const cookies = state.cookies.map((cookie) =>
    cookie.name === ACCESS_COOKIE_NAME ? { ...cookie, expires: expiresAtSec } : cookie,
  );
  return JSON.stringify({ ...state, cookies });
}

export function judgeAccessSession(remainingSec: number): AccessSessionVerdict {
  if (remainingSec <= 0) {
    return "expired";
  }
  return remainingSec < ACCESS_SESSION_WARNING_SEC ? "expiring" : "alive";
}
