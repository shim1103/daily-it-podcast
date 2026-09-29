import { afterEach, describe, expect, it, vi } from "vitest";
import { logError, logInfo, type RequestId } from "./logger.ts";

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
