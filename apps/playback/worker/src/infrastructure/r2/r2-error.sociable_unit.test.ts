import { describe, expect, it } from "vitest";
import { R2Error } from "./r2-error.ts";

describe("R2Error", () => {
  it("message と name を持ち Error を継承する", () => {
    // Given: 診断用 message
    const message = "R2 読取に失敗";

    // When: Infrastructure Error を生成する
    const got = new R2Error(message);

    // Then: R2 起因の Internal Error
    expect(got).toBeInstanceOf(Error);
    expect(got).toBeInstanceOf(R2Error);
    expect(got.name).toBe("R2Error");
    expect(got.message).toBe(message);
  });

  it("cause を保持する", () => {
    // Given: 外部 SDK の元 Error
    const cause = new Error("network");

    // When: cause 付きで生成する
    const got = new R2Error("R2 読取に失敗", { cause });

    // Then: cause chain が残る
    expect(got.cause).toBe(cause);
  });
});
