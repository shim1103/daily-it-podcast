import { describe, expect, it } from "vitest";
import { DriveError } from "./drive-error.ts";

describe("DriveError", () => {
  it("message と name を持ち Error を継承する", () => {
    // Given: 診断用 message
    const message = "Drive 読取に失敗";

    // When: Infrastructure Error を生成する
    const got = new DriveError(message);

    // Then: Drive 起因の Internal Error
    expect(got).toBeInstanceOf(Error);
    expect(got).toBeInstanceOf(DriveError);
    expect(got.name).toBe("DriveError");
    expect(got.message).toBe(message);
  });

  it("cause を保持する", () => {
    // Given: 外部 SDK の元 Error
    const cause = new Error("network");

    // When: cause 付きで生成する
    const got = new DriveError("Drive 読取に失敗", { cause });

    // Then: cause chain が残る
    expect(got.cause).toBe(cause);
  });
});
