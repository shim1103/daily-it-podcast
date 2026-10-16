import { describe, expect, it } from "vitest";
import { NotFoundError, UnavailableError } from "../../../contracts/index.ts";
import { EpisodeContentError } from "../entities/errors/episode-content-error.ts";
import { R2Error } from "../infrastructure/r2/r2-error.ts";
import { validAudioBytes } from "../test/fixtures/audio-bytes.ts";
import { createGetAudioController } from "./get-audio-controller.ts";

/**
 * scope: Sociable Unit
 * real: createGetAudioController, map-internal-error
 * double: GetAudioUseCase を test 内の Stub に差し替え
 */
describe("createGetAudioController", () => {
  it("UseCase が成功する時、UseCase が返した音声 byte をそのまま返す", async () => {
    // Given: mp3 byte を返す Stub UseCase
    const controller = createGetAudioController(async () => validAudioBytes);

    // When: 検証済み episodeId を渡す
    const got = await controller("ep-1");

    // Then: 同一参照の byte を返す
    expect(got).toBe(validAudioBytes);
  });

  it("episodeId は { episodeId } の Input として UseCase に渡される", async () => {
    // Given: 受け取った Input を記録する Stub UseCase
    const received: unknown[] = [];
    const controller = createGetAudioController(async (input) => {
      received.push(input);
      return validAudioBytes;
    });

    // When: episodeId を渡す
    await controller("ep-7");

    // Then: episodeId だけの Input を 1 回渡す
    expect(received).toEqual([{ episodeId: "ep-7" }]);
  });

  it("UseCase が EpisodeContentError を throw する時、NotFoundError に cause 付きで変換する", async () => {
    // Given: Domain 不在を throw する Stub UseCase
    const domainError = new EpisodeContentError("音声が無い: ep-1");
    const controller = createGetAudioController(async () => {
      throw domainError;
    });

    // When: 有効な episodeId で呼ぶ
    const act = controller("ep-1");

    // Then: External NotFoundError が Domain を cause に持つ
    await expect(act).rejects.toSatisfy(
      (error: unknown) => error instanceof NotFoundError && error.cause === domainError,
    );
  });

  it("UseCase が R2Error を throw する時、UnavailableError に cause 付きで変換する", async () => {
    // Given: Infrastructure 失敗を throw する Stub UseCase
    const r2Error = new R2Error("R2 読取に失敗");
    const controller = createGetAudioController(async () => {
      throw r2Error;
    });

    // When: 有効な episodeId で呼ぶ
    const act = controller("ep-1");

    // Then: External UnavailableError が Infrastructure を cause に持つ
    await expect(act).rejects.toSatisfy(
      (error: unknown) => error instanceof UnavailableError && error.cause === r2Error,
    );
  });
});
