import { describe, expect, it } from "vitest";
import {
  ACCESS_COOKIE_NAME,
  ACCESS_SESSION_WARNING_SEC,
  checkAccessSession,
  judgeAccessSession,
} from "./storage-state.ts";

const EXPIRES_AT_SEC = 1_800_000_000;

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
