import { describe, expect, it } from "vitest";
import {
  GetEpisodeResponseSchema,
  NotFoundError,
  UnavailableError,
  ValidationError,
} from "../../../contracts/index.ts";
import { EpisodeContentError } from "../entities/errors/episode-content-error.ts";
import { DriveError } from "../infrastructure/drive/drive-error.ts";
import { createGetEpisodeController } from "./get-episode-controller.ts";
import { createFakeGetEpisodeUseCase, validGetEpisodeResponse } from "./fake-use-cases.ts";

describe("createGetEpisodeController", () => {
  it("UseCase が成功する時、GetEpisodeResponse schema を満たす", async () => {
    // Given: 契約どおりの 1件を返す Fake UseCase
    const useCase = createFakeGetEpisodeUseCase();
    const controller = createGetEpisodeController(useCase);

    // When: 有効な episodeId を unknown として渡す
    const got = await controller({ episodeId: "ep-1" });

    // Then: 契約 schema を満たし audioRef がある
    expect(GetEpisodeResponseSchema.safeParse(got).success).toBe(true);
    expect(got.audioRef).toBe(validGetEpisodeResponse.audioRef);
  });

  it("episodeId が空の時、ValidationError を throw する", async () => {
    // Given: 空 episodeId
    const useCase = createFakeGetEpisodeUseCase();
    const controller = createGetEpisodeController(useCase);

    // When: schema が拒否する入力を渡す
    const act = controller({ episodeId: "" });

    // Then: External ValidationError
    await expect(act).rejects.toBeInstanceOf(ValidationError);
  });

  it("UseCase が EpisodeContentError を throw する時、NotFoundError に cause 付きで変換する", async () => {
    // Given: Domain 不在を throw する Fake UseCase
    const domainError = new EpisodeContentError("JSON エントリが無い: ep-1");
    const useCase = createFakeGetEpisodeUseCase(async () => {
      throw domainError;
    });
    const controller = createGetEpisodeController(useCase);

    // When: 有効な episodeId で呼ぶ
    const act = controller({ episodeId: "ep-1" });

    // Then: External NotFoundError が Domain を cause に持つ
    await expect(act).rejects.toSatisfy(
      (error: unknown) => error instanceof NotFoundError && error.cause === domainError,
    );
  });

  it("default Fake UseCase が未知の episodeId の時、NotFoundError に変換する", async () => {
    // Given: JSON に無い episodeId を default Fake に渡す
    const useCase = createFakeGetEpisodeUseCase();
    const controller = createGetEpisodeController(useCase);

    // When: 未知 id で呼ぶ
    const act = controller({ episodeId: "ep-missing" });

    // Then: Fake の欠落経路が Domain → External NotFound へ写る
    await expect(act).rejects.toSatisfy(
      (error: unknown) =>
        error instanceof NotFoundError && error.cause instanceof EpisodeContentError,
    );
  });

  it("UseCase が DriveError を throw する時、UnavailableError に cause 付きで変換する", async () => {
    // Given: Infrastructure 失敗を throw する Fake UseCase
    const driveError = new DriveError("Drive 読取に失敗");
    const useCase = createFakeGetEpisodeUseCase(async () => {
      throw driveError;
    });
    const controller = createGetEpisodeController(useCase);

    // When: 有効な episodeId で呼ぶ
    const act = controller({ episodeId: "ep-1" });

    // Then: External UnavailableError が Infrastructure を cause に持つ
    await expect(act).rejects.toSatisfy(
      (error: unknown) => error instanceof UnavailableError && error.cause === driveError,
    );
  });
});
