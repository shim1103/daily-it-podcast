import { describe, expect, it } from "vitest";
import { utc } from "../../../../test/support/utc-instant.ts";
import {
  PROGRESS_CLIENT_AT_MAX_FUTURE_SKEW_SEC,
  PROGRESS_CLIENT_AT_MAX_PAST_SKEW_SEC,
} from "../../entities/constants/progress.ts";
import { ProgressRuleError } from "../../entities/errors/progress-rule-error.ts";
import { assertClientAtWithinSkew } from "./assert-client-at-within-skew.ts";

/**
 * scope: Sociable Unit
 * real: assertClientAtWithinSkew（clientAt と serverAt の skew 境界）
 * double: なし
 */
const SERVER_AT = new Date("2026-09-22T12:00:00.000Z");
const FUTURE_SKEW_MS = PROGRESS_CLIENT_AT_MAX_FUTURE_SKEW_SEC * 1000;
const PAST_SKEW_MS = PROGRESS_CLIENT_AT_MAX_PAST_SKEW_SEC * 1000;

describe("assertClientAtWithinSkew", () => {
  it("allows_clientAt_exactly_at_future_skew_limit", () => {
    // Given: serverAt から未来側ちょうど許容幅
    const clientAt = utc(new Date(SERVER_AT.getTime() + FUTURE_SKEW_MS).toISOString());

    // When / Then: ちょうどは拒否しない
    expect(() => assertClientAtWithinSkew(clientAt, SERVER_AT)).not.toThrow();
  });

  it("rejects_clientAt_one_ms_beyond_future_skew_limit", () => {
    // Given: 未来側上限の 1ms 超過
    const clientAt = utc(new Date(SERVER_AT.getTime() + FUTURE_SKEW_MS + 1).toISOString());

    // When / Then: ProgressRuleError
    expect(() => assertClientAtWithinSkew(clientAt, SERVER_AT)).toThrow(ProgressRuleError);
  });

  it("allows_clientAt_exactly_at_past_skew_limit", () => {
    // Given: serverAt から過去側ちょうど許容幅
    const clientAt = utc(new Date(SERVER_AT.getTime() - PAST_SKEW_MS).toISOString());

    // When / Then: ちょうどは拒否しない
    expect(() => assertClientAtWithinSkew(clientAt, SERVER_AT)).not.toThrow();
  });

  it("rejects_clientAt_one_ms_beyond_past_skew_limit", () => {
    // Given: 過去側上限の 1ms 超過
    const clientAt = utc(new Date(SERVER_AT.getTime() - PAST_SKEW_MS - 1).toISOString());

    // When / Then: ProgressRuleError
    expect(() => assertClientAtWithinSkew(clientAt, SERVER_AT)).toThrow(ProgressRuleError);
  });
});
