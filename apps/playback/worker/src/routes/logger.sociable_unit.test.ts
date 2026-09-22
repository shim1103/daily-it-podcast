import { afterEach, describe, expect, it, vi } from "vitest";
import { logError, logInfo } from "./logger.ts";

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
    logInfo({ event: "request_start", requestId: "req-1", method: "GET" });

    // Then: console.log がそのまま渡す
    expect(logSpy).toHaveBeenCalledWith({
      event: "request_start",
      requestId: "req-1",
      method: "GET",
    });
  });
});

describe("logError", () => {
  it("console.error を structured payload で呼ぶ", () => {
    // Given: name/message を持つ error payload
    // When: logError を呼ぶ
    logError({ name: "Error", message: "boom", requestId: "req-1" });

    // Then: console.error がそのまま渡す
    expect(errorSpy).toHaveBeenCalledWith({ name: "Error", message: "boom", requestId: "req-1" });
  });
});
