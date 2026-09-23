import { describe, expect, it } from "vitest";
import { ProgressPullResponseSchema, UnavailableError } from "../../../contracts/index.ts";
import { R2Error } from "../infrastructure/r2/r2-error.ts";
import { createFakePullProgressUseCase, validProgressPullResponse } from "./fake-use-cases.ts";
import { createPullProgressController } from "./pull-progress-controller.ts";

describe("createPullProgressController", () => {
  it("UseCase が成功する時、ProgressPullResponse schema を満たす", async () => {
    // Given: 契約どおりの pull 応答を返す Fake UseCase
    const useCase = createFakePullProgressUseCase();
    const controller = createPullProgressController(useCase);

    // When: 検証済み since を渡す
    const got = await controller("2026-09-22T10:00:00.000Z");

    // Then: 契約 schema を満たす
    expect(ProgressPullResponseSchema.safeParse(got).success).toBe(true);
    expect(got).toEqual(validProgressPullResponse);
  });

  it("UseCase が Infrastructure Error を throw する時、UnavailableError に cause 付きで変換する", async () => {
    // Given: Infrastructure 失敗を throw する Fake UseCase
    const infraError = new R2Error("storage 読取に失敗");
    const useCase = createFakePullProgressUseCase(async () => {
      throw infraError;
    });
    const controller = createPullProgressController(useCase);

    // When: 呼ぶ
    const act = controller("2026-09-22T10:00:00.000Z");

    // Then: External UnavailableError が Infrastructure を cause に持つ
    await expect(act).rejects.toSatisfy(
      (error: unknown) => error instanceof UnavailableError && error.cause === infraError,
    );
  });
});
