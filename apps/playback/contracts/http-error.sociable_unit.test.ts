import { describe, expect, it } from "vitest";
import { mapHttpStatusToError } from "./http-error.ts";

describe("mapHttpStatusToError", () => {
  it("404 の時 episode_not_found を返し 400 に畳まない", () => {
    // Given: 宣言表にある 404
    const status = 404;

    // When: 分類する
    const got = mapHttpStatusToError(status);

    // Then: 404 専用 code
    expect(got).toBe("episode_not_found");
  });

  it("400 の時 validation_error を返す", () => {
    // Given: 宣言表にある 400
    const status = 400;

    // When: 分類する
    const got = mapHttpStatusToError(status);

    // Then: validation_error
    expect(got).toBe("validation_error");
  });

  it("503 の時 unavailable を返す", () => {
    // Given: 宣言表にある 503
    const status = 503;

    // When: 分類する
    const got = mapHttpStatusToError(status);

    // Then: unavailable
    expect(got).toBe("unavailable");
  });

  it("500 の時 configuration_error を返す", () => {
    // Given: runtime config 不備を表す 500
    const status = 500;

    // When: 分類する
    const got = mapHttpStatusToError(status);

    // Then: configuration_error
    expect(got).toBe("configuration_error");
  });

  it("表に無い 4xx の時 undefined を返す", () => {
    // Given: 宣言表に無い 418
    const status = 418;

    // When: 分類する
    const got = mapHttpStatusToError(status);

    // Then: 契約に対応する code は無い
    expect(got).toBeUndefined();
  });

  it("表に無い 5xx の時 unavailable を返す", () => {
    // Given: 宣言表に無い 502
    const status = 502;

    // When: 分類する
    const got = mapHttpStatusToError(status);

    // Then: 級 5 は unavailable
    expect(got).toBeUndefined();
  });

  it("表に無い 3xx の時 unavailable を返す", () => {
    // Given: 宣言表に無い 301
    const status = 301;

    // When: 分類する
    const got = mapHttpStatusToError(status);

    // Then: 級 3 は unavailable
    expect(got).toBeUndefined();
  });

  it("1xx の時 unavailable を返す", () => {
    // Given: 契約外の 101
    const status = 101;

    // When: 分類する
    const got = mapHttpStatusToError(status);

    // Then: 級 1 は unavailable
    expect(got).toBeUndefined();
  });

  it("600 の時 unavailable を返す", () => {
    // Given: 契約外の 600
    const status = 600;

    // When: 分類する
    const got = mapHttpStatusToError(status);

    // Then: 級 6 は unavailable
    expect(got).toBeUndefined();
  });
});
