/** Cloudflare Access が入場後に付ける認可 cookie の名前。 */
export const ACCESS_COOKIE_NAME = "CF_Authorization";

export const SECONDS_PER_DAY = 24 * 60 * 60;

/** 残りがこれより短い間は、失効間近として警告する。 */
export const ACCESS_SESSION_WARNING_SEC = 7 * SECONDS_PER_DAY;

export type AccessSessionStatus = { expiresAtSec: number; remainingSec: number };

export type AccessSessionVerdict = "alive" | "expiring" | "expired";

type CookieLike = { name?: unknown; expires?: unknown };

function findAccessCookieExpiry(state: unknown): number | undefined {
  if (
    typeof state !== "object" ||
    state === null ||
    !("cookies" in state) ||
    !Array.isArray(state.cookies)
  ) {
    return undefined;
  }
  const cookies: CookieLike[] = state.cookies;
  const accessCookie = cookies.find((cookie) => cookie.name === ACCESS_COOKIE_NAME);
  return typeof accessCookie?.expires === "number" ? accessCookie.expires : undefined;
}

/** storageState の JSON から、Access cookie の期限と、現在時刻からの残り秒を求める。 */
export function checkAccessSession(storageStateJson: string, nowSec: number): AccessSessionStatus {
  let state: unknown;
  try {
    state = JSON.parse(storageStateJson);
  } catch {
    throw new Error("storageState を JSON として読めない");
  }
  const expiresAtSec = findAccessCookieExpiry(state);
  if (expiresAtSec === undefined) {
    throw new Error(`storageState に ${ACCESS_COOKIE_NAME} cookie の期限が無い`);
  }
  return { expiresAtSec, remainingSec: expiresAtSec - nowSec };
}

export function judgeAccessSession(remainingSec: number): AccessSessionVerdict {
  if (remainingSec <= 0) {
    return "expired";
  }
  return remainingSec < ACCESS_SESSION_WARNING_SEC ? "expiring" : "alive";
}
