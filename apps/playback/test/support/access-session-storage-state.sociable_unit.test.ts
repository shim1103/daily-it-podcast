import { describe, expect, it } from "vitest";
import {
  ACCESS_COOKIE_NAME,
  ACCESS_SESSION_WARNING_SEC,
  buildAccessSessionStorageState,
  checkAccessSession,
  judgeAccessSession,
} from "./access-session-storage-state.ts";
import { createFakeAccessJwt, createFakeAccessJwtWithoutExpiry } from "./fake-access-jwt.ts";

const ORIGIN = "https://playback.example.workers.dev";
const EXPIRES_AT_SEC = 1_800_000_000;

describe("buildAccessSessionStorageState", () => {
  it("builds_the_access_cookie_with_the_jwt_exp_as_expiry_when_given_a_valid_jwt", () => {
    // Given: exp を持つ JWT と本番 origin
    const jwt = createFakeAccessJwt(EXPIRES_AT_SEC);

    // When: storageState を組み立てる
    const state = buildAccessSessionStorageState(ORIGIN, jwt);

    // Then: origin の host に属する Access cookie が 1 つあり、期限は JWT の exp と一致する
    expect(state).toEqual({
      cookies: [
        {
          name: ACCESS_COOKIE_NAME,
          value: jwt,
          domain: "playback.example.workers.dev",
          path: "/",
          expires: EXPIRES_AT_SEC,
          httpOnly: true,
          secure: true,
          sameSite: "Lax",
        },
      ],
      origins: [],
    });
  });

  it("rejects_a_jwt_that_does_not_have_three_segments", () => {
    // Given: 2 segment しか無い値
    const notJwt = createFakeAccessJwt(EXPIRES_AT_SEC).split(".").slice(0, 2).join(".");

    // When / Then: JWT ではないので組み立てを拒否する
    expect(() => buildAccessSessionStorageState(ORIGIN, notJwt)).toThrow("JWT");
  });

  it("rejects_a_jwt_whose_payload_has_no_numeric_exp", () => {
    // Given: exp を持たない payload の JWT
    const jwt = createFakeAccessJwtWithoutExpiry();

    // When / Then: 期限が分からないので組み立てを拒否する
    expect(() => buildAccessSessionStorageState(ORIGIN, jwt)).toThrow("exp");
  });

  it("rejects_an_origin_that_is_not_a_url", () => {
    // Given: URL ではない origin
    const jwt = createFakeAccessJwt(EXPIRES_AT_SEC);

    // When / Then: host を取り出せないので組み立てを拒否する
    expect(() => buildAccessSessionStorageState("not a url", jwt)).toThrow("origin");
  });
});

describe("checkAccessSession", () => {
  it("reports_the_expiry_and_remaining_seconds_when_the_access_cookie_is_present", () => {
    // Given: 組み立て済みの storageState JSON と、期限の 100 秒前の現在時刻
    const json = JSON.stringify(
      buildAccessSessionStorageState(ORIGIN, createFakeAccessJwt(EXPIRES_AT_SEC)),
    );

    // When: 期限を確認する
    const status = checkAccessSession(json, EXPIRES_AT_SEC - 100);

    // Then: 期限と残り秒が返る
    expect(status).toEqual({ expiresAtSec: EXPIRES_AT_SEC, remainingSec: 100 });
  });

  it("reports_zero_remaining_seconds_when_now_equals_the_expiry", () => {
    // Given: 組み立て済みの storageState JSON
    const json = JSON.stringify(
      buildAccessSessionStorageState(ORIGIN, createFakeAccessJwt(EXPIRES_AT_SEC)),
    );

    // When: 期限ちょうどの時刻で確認する
    const status = checkAccessSession(json, EXPIRES_AT_SEC);

    // Then: 残りは 0 秒
    expect(status.remainingSec).toBe(0);
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
