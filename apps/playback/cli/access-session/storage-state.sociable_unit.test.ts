import { describe, expect, it } from "vitest";
import {
  ACCESS_COOKIE_NAME,
  ACCESS_SESSION_WARNING_SEC,
  checkAccessSession,
  judgeAccessSession,
  normalizeAccessSession,
} from "./storage-state.ts";

const EXPIRES_AT_SEC = 1_800_000_000;
const EXPIRES_AT_ISO = "2027-01-15T08:00:00.000Z";

function createAccessCookie(expiresAtSec: number) {
  return {
    name: ACCESS_COOKIE_NAME,
    value: "fake-cookie-value",
    domain: "playback.example.workers.dev",
    path: "/",
    expires: expiresAtSec,
    httpOnly: true,
    secure: true,
    sameSite: "Lax",
  };
}

function createStorageStateJson(expiresAtSec: number): string {
  return JSON.stringify({ cookies: [createAccessCookie(expiresAtSec)], origins: [] });
}

describe("checkAccessSession", () => {
  it("reports_the_expiry_and_remaining_seconds_when_the_access_cookie_is_present", () => {
    // Given: Access cookie を持つ storageState の JSON と、期限の 100 秒前の現在時刻
    const json = createStorageStateJson(EXPIRES_AT_SEC);

    // When: 期限を確認する
    const status = checkAccessSession(json, EXPIRES_AT_SEC - 100);

    // Then: 期限と残り秒が返る
    expect(status).toEqual({ expiresAtSec: EXPIRES_AT_SEC, remainingSec: 100 });
  });

  it("reports_zero_remaining_seconds_when_now_equals_the_expiry", () => {
    // Given: Access cookie を持つ storageState の JSON
    const json = createStorageStateJson(EXPIRES_AT_SEC);

    // When: 期限ちょうどの時刻で確認する
    const status = checkAccessSession(json, EXPIRES_AT_SEC);

    // Then: 残りは 0 秒
    expect(status.remainingSec).toBe(0);
  });

  it("finds_the_access_cookie_when_other_cookies_are_present", () => {
    // Given: 他の cookie の後ろに Access cookie がある storageState（人が cookie を足した場合）
    const json = JSON.stringify({
      cookies: [{ name: "CF_Other", value: "x", expires: 1 }, createAccessCookie(EXPIRES_AT_SEC)],
      origins: [],
    });

    // When: 期限を確認する
    const status = checkAccessSession(json, EXPIRES_AT_SEC - 100);

    // Then: 他の cookie ではなく Access cookie の期限を読む
    expect(status.expiresAtSec).toBe(EXPIRES_AT_SEC);
  });

  it("rejects_a_storage_state_without_the_access_cookie", () => {
    // Given: Access cookie を持たない storageState
    const json = JSON.stringify({ cookies: [], origins: [] });

    // When / Then: 期限を読めないので拒否する
    expect(() => checkAccessSession(json, EXPIRES_AT_SEC)).toThrow(ACCESS_COOKIE_NAME);
  });

  it("rejects_input_that_is_not_json", () => {
    // Given: JSON ではない文字列
    const json = "not json";

    // When / Then: 拒否する
    expect(() => checkAccessSession(json, EXPIRES_AT_SEC)).toThrow("JSON");
  });

  it("rejects_a_storage_state_that_is_json_but_not_an_object", () => {
    // Given: JSON としては正しいが object ではない値（null）
    const json = "null";

    // When / Then: cookie を読めないので拒否する
    expect(() => checkAccessSession(json, EXPIRES_AT_SEC)).toThrow(ACCESS_COOKIE_NAME);
  });

  it("rejects_a_storage_state_without_a_cookies_field", () => {
    // Given: cookies を持たない object
    const json = JSON.stringify({ origins: [] });

    // When / Then: cookie を読めないので拒否する
    expect(() => checkAccessSession(json, EXPIRES_AT_SEC)).toThrow(ACCESS_COOKIE_NAME);
  });

  it("rejects_a_storage_state_whose_cookies_field_is_not_an_array", () => {
    // Given: cookies が配列ではない object
    const json = JSON.stringify({ cookies: "not an array" });

    // When / Then: cookie を読めないので拒否する
    expect(() => checkAccessSession(json, EXPIRES_AT_SEC)).toThrow(ACCESS_COOKIE_NAME);
  });

  it("rejects_an_access_cookie_whose_expiry_is_not_a_number", () => {
    // Given: Access cookie はあるが、expires が数値ではない
    const json = JSON.stringify({ cookies: [{ name: ACCESS_COOKIE_NAME, expires: "soon" }] });

    // When / Then: 期限を読めないので拒否する
    expect(() => checkAccessSession(json, EXPIRES_AT_SEC)).toThrow(ACCESS_COOKIE_NAME);
  });
});

describe("expires given as a DevTools date string", () => {
  function createStateWithExpires(expires: unknown): string {
    return JSON.stringify({
      cookies: [{ ...createAccessCookie(0), expires }],
      origins: [],
    });
  }

  it("checks_the_expiry_of_an_iso_date_string_as_unix_seconds", () => {
    // Given: DevTools の Expires 列の日時（小数秒と Z 付き）をそのまま expires に貼った storageState
    const json = createStateWithExpires(EXPIRES_AT_ISO);
    const expiresAtSec = Date.parse(EXPIRES_AT_ISO) / 1000;

    // When: 期限を確認する
    const status = checkAccessSession(json, expiresAtSec - 100);

    // Then: Unix 秒に直した期限と残り秒が返る
    expect(status).toEqual({ expiresAtSec, remainingSec: 100 });
  });

  it("rejects_an_access_cookie_without_an_expires_field", () => {
    // Given: expires を持たない Access cookie
    const json = createStateWithExpires(undefined);

    // When / Then: 期限を読めないので拒否する
    expect(() => checkAccessSession(json, EXPIRES_AT_SEC)).toThrow(ACCESS_COOKIE_NAME);
  });

  it("rejects_a_string_that_is_not_an_iso_date", () => {
    // Given: 日時ではない文字列（placeholder の置き換え忘れ）
    const json = createStateWithExpires("<期限の日時>");

    // When / Then: 期限を読めないので拒否する
    expect(() => checkAccessSession(json, EXPIRES_AT_SEC)).toThrow(ACCESS_COOKIE_NAME);
  });

  it("rejects_a_string_that_starts_like_a_date_but_cannot_be_parsed", () => {
    // Given: 日時の形に見えるが、存在しない日時
    const json = createStateWithExpires("2027-13-45T99:99:99Z");

    // When / Then: 期限を読めないので拒否する
    expect(() => checkAccessSession(json, EXPIRES_AT_SEC)).toThrow(ACCESS_COOKIE_NAME);
  });
});

describe("normalizeAccessSession", () => {
  it("rewrites_an_iso_expiry_to_unix_seconds_and_keeps_the_other_fields", () => {
    // Given: expires に日時の文字列を貼った storageState
    const json = JSON.stringify({
      cookies: [{ ...createAccessCookie(0), expires: EXPIRES_AT_ISO }],
      origins: [],
    });

    // When: Playwright へ渡せる形へ直す
    const normalized = JSON.parse(normalizeAccessSession(json));

    // Then: expires だけが Unix 秒（数値）になり、他の項目は変わらない
    expect(normalized).toEqual({
      cookies: [createAccessCookie(Date.parse(EXPIRES_AT_ISO) / 1000)],
      origins: [],
    });
  });

  it("keeps_a_numeric_expiry_as_it_is", () => {
    // Given: expires が既に Unix 秒の storageState
    const json = createStorageStateJson(EXPIRES_AT_SEC);

    // When: 直す
    const normalized = JSON.parse(normalizeAccessSession(json));

    // Then: 期限は変わらない
    expect(normalized.cookies[0].expires).toBe(EXPIRES_AT_SEC);
  });

  it("rewrites_only_the_access_cookie_when_other_cookies_are_present", () => {
    // Given: 他の cookie の後ろに Access cookie がある storageState
    const other = { name: "CF_Other", value: "x", expires: "keep-me" };
    const json = JSON.stringify({
      cookies: [other, { ...createAccessCookie(0), expires: EXPIRES_AT_ISO }],
      origins: [],
    });

    // When: 直す
    const normalized = JSON.parse(normalizeAccessSession(json));

    // Then: 他の cookie は変わらず、Access cookie だけが直る
    expect(normalized.cookies[0]).toEqual(other);
    expect(normalized.cookies[1].expires).toBe(Date.parse(EXPIRES_AT_ISO) / 1000);
  });

  it("rejects_a_storage_state_whose_expiry_cannot_be_read", () => {
    // Given: 期限を読めない storageState
    const json = JSON.stringify({ cookies: [{ name: ACCESS_COOKIE_NAME, expires: "soon" }] });

    // When / Then: 直せないので拒否する
    expect(() => normalizeAccessSession(json)).toThrow(ACCESS_COOKIE_NAME);
  });

  it("rejects_input_that_is_not_json", () => {
    // Given: JSON ではない文字列

    // When / Then: 拒否する
    expect(() => normalizeAccessSession("not json")).toThrow("JSON");
  });
});

describe("judgeAccessSession", () => {
  it("returns_alive_when_more_than_the_warning_window_remains", () => {
    // Given: 警告の窓を 1 秒超える残り
    const remainingSec = ACCESS_SESSION_WARNING_SEC + 1;

    // When: 判定する
    const verdict = judgeAccessSession(remainingSec);

    // Then: 有効
    expect(verdict).toBe("alive");
  });

  it("returns_alive_when_the_remaining_time_equals_the_warning_window", () => {
    // Given: 警告の窓とちょうど同じ残り
    const remainingSec = ACCESS_SESSION_WARNING_SEC;

    // When: 判定する
    const verdict = judgeAccessSession(remainingSec);

    // Then: 有効（窓に入るのは窓より短い時）
    expect(verdict).toBe("alive");
  });

  it("returns_expiring_when_the_remaining_time_is_shorter_than_the_warning_window", () => {
    // Given: 警告の窓より 1 秒短い残り
    const remainingSec = ACCESS_SESSION_WARNING_SEC - 1;

    // When: 判定する
    const verdict = judgeAccessSession(remainingSec);

    // Then: 失効間近
    expect(verdict).toBe("expiring");
  });

  it("returns_expired_when_no_time_remains", () => {
    // Given: 残りが 0 秒
    const remainingSec = 0;

    // When: 判定する
    const verdict = judgeAccessSession(remainingSec);

    // Then: 失効
    expect(verdict).toBe("expired");
  });

  it("returns_expired_when_the_remaining_time_is_negative", () => {
    // Given: 期限を過ぎた残り
    const remainingSec = -1;

    // When: 判定する
    const verdict = judgeAccessSession(remainingSec);

    // Then: 失効
    expect(verdict).toBe("expired");
  });
});
