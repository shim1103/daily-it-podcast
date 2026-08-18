import { describe, expect, it } from "vitest";
import {
  NotFoundError,
  UnavailableError,
  ValidationError,
} from "./external-errors.ts";

describe("ValidationError", () => {
  it("message と name を持ち Error を継承する", () => {
    // Given: 診断用 message
    const message = "入力が契約に不適合";

    // When: External Error を生成する
    const got = new ValidationError(message);

    // Then: 契約どおりの class
    expect(got).toBeInstanceOf(Error);
    expect(got).toBeInstanceOf(ValidationError);
    expect(got.name).toBe("ValidationError");
    expect(got.message).toBe(message);
  });

  it("cause を保持する", () => {
    // Given: schema 失敗を表す元 Error
    const cause = new Error("zod");

    // When: cause 付きで生成する
    const got = new ValidationError("入力が契約に不適合", { cause });

    // Then: cause chain が残る
    expect(got.cause).toBe(cause);
  });
});

describe("NotFoundError", () => {
  it("message と name を持ち Error を継承する", () => {
    // Given: 診断用 message
    const message = "エピソードが無い";

    // When: External Error を生成する
    const got = new NotFoundError(message);

    // Then: 契約どおりの class
    expect(got).toBeInstanceOf(Error);
    expect(got).toBeInstanceOf(NotFoundError);
    expect(got.name).toBe("NotFoundError");
    expect(got.message).toBe(message);
  });
});

describe("UnavailableError", () => {
  it("message と name を持ち Error を継承する", () => {
    // Given: 診断用 message
    const message = "利用できない";

    // When: External Error を生成する
    const got = new UnavailableError(message);

    // Then: 契約どおりの class
    expect(got).toBeInstanceOf(Error);
    expect(got).toBeInstanceOf(UnavailableError);
    expect(got.name).toBe("UnavailableError");
    expect(got.message).toBe(message);
  });
});
