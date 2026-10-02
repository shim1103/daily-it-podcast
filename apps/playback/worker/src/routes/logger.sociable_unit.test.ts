import { afterEach, describe, expect, it, vi } from "vitest";
import { logError, logInfo, type RequestId, writeErrorLine, writeLine } from "./logger.ts";

const testRequestId = "req-1" as RequestId;

const logSpy = vi.spyOn(console, "log").mockImplementation(() => {});
const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

afterEach(() => {
  logSpy.mockClear();
  errorSpy.mockClear();
});

describe("logInfo", () => {
  it("console.log を structured payload で呼ぶ", () => {
    // Given: event・requestId と任意 field を持つ payload
    // When: logInfo を呼ぶ
    logInfo({ event: "request_start", requestId: testRequestId, method: "GET" });

    // Then: console.log がそのまま渡す
    expect(logSpy).toHaveBeenCalledWith({
      event: "request_start",
      requestId: testRequestId,
      method: "GET",
    });
  });
});

describe("writeLine", () => {
  it("passes_the_line_unchanged_to_console_log", () => {
    // Given: request に紐づかない 1 行の文字列（CLI の出力など）
    // When: writeLine を呼ぶ
    writeLine("標準出力へ出す行");

    // Then: console.log がその 1 行だけを受け取る
    expect(logSpy).toHaveBeenCalledTimes(1);
    expect(logSpy).toHaveBeenCalledWith("標準出力へ出す行");
  });
});

describe("writeErrorLine", () => {
  it("passes_the_line_unchanged_to_console_error", () => {
    // Given: request に紐づかない 1 行の文字列（CLI のエラー出力など）
    // When: writeErrorLine を呼ぶ
    writeErrorLine("標準エラーへ出す行");

    // Then: console.error がその 1 行だけを受け取る
    expect(errorSpy).toHaveBeenCalledTimes(1);
    expect(errorSpy).toHaveBeenCalledWith("標準エラーへ出す行");
  });
});

describe("logError", () => {
  it("console.error を structured payload で呼ぶ", () => {
    // Given: name/message/requestId を持つ error payload
    // When: logError を呼ぶ
    logError({ name: "Error", message: "boom", requestId: testRequestId });

    // Then: console.error がそのまま渡す
    expect(errorSpy).toHaveBeenCalledWith({
      name: "Error",
      message: "boom",
      requestId: testRequestId,
    });
  });
});
