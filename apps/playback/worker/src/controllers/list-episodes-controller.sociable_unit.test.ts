import { describe, expect, it } from "vitest";
import { ListEpisodesResponseSchema, UnavailableError } from "../../../contracts/index.ts";
import { R2Error } from "../infrastructure/r2/r2-error.ts";
import { createListEpisodesController } from "./list-episodes-controller.ts";
import { createFakeListEpisodesUseCase, validListEpisodesResponse } from "./fake-use-cases.ts";

describe("createListEpisodesController", () => {
  it("UseCase が成功する時、ListEpisodesResponse schema を満たす", async () => {
    // Given: 契約どおりの一覧を返す Fake UseCase
    const useCase = createFakeListEpisodesUseCase();
    const controller = createListEpisodesController(useCase);

    // When: unknown 入力で呼ぶ
    const got = await controller({});

    // Then: 契約 schema を満たす
    expect(ListEpisodesResponseSchema.safeParse(got).success).toBe(true);
    expect(got).toEqual(validListEpisodesResponse);
  });

  it("UseCase が R2Error を throw する時、UnavailableError に cause 付きで変換する", async () => {
    // Given: Infrastructure 失敗を throw する Fake UseCase
    const r2Error = new R2Error("R2 読取に失敗");
    const useCase = createFakeListEpisodesUseCase(async () => {
      throw r2Error;
    });
    const controller = createListEpisodesController(useCase);

    // When: 一覧を取得する
    const act = controller({});

    // Then: External UnavailableError が Infrastructure を cause に持つ
    await expect(act).rejects.toSatisfy(
      (error: unknown) => error instanceof UnavailableError && error.cause === r2Error,
    );
  });
});
