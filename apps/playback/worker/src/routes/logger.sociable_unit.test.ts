import { afterEach, describe, expect, it, vi } from "vitest";
import { logError, logInfo } from "./logger.ts";
import type { RequestId } from "./request-context.ts";

const testRequestId = "req-1" as RequestId;

const logSpy = vi.spyOn(console, "log").mockImplementation(() => {});
const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

afterEach(() => {
  logSpy.mockClear();
  errorSpy.mockClear();
});

describe("logInfo", () => {
  it("console.log を structured payload で呼ぶ", () => {
    // Given: event と任意 field を持つ payload
    // When: logInfo を呼ぶ
    logInfo({ event: "request_start", requestId: testRequestId, method: "GET" });

    // Then: console.log がそのまま渡す
    expect(logSpy).toHaveBeenCalledWith({
      event: "request_start",
      requestId: testRequestId,
      method: "GET",
    });
  });

  it("requestId が無い payload も受け付ける", () => {
    // Given: HTTP request に紐づかない payload（application 層の集約結果等）
    // When: logInfo を呼ぶ
    logInfo({ event: "episode_skipped", stem: "bad" });

    // Then: console.log がそのまま渡す
    expect(logSpy).toHaveBeenCalledWith({ event: "episode_skipped", stem: "bad" });
  });
});

describe("logError", () => {
  it("console.error を structured payload で呼ぶ", () => {
    // Given: name/message を持つ error payload
    // When: logError を呼ぶ
    logError({ event: "request_error", requestId: testRequestId, name: "Error", message: "boom" });

    // Then: console.error がそのまま渡す
    expect(errorSpy).toHaveBeenCalledWith({
      event: "request_error",
      requestId: testRequestId,
      name: "Error",
      message: "boom",
    });
  });
});
