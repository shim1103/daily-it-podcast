import { describe, expect, it } from "vitest";
import { ProgressNotFoundError } from "./progress-not-found-error.ts";
import { ProgressRuleError } from "./progress-rule-error.ts";
import { ProgressWriteConflictError } from "./progress-write-conflict-error.ts";

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

describe("ProgressWriteConflictError", () => {
  it("keeps_name_message_and_cause", () => {
    const cause = new Error("条件付き書込が不成立");

    const got = new ProgressWriteConflictError("競合が最大試行回数まで続いた: ep-1", { cause });

    expect(got).toBeInstanceOf(Error);
    expect(got.name).toBe("ProgressWriteConflictError");
    expect(got.message).toBe("競合が最大試行回数まで続いた: ep-1");
    expect(got.cause).toBe(cause);
  });
});
