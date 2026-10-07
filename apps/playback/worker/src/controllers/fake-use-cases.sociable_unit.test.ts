import { describe, expect, it } from "vitest";
import { EpisodeContentError } from "../entities/errors/episode-content-error.ts";
import { createFakeGetAudioUseCase } from "./fake-use-cases.ts";

/**
 * scope: Sociable Unit
 * real: createFakeGetAudioUseCase, fake-episodes.json, createFakeEpisodeAudioBytes
 * double: なし
 */
describe("createFakeGetAudioUseCase", () => {
  it("returns_same_audio_bytes_instance_when_same_episode_is_requested_twice", async () => {
    // Given: impl を渡さない fake の GetAudio use case
    const getAudio = createFakeGetAudioUseCase();

    // When: 同じ episodeId で 2 回呼ぶ
    const first = await getAudio({ episodeId: "ep-1" });
    const second = await getAudio({ episodeId: "ep-1" });

    // Then: 2 回目は cache した同一の byte 列を返す
    expect(second).toBe(first);
  });

  it("rejects_with_episode_content_error_when_episode_is_not_in_fake_episodes", async () => {
    // Given: impl を渡さない fake の GetAudio use case
    const getAudio = createFakeGetAudioUseCase();

    // When: fake の原稿に無い episodeId で呼ぶ
    const got = getAudio({ episodeId: "missing" });

    // Then: Domain Error の EpisodeContentError で reject する
    await expect(got).rejects.toBeInstanceOf(EpisodeContentError);
  });
});
