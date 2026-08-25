import { afterEach, describe, expect, it, vi } from "vitest";
import {
  ConfigurationError,
  NotFoundError,
  UnavailableError,
  ValidationError,
} from "../../../contracts/index.ts";
import { createHttpErrorResponse } from "./http-error-response.ts";

const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

afterEach(() => {
  errorSpy.mockClear();
});

describe("createHttpErrorResponse", () => {
  it("ValidationError の時、400 と validation_error を返す", async () => {
    // Given: ValidationError
    const error = new ValidationError("入力が契約に不適合");

    // When: HTTP Error Response を作る
    const got = createHttpErrorResponse(error, "req-1");

    // Then: 400 と契約 code のみ
    expect(got.status).toBe(400);
    expect(await got.json()).toEqual({ code: "validation_error" });
  });

  it("NotFoundError の時、404 と episode_not_found を返す", async () => {
    // Given: NotFoundError
    const error = new NotFoundError("エピソードが無い");

    // When: HTTP Error Response を作る
    const got = createHttpErrorResponse(error, "req-1");

    // Then: 404 と契約 code のみ
    expect(got.status).toBe(404);
    expect(await got.json()).toEqual({ code: "episode_not_found" });
  });

  it("ConfigurationError の時、500 と configuration_error を返す", async () => {
    // Given: ConfigurationError
    const error = new ConfigurationError("設定を確認できません");

    // When: HTTP Error Response を作る
    const got = createHttpErrorResponse(error, "req-1");

    // Then: 500 と契約 code のみ
    expect(got.status).toBe(500);
    expect(await got.json()).toEqual({ code: "configuration_error" });
  });

  it("UnavailableError の時、503 と unavailable を返す", async () => {
    // Given: UnavailableError
    const error = new UnavailableError("利用できない");

    // When: HTTP Error Response を作る
    const got = createHttpErrorResponse(error, "req-1");

    // Then: 503 と契約 code のみ
    expect(got.status).toBe(503);
    expect(await got.json()).toEqual({ code: "unavailable" });
  });

  it("mapping に無い名前の Error の時、500 と empty body を返す", async () => {
    // Given: mapping 対象外の Error
    const error = new Error("予期しない失敗");

    // When: HTTP Error Response を作る
    const got = createHttpErrorResponse(error, "req-1");

    // Then: 500 と empty body
    expect(got.status).toBe(500);
    expect(await got.text()).toBe("");
  });

  it("Error でない値の時、500 と empty body を返し message へ String 化して log する", async () => {
    // Given: Error ではない throw 値
    const thrown = "文字列で throw された失敗";

    // When: HTTP Error Response を作る
    const got = createHttpErrorResponse(thrown, "req-1");

    // Then: 500 を返し、UnmappedError として String 化した message を log する
    expect(got.status).toBe(500);
    expect(await got.text()).toBe("");
    expect(errorSpy).toHaveBeenCalledWith({
      name: "UnmappedError",
      message: thrown,
      requestId: "req-1",
    });
  });

  it("cause が 2 段以上 Error で連なる時、cause chain 全体を log payload に含める", () => {
    // Given: cause が 2 段連なる mapping 対象外の Error
    const rootCause = new Error("root 失敗");
    const middleCause = new Error("middle 失敗", { cause: rootCause });
    const error = new Error("予期しない失敗", { cause: middleCause });

    // When: HTTP Error Response を作る
    createHttpErrorResponse(error, "req-1");

    // Then: cause chain 全体（root まで）が log payload に含まれる
    expect(errorSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "UnmappedError",
        message: error.message,
        cause: expect.objectContaining({
          name: "Error",
          message: middleCause.message,
          cause: expect.objectContaining({
            name: "Error",
            message: rootCause.message,
          }),
        }),
      }),
    );
  });
});
