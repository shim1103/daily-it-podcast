import { describe, expect, it } from "vitest";
import { ProgressNotFoundError } from "./progress-not-found-error.ts";
import { ProgressRuleError } from "./progress-rule-error.ts";

describe("ProgressRuleError", () => {
  it("name を ProgressRuleError にし message を保持する", () => {
    const got = new ProgressRuleError("clientAt が許容 skew を超える");

    expect(got).toBeInstanceOf(Error);
    expect(got.name).toBe("ProgressRuleError");
    expect(got.message).toBe("clientAt が許容 skew を超える");
  });
});

describe("ProgressNotFoundError", () => {
  it("name を ProgressNotFoundError にし message を保持する", () => {
    const got = new ProgressNotFoundError("進捗行が無い: ep-1");

    expect(got).toBeInstanceOf(Error);
    expect(got.name).toBe("ProgressNotFoundError");
    expect(got.message).toBe("進捗行が無い: ep-1");
  });
});
