import { describe, expect, it } from "vitest";
import { classifyHttpStatus } from "./http-error.ts";

describe("classifyHttpStatus", () => {
  it("200 の時 success を返す", () => {
    // Given: 宣言表にある成功 status
    const status = 200;

    // When: 分類する
    const got = classifyHttpStatus(status);

    // Then: success
    expect(got).toEqual({ kind: "success" });
  });

  it("404 の時 episode_not_found を返し 400 に畳まない", () => {
    // Given: 宣言表にある 404
    const status = 404;

    // When: 分類する
    const got = classifyHttpStatus(status);

    // Then: 404 専用 code
    expect(got).toEqual({ kind: "error", code: "episode_not_found" });
  });

  it("400 の時 validation_error を返す", () => {
    // Given: 宣言表にある 400
    const status = 400;

    // When: 分類する
    const got = classifyHttpStatus(status);

    // Then: validation_error
    expect(got).toEqual({ kind: "error", code: "validation_error" });
  });

  it("503 の時 unavailable を返す", () => {
    // Given: 宣言表にある 503
    const status = 503;

    // When: 分類する
    const got = classifyHttpStatus(status);

    // Then: unavailable
    expect(got).toEqual({ kind: "error", code: "unavailable" });
  });

  it("500 の時 configuration_error を返す", () => {
    // Given: runtime config 不備を表す 500
    const status = 500;

    // When: 分類する
    const got = classifyHttpStatus(status);

    // Then: configuration_error
    expect(got).toEqual({ kind: "error", code: "configuration_error" });
  });

  it("表に無い 2xx の時 success を返す", () => {
    // Given: 宣言表に無い 201
    const status = 201;

    // When: 分類する
    const got = classifyHttpStatus(status);

    // Then: 級 2 は success
    expect(got).toEqual({ kind: "success" });
  });

  it("表に無い 4xx の時 client_error を返す", () => {
    // Given: 宣言表に無い 418
    const status = 418;

    // When: 分類する
    const got = classifyHttpStatus(status);

    // Then: 級 4 は client_error
    expect(got).toEqual({ kind: "client_error" });
  });

  it("表に無い 5xx の時 unavailable を返す", () => {
    // Given: 宣言表に無い 502
    const status = 502;

    // When: 分類する
    const got = classifyHttpStatus(status);

    // Then: 級 5 は unavailable
    expect(got).toEqual({ kind: "error", code: "unavailable" });
  });

  it("表に無い 3xx の時 unavailable を返す", () => {
    // Given: 宣言表に無い 301
    const status = 301;

    // When: 分類する
    const got = classifyHttpStatus(status);

    // Then: 級 3 は unavailable
    expect(got).toEqual({ kind: "error", code: "unavailable" });
  });

  it("1xx の時 unavailable を返す", () => {
    // Given: 契約外の 101
    const status = 101;

    // When: 分類する
    const got = classifyHttpStatus(status);

    // Then: 級 1 は unavailable
    expect(got).toEqual({ kind: "error", code: "unavailable" });
  });

  it("600 の時 unavailable を返す", () => {
    // Given: 契約外の 600
    const status = 600;

    // When: 分類する
    const got = classifyHttpStatus(status);

    // Then: 級 6 は unavailable
    expect(got).toEqual({ kind: "error", code: "unavailable" });
  });
});
