import { describe, expect, it } from "vitest";
import { UnavailableError, ValidationError } from "../../../contracts/index.ts";
import type {
  ProgressWriteUseCaseInput,
  ProgressWriteUseCaseOutput,
} from "../application/progress/progress-write-use-case-io.ts";
import { ProgressRuleError } from "../entities/errors/progress-rule-error.ts";
import { R2Error } from "../infrastructure/r2/r2-error.ts";
import { createProgressWriteController } from "./progress-write-controller.ts";

const CLIENT_AT = "2026-09-22T10:00:00.000Z";
const FIRST_PLAYED_AT = "2026-09-21T09:00:00.000Z";
const FIRST_COMPLETED_AT = "2026-09-21T09:30:00.000Z";

describe("createProgressWriteController", () => {
  it("passes_episode_id_and_every_body_field_to_use_case_input_when_called", async () => {
    // Given: 受け取った Input を記録する use case の stub
    const received: ProgressWriteUseCaseInput[] = [];
    const controller = createProgressWriteController(async (input) => {
      received.push(input);
      return { firstPlayedAt: FIRST_PLAYED_AT, firstCompletedAt: null };
    });

    // When: 検証済みの episodeId と body を渡す
    await controller("ep-1", { positionSec: 12.5, clientAt: CLIENT_AT });

    // Then: use case が episodeId・positionSec・clientAt だけを持つ Input を 1 回受け取る
    expect(received).toStrictEqual([{ episodeId: "ep-1", positionSec: 12.5, clientAt: CLIENT_AT }]);
  });

  it("returns_response_with_first_played_and_first_completed_when_use_case_output_is_completed", async () => {
    // Given: 完走済みの Output を返す use case の stub
    const output: ProgressWriteUseCaseOutput = {
      firstPlayedAt: FIRST_PLAYED_AT,
      firstCompletedAt: FIRST_COMPLETED_AT,
    };
    const controller = createProgressWriteController(async () => output);

    // When: 呼ぶ
    const got = await controller("ep-1", { positionSec: 12, clientAt: CLIENT_AT });

    // Then: Output の 2 field が Response の同名 field へ写る
    expect(got).toStrictEqual({
      firstPlayedAt: FIRST_PLAYED_AT,
      firstCompletedAt: FIRST_COMPLETED_AT,
    });
  });

  it("returns_response_with_null_first_completed_when_use_case_output_is_not_completed", async () => {
    // Given: 未完走の Output を返す use case の stub
    const controller = createProgressWriteController(async () => ({
      firstPlayedAt: FIRST_PLAYED_AT,
      firstCompletedAt: null,
    }));

    // When: 呼ぶ
    const got = await controller("ep-1", { positionSec: 12, clientAt: CLIENT_AT });

    // Then: firstCompletedAt は null のまま写る
    expect(got).toStrictEqual({ firstPlayedAt: FIRST_PLAYED_AT, firstCompletedAt: null });
  });

  it("omits_fields_outside_response_contract_when_use_case_output_has_extra_fields", async () => {
    // Given: 契約に無い field（positionSec）を余分に持つ Output を返す use case の stub
    const outputWithExtra = {
      firstPlayedAt: FIRST_PLAYED_AT,
      firstCompletedAt: null,
      positionSec: 99,
    };
    const controller = createProgressWriteController(async () => outputWithExtra);

    // When: 呼ぶ
    const got = await controller("ep-1", { positionSec: 12, clientAt: CLIENT_AT });

    // Then: Response は契約の 2 field だけを持つ
    expect(got).toStrictEqual({ firstPlayedAt: FIRST_PLAYED_AT, firstCompletedAt: null });
  });

  it("throws_validation_error_with_cause_when_use_case_throws_progress_rule_error", async () => {
    // Given: Domain 規則違反を throw する use case の stub
    const domainError = new ProgressRuleError("clientAt が許容 skew を超える");
    const controller = createProgressWriteController(async () => {
      throw domainError;
    });

    // When: 呼ぶ
    const act = controller("ep-1", { positionSec: 12, clientAt: CLIENT_AT });

    // Then: External ValidationError が Domain を cause に持つ
    await expect(act).rejects.toSatisfy(
      (error: unknown) => error instanceof ValidationError && error.cause === domainError,
    );
  });

  it("throws_unavailable_error_with_cause_when_use_case_throws_infrastructure_error", async () => {
    // Given: Infrastructure 失敗を throw する use case の stub（D1 失敗も同じ写像）
    const infraError = new R2Error("storage 読取に失敗");
    const controller = createProgressWriteController(async () => {
      throw infraError;
    });

    // When: 呼ぶ
    const act = controller("ep-1", { positionSec: 12, clientAt: CLIENT_AT });

    // Then: External UnavailableError が Infrastructure を cause に持つ
    await expect(act).rejects.toSatisfy(
      (error: unknown) => error instanceof UnavailableError && error.cause === infraError,
    );
  });
});
