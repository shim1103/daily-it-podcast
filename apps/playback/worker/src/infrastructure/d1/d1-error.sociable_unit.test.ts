import { describe, expect, it } from "vitest";
import { D1Error } from "./d1-error.ts";

describe("D1Error", () => {
  it("has_name_and_message_and_extends_error", () => {
    // Given: 診断用 message
    const message = "D1 進捗の読取に失敗";

    // When: Infrastructure Error を生成する
    const got = new D1Error(message);

    // Then: D1 起因の Internal Error
    expect(got).toBeInstanceOf(Error);
    expect(got).toBeInstanceOf(D1Error);
    expect(got.name).toBe("D1Error");
    expect(got.message).toBe(message);
  });

  it("keeps_cause_when_created_with_cause", () => {
    // Given: D1 binding の元 Error
    const cause = new Error("D1_ERROR");

    // When: cause 付きで生成する
    const got = new D1Error("D1 進捗の読取に失敗", { cause });

    // Then: cause chain が残る
    expect(got.cause).toBe(cause);
  });
});
