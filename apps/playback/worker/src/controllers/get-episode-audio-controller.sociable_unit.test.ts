import { describe, expect, it } from "vitest";
import { NotFoundError, UnavailableError, ValidationError } from "../../../contracts/index.ts";
import { EpisodeNotFoundError } from "../entities/errors/episode-not-found-error.ts";
import { DriveError } from "../infrastructure/drive/drive-error.ts";
import { createGetEpisodeAudioController } from "./get-episode-audio-controller.ts";
import { validAudioBytes } from "../test/fixtures/audio-bytes.ts";
import { createFakeGetEpisodeAudioUseCase } from "./fake-use-cases.ts";

describe("createGetEpisodeAudioController", () => {
  it("UseCase が成功する時、音声 byte を返す", async () => {
    // Given: wav byte を返す Fake UseCase
    const useCase = createFakeGetEpisodeAudioUseCase();
    const controller = createGetEpisodeAudioController(useCase);

    // When: 有効な episodeId を unknown として渡す
    const got = await controller({ episodeId: "ep-1" });

    // Then: Fake が返した byte と一致する
    expect(got).toEqual(validAudioBytes);
  });

  it("episodeId が空の時、ValidationError を throw する", async () => {
    // Given: 空 episodeId
    const useCase = createFakeGetEpisodeAudioUseCase();
    const controller = createGetEpisodeAudioController(useCase);

    // When: schema が拒否する入力を渡す
    const act = controller({ episodeId: "" });

    // Then: External ValidationError
    await expect(act).rejects.toBeInstanceOf(ValidationError);
  });

  it("UseCase が EpisodeNotFoundError を throw する時、NotFoundError に cause 付きで変換する", async () => {
    // Given: Domain 不在を throw する Fake UseCase
    const domainError = new EpisodeNotFoundError("音声エントリが無い: ep-1");
    const useCase = createFakeGetEpisodeAudioUseCase(async () => {
      throw domainError;
    });
    const controller = createGetEpisodeAudioController(useCase);

    // When: 有効な episodeId で呼ぶ
    const act = controller({ episodeId: "ep-1" });

    // Then: External NotFoundError が Domain を cause に持つ
    await expect(act).rejects.toSatisfy(
      (error: unknown) => error instanceof NotFoundError && error.cause === domainError,
    );
  });

  it("UseCase が DriveError を throw する時、UnavailableError に cause 付きで変換する", async () => {
    // Given: Infrastructure 失敗を throw する Fake UseCase
    const driveError = new DriveError("Drive 読取に失敗");
    const useCase = createFakeGetEpisodeAudioUseCase(async () => {
      throw driveError;
    });
    const controller = createGetEpisodeAudioController(useCase);

    // When: 有効な episodeId で呼ぶ
    const act = controller({ episodeId: "ep-1" });

    // Then: External UnavailableError が Infrastructure を cause に持つ
    await expect(act).rejects.toSatisfy(
      (error: unknown) => error instanceof UnavailableError && error.cause === driveError,
    );
  });
});
