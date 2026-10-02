import { describe, expect, it } from "vitest";
import { createFakeLocalD1Binding } from "../../../test/support/create-fake-local-d1-binding.ts";
import {
  createFakeGetAudioUseCase,
  createFakeListEpisodesUseCase,
  createFakeProgressWriteUseCase,
  createFakePullProgressUseCase,
  validListEpisodesResponse,
  validProgressPullResponse,
  validProgressWriteResponse,
} from "../controllers/fake-use-cases.ts";
import { D1ProgressRepository } from "../infrastructure/d1/d1-progress-repository.ts";
import { InMemoryEpisodeRepository } from "../infrastructure/in-memory/in-memory-episode-repository.ts";
import { InMemoryProgressRepository } from "../infrastructure/in-memory/in-memory-progress-repository.ts";
import { R2EpisodeRepository } from "../infrastructure/r2/r2-episode-repository.ts";
import { SystemClock } from "../infrastructure/system/system-clock.ts";
import { PlaybackRuntimeConfigError } from "./runtime-config-error.ts";
import {
  createClock,
  createEpisodeRepository,
  createPlaybackControllers,
  createProgressRepository,
  type PlaybackRepositoryMode,
} from "./root.ts";

const localMode: PlaybackRepositoryMode = "in-memory";
const r2Mode: PlaybackRepositoryMode = "r2";

const emptyBucket = { get: async () => null, list: async () => ({ objects: [] }) };

function fakeProgressUseCases() {
  return {
    createProgress: createFakeProgressWriteUseCase(),
    updateProgress: createFakeProgressWriteUseCase(),
    completeProgress: createFakeProgressWriteUseCase(),
    pullProgress: createFakePullProgressUseCase(),
  };
}

describe("createEpisodeRepository", () => {
  it("selects_in_memory_episode_repository_when_mode_is_explicit_in_memory", () => {
    // Given: 空 env と明示的な local / unit test mode

    // When: repository を組み立てる
    const got = createEpisodeRepository({}, { mode: localMode });

    // Then: mode の分岐が in-memory 側を選び、Fake の実体が返る
    expect(got.kind).toBe("in-memory");
    expect(got.repository).toBeInstanceOf(InMemoryEpisodeRepository);
  });

  it("selects_r2_episode_repository_when_mode_is_r2_and_episodes_binding_exists", async () => {
    // Given: R2 binding と D1 binding を持つ env と明示的な r2 mode
    const { database } = await createFakeLocalD1Binding();
    const env = { EPISODES: emptyBucket, EPISODE_PROGRESS: database };

    // When: repository を組み立てる
    const got = createEpisodeRepository(env, { mode: r2Mode });

    // Then: mode の分岐が r2 側を選び、R2 Adapter の実体が返る
    expect(got.kind).toBe("r2");
    expect(got.repository).toBeInstanceOf(R2EpisodeRepository);
  });

  it("throws_runtime_config_error_when_mode_is_missing", () => {
    // Given: mode 未指定
    // When / Then: Composition Root は設定不足を返さず throw する
    expect(() => createEpisodeRepository({})).toThrow(PlaybackRuntimeConfigError);
  });

  it("throws_when_mode_is_r2_and_episodes_binding_is_missing", () => {
    // Given: R2 binding が無い env と明示的な r2 mode
    const env = {};

    // When / Then: 未結線を無言 fallback しない
    expect(() => createEpisodeRepository(env, { mode: r2Mode })).toThrow(
      PlaybackRuntimeConfigError,
    );
  });
});

describe("createProgressRepository", () => {
  it("selects_in_memory_progress_repository_when_mode_is_explicit_in_memory", () => {
    // Given: 空 env と明示的な local / unit test mode

    // When: repository を組み立てる
    const got = createProgressRepository({}, { mode: localMode });

    // Then: mode の分岐が in-memory 側を選び、状態を持つ Fake の実体が返る
    //（具象型は構造の確認ではなく、分岐の選択結果を見る）
    expect(got).toBeInstanceOf(InMemoryProgressRepository);
  });

  it("selects_d1_progress_repository_when_mode_is_r2_and_d1_binding_exists", async () => {
    // Given: R2 binding と D1 binding を持つ env と明示的な r2 mode
    const { database } = await createFakeLocalD1Binding();
    const env = { EPISODES: emptyBucket, EPISODE_PROGRESS: database };

    // When: repository を組み立てる
    const got = createProgressRepository(env, { mode: r2Mode });

    // Then: mode の分岐が r2 側を選び、D1 Adapter の実体が返る
    //（具象型は構造の確認ではなく、分岐の選択結果を見る）
    expect(got).toBeInstanceOf(D1ProgressRepository);
  });

  it("throws_runtime_config_error_when_mode_is_missing", () => {
    // Given: mode 未指定
    // When / Then: Composition Root は設定不足を返さず throw する
    expect(() => createProgressRepository({})).toThrow(PlaybackRuntimeConfigError);
  });

  it("throws_without_falling_back_to_in_memory_when_mode_is_r2_and_d1_binding_is_missing", () => {
    // Given: R2 binding だけを持つ env（進捗が保存できたように見えて消える構成）
    const env = { EPISODES: emptyBucket };

    // When / Then: 無言 fallback せず、D1 の未結線を runtime config error にする
    expect(() => createProgressRepository(env, { mode: r2Mode })).toThrow(
      PlaybackRuntimeConfigError,
    );
    expect(() => createProgressRepository(env, { mode: r2Mode })).toThrow("EPISODE_PROGRESS");
  });
});

describe("createPlaybackControllers", () => {
  describe("episode 系 Controller は選ばれた EpisodeRepository を通る", () => {
    it("lists_no_episodes_when_mode_is_in_memory_and_episode_repository_is_empty", async () => {
      // Given: override 無し・in-memory mode の空 env（原稿が 1 件も無い EpisodeRepository）
      const got = createPlaybackControllers({}, { mode: localMode });

      // When: 一覧 Controller を叩く
      const list = await got.listEpisodesController();

      // Then: 空の in-memory EpisodeRepository を通り、一覧は空になる
      expect(list.episodes).toEqual([]);
    });

    it("rejects_audio_as_not_found_when_mode_is_in_memory_and_episode_repository_is_empty", async () => {
      // Given: override 無し・in-memory mode の空 env（音声も原稿も無い EpisodeRepository）
      const got = createPlaybackControllers({}, { mode: localMode });

      // When: 存在しない episode の音声 Controller を叩く
      const audio = got.getAudioController("missing");

      // Then: 空の in-memory EpisodeRepository を通り、Domain 経由の External NotFound になる
      await expect(audio).rejects.toMatchObject({ name: "NotFoundError" });
    });

    it("rejects_complete_as_not_found_when_mode_is_in_memory_and_episode_repository_has_no_manuscript", async () => {
      // Given: override 無し・in-memory mode の空 env（原稿が無い EpisodeRepository）
      const got = createPlaybackControllers({}, { mode: localMode });

      // When: complete Controller を叩く
      const complete = got.completeProgressController("ep-1", {
        positionSec: 57,
        clientAt: "2026-09-22T10:05:00.000Z",
      });

      // Then: 原稿を選ばれた EpisodeRepository から引けず、EpisodeContentError → External NotFound になる
      await expect(complete).rejects.toMatchObject({ name: "NotFoundError" });
    });
  });

  describe("progress 系 Controller は選ばれた ProgressRepository を通る", () => {
    it("persists_created_progress_for_pull_when_mode_is_in_memory", async () => {
      // Given: override 無し・in-memory mode の空 env で組み立てた Controller 一式
      const got = createPlaybackControllers({}, { mode: localMode });

      // When: 同じ組み立ての中で progress を作成し、作成より前の since で pull する
      await got.createProgressController("ep-1", {
        positionSec: 12,
        clientAt: "2026-09-22T10:00:00.000Z",
      });
      const pull = await got.pullProgressController("2026-09-22T09:00:00.000Z");

      // Then: create と pull が同じ ProgressRepository を共有し、作成した進捗が読める
      expect(pull).toEqual({
        episodes: [
          {
            episodeId: "ep-1",
            progress: {
              positionSec: 12,
              firstPlayedAt: "2026-09-22T10:00:00.000Z",
              firstCompletedAt: null,
              lastPlayedAt: "2026-09-22T10:00:00.000Z",
            },
          },
        ],
      });
    });

    it("updates_created_progress_when_mode_is_in_memory", async () => {
      // Given: in-memory mode の Controller 一式と、同じ組み立ての中で作成済みの progress
      const got = createPlaybackControllers({}, { mode: localMode });
      await got.createProgressController("ep-1", {
        positionSec: 12,
        clientAt: "2026-09-22T10:00:00.000Z",
      });

      // When: 同じ episode を後の clientAt で更新し、pull する
      const update = await got.updateProgressController("ep-1", {
        positionSec: 30,
        clientAt: "2026-09-22T10:01:00.000Z",
      });
      const pull = await got.pullProgressController("2026-09-22T09:00:00.000Z");

      // Then: update が create と同じ ProgressRepository の行を見つけて通り（行なし 404 にならない）、
      // pull は後勝ちの位置を返す
      expect(update).toEqual({
        firstPlayedAt: "2026-09-22T10:00:00.000Z",
        firstCompletedAt: null,
      });
      expect(pull.episodes).toHaveLength(1);
      expect(pull.episodes[0]?.progress).toMatchObject({
        positionSec: 30,
        lastPlayedAt: "2026-09-22T10:01:00.000Z",
      });
    });

    it("sends_no_sql_to_d1_when_mode_is_in_memory_even_if_env_has_d1_binding", async () => {
      // Given: D1 binding を持つ env と in-memory mode（in-memory は env の中身を見ない）
      const { database, calls } = await createFakeLocalD1Binding();
      const got = createPlaybackControllers({ EPISODE_PROGRESS: database }, { mode: localMode });

      // When: progress を作成する
      const write = await got.createProgressController("ep-1", {
        positionSec: 1,
        clientAt: "2026-10-01T00:00:00.000Z",
      });

      // Then: D1 へは SQL を渡さず、merge 結果の応答だけを返す
      expect(write).toEqual({ firstPlayedAt: "2026-10-01T00:00:00.000Z", firstCompletedAt: null });
      expect(calls).toHaveLength(0);
    });

    it("persists_progress_via_d1_binding_when_mode_is_r2_and_d1_binding_exists", async () => {
      // Given: 行の無い D1 binding と、R2 binding を持つ env
      const { database, calls } = await createFakeLocalD1Binding();
      const got = createPlaybackControllers(
        { EPISODES: emptyBucket, EPISODE_PROGRESS: database },
        { mode: r2Mode },
      );

      // When: progress を作成する
      const write = await got.createProgressController("ep-1", {
        positionSec: 1,
        clientAt: "2026-10-01T00:00:00.000Z",
      });

      // Then: merge 結果の応答が返り、D1 へは読取と upsert の 2 文が渡る
      expect(write).toEqual({ firstPlayedAt: "2026-10-01T00:00:00.000Z", firstCompletedAt: null });
      expect(calls).toHaveLength(2);
    });
  });

  it("uses_stub_use_cases_ignoring_missing_mode_when_use_cases_override_exists", async () => {
    // Given: mode 未指定の env と、stub use case 一式の override
    const env = {};
    const useCases = {
      listEpisodes: createFakeListEpisodesUseCase(),
      getAudio: createFakeGetAudioUseCase(),
      ...fakeProgressUseCases(),
    };

    // When: override 付きで Controller 一式を組み立てる
    const got = createPlaybackControllers(env, {}, { useCases });

    // Then: repository 解決を経由せず、stub use case の応答をそのまま返す
    await expect(got.listEpisodesController()).resolves.toEqual(validListEpisodesResponse);
    await expect(
      got.createProgressController("ep-1", {
        positionSec: 1,
        clientAt: "2026-09-22T10:00:00.000Z",
      }),
    ).resolves.toEqual(validProgressWriteResponse);
    await expect(got.pullProgressController("2026-09-22T10:00:00.000Z")).resolves.toEqual(
      validProgressPullResponse,
    );
  });

  it("throws_without_building_controllers_when_mode_is_missing", () => {
    // Given: mode 未指定の env
    const env = {};

    // When / Then: 設定不足を返さず、Controller も組み立てない
    expect(() => createPlaybackControllers(env)).toThrow(PlaybackRuntimeConfigError);
  });
});

describe("createClock", () => {
  it("returns_system_clock_regardless_of_env", () => {
    // Given: env も mode も渡さない

    // When: Clock を組み立てる
    const got = createClock();

    // Then: システム時計の実装が返る
    expect(got).toBeInstanceOf(SystemClock);
  });
});
