/** Cloudflare Access が入場後に付ける認可 cookie の名前。 */
export const ACCESS_COOKIE_NAME = "CF_Authorization";

export const SECONDS_PER_DAY = 24 * 60 * 60;

/** 残りがこれより短い間は、失効間近として警告する。 */
export const ACCESS_SESSION_WARNING_SEC = 7 * SECONDS_PER_DAY;

const JWT_SEGMENT_COUNT = 3;
const JWT_PAYLOAD_INDEX = 1;

export type StorageStateCookie = {
  name: string;
  value: string;
  domain: string;
  path: string;
  expires: number;
  httpOnly: boolean;
  secure: boolean;
  sameSite: "Lax";
};

type StorageStateLocalStorageItem = { name: string; value: string };

export type StorageStateOrigin = {
  origin: string;
  localStorage: StorageStateLocalStorageItem[];
};

/** Playwright の `storageState` に渡せる JSON 形。 */
export type AccessSessionStorageState = {
  cookies: StorageStateCookie[];
  origins: StorageStateOrigin[];
};

export type AccessSessionStatus = { expiresAtSec: number; remainingSec: number };

export type AccessSessionVerdict = "alive" | "expiring" | "expired";

type JwtPayloadWithExp = { exp: number };

type CookieLike = { name?: unknown; expires?: unknown };

function hasNumericExp(payload: unknown): payload is JwtPayloadWithExp {
  return (
    typeof payload === "object" &&
    payload !== null &&
    "exp" in payload &&
    typeof payload.exp === "number"
  );
}

function readJwtExpiry(jwt: string): number {
  const segments = jwt.split(".");
  if (segments.length !== JWT_SEGMENT_COUNT) {
    throw new Error("値が JWT の形（3 segment）ではない");
  }
  let payload: unknown;
  try {
    payload = JSON.parse(Buffer.from(segments[JWT_PAYLOAD_INDEX], "base64url").toString("utf8"));
  } catch {
    throw new Error("JWT の payload を JSON として読めない");
  }
  if (!hasNumericExp(payload)) {
    throw new Error("JWT の payload に数値の exp が無い");
  }
  return payload.exp;
}

function readHostname(origin: string): string {
  try {
    return new URL(origin).hostname;
  } catch {
    throw new Error("origin が URL として読めない");
  }
}

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

/** Access の JWT（`CF_Authorization` の値）から、Playwright の storageState を組み立てる。 */
export function buildAccessSessionStorageState(
  origin: string,
  jwt: string,
): AccessSessionStorageState {
  const hostname = readHostname(origin);
  const expires = readJwtExpiry(jwt);
  return {
    cookies: [
      {
        name: ACCESS_COOKIE_NAME,
        value: jwt,
        domain: hostname,
        path: "/",
        expires,
        httpOnly: true,
        secure: true,
        sameSite: "Lax",
      },
    ],
    origins: [],
  };
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
