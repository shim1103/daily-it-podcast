import { describe, expect, it } from "vitest";
import { NotFoundError, UnavailableError } from "../../../contracts/index.ts";
import { EpisodeContentError } from "../entities/errors/episode-content-error.ts";
import { R2Error } from "../infrastructure/r2/r2-error.ts";
import { createGetAudioController } from "./get-audio-controller.ts";
import { createFakeEpisodeAudioBytes } from "../test/fixtures/audio-bytes.ts";
import { createFakeGetAudioUseCase, validEpisodeItem } from "./fake-use-cases.ts";

describe("createGetAudioController", () => {
  it("UseCase が成功する時、音声 byte を返す", async () => {
    // Given: mp3 byte を返す Fake UseCase
    const useCase = createFakeGetAudioUseCase();
    const controller = createGetAudioController(useCase);

    // When: 検証済み episodeId を渡す
    const got = await controller("ep-1");
    const expected = createFakeEpisodeAudioBytes(validEpisodeItem.durationSec);

    // Then: Fake が返した再生可能 mp3 と一致する
    expect(Buffer.from(got).equals(Buffer.from(expected))).toBe(true);
  });

  it("同一 episodeId を連続取得する時、キャッシュ済み mp3 を返す", async () => {
    // Given: mp3 byte を返す Fake UseCase
    const useCase = createFakeGetAudioUseCase();
    const controller = createGetAudioController(useCase);

    // When: 同じ episodeId で2回取得する
    const first = await controller("ep-1");
    const second = await controller("ep-1");

    // Then: 同一参照の byte を返す
    expect(second).toBe(first);
  });

  it("存在しない episodeId の時、NotFoundError に変換して throw する", async () => {
    // Given: fake data に無い episodeId
    const useCase = createFakeGetAudioUseCase();
    const controller = createGetAudioController(useCase);

    // When: 存在しない episodeId で呼ぶ
    const act = controller("missing");

    // Then: Domain 不在が External NotFound になる
    await expect(act).rejects.toBeInstanceOf(NotFoundError);
  });

  it("UseCase が EpisodeContentError を throw する時、NotFoundError に cause 付きで変換する", async () => {
    // Given: Domain 不在を throw する Fake UseCase
    const domainError = new EpisodeContentError("音声エントリが無い: ep-1");
    const useCase = createFakeGetAudioUseCase(async () => {
      throw domainError;
    });
    const controller = createGetAudioController(useCase);

    // When: 有効な episodeId で呼ぶ
    const act = controller("ep-1");

    // Then: External NotFoundError が Domain を cause に持つ
    await expect(act).rejects.toSatisfy(
      (error: unknown) => error instanceof NotFoundError && error.cause === domainError,
    );
  });

  it("UseCase が R2Error を throw する時、UnavailableError に cause 付きで変換する", async () => {
    // Given: Infrastructure 失敗を throw する Fake UseCase
    const r2Error = new R2Error("R2 読取に失敗");
    const useCase = createFakeGetAudioUseCase(async () => {
      throw r2Error;
    });
    const controller = createGetAudioController(useCase);

    // When: 有効な episodeId で呼ぶ
    const act = controller("ep-1");

    // Then: External UnavailableError が Infrastructure を cause に持つ
    await expect(act).rejects.toSatisfy(
      (error: unknown) => error instanceof UnavailableError && error.cause === r2Error,
    );
  });
});
