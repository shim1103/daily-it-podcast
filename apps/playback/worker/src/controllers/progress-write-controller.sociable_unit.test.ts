import { describe, expect, it } from "vitest";
import {
  ProgressWriteResponseSchema,
  UnavailableError,
  ValidationError,
} from "../../../contracts/index.ts";
import { ProgressRuleError } from "../entities/errors/progress-rule-error.ts";
import { R2Error } from "../infrastructure/r2/r2-error.ts";
import { createFakeProgressWriteUseCase, validProgressWriteResponse } from "./fake-use-cases.ts";
import { createProgressWriteController } from "./progress-write-controller.ts";

describe("createProgressWriteController", () => {
  it("UseCase が成功する時、ProgressWriteResponse schema を満たす", async () => {
    // Given: 契約どおりの Write 応答を返す Fake UseCase
    const useCase = createFakeProgressWriteUseCase();
    const controller = createProgressWriteController(useCase);

    // When: 検証済み episodeId と body を渡す
    const got = await controller("ep-1", {
      positionSec: 12,
      clientAt: "2026-09-22T10:00:00.000Z",
    });

    // Then: 契約 schema を満たす
    expect(ProgressWriteResponseSchema.safeParse(got).success).toBe(true);
    expect(got).toEqual(validProgressWriteResponse);
  });

  it("UseCase が ProgressRuleError を throw する時、ValidationError に cause 付きで変換する", async () => {
    // Given: Domain 規則違反を throw する Fake UseCase
    const domainError = new ProgressRuleError("clientAt が許容 skew を超える");
    const useCase = createFakeProgressWriteUseCase(async () => {
      throw domainError;
    });
    const controller = createProgressWriteController(useCase);

    // When: 呼ぶ
    const act = controller("ep-1", {
      positionSec: 12,
      clientAt: "2026-09-22T10:00:00.000Z",
    });

    // Then: External ValidationError が Domain を cause に持つ
    await expect(act).rejects.toSatisfy(
      (error: unknown) => error instanceof ValidationError && error.cause === domainError,
    );
  });

  it("UseCase が R2Error を throw する時、UnavailableError に cause 付きで変換する", async () => {
    // Given: Infrastructure 失敗を throw する Fake UseCase（D1 失敗も同写像）
    const infraError = new R2Error("storage 読取に失敗");
    const useCase = createFakeProgressWriteUseCase(async () => {
      throw infraError;
    });
    const controller = createProgressWriteController(useCase);

    // When: 呼ぶ
    const act = controller("ep-1", {
      positionSec: 12,
      clientAt: "2026-09-22T10:00:00.000Z",
    });

    // Then: External UnavailableError が Infrastructure を cause に持つ
    await expect(act).rejects.toSatisfy(
      (error: unknown) => error instanceof UnavailableError && error.cause === infraError,
    );
  });
});
