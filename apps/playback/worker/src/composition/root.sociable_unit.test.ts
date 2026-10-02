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
import { InMemoryEpisodeRepository } from "../infrastructure/in-memory/in-memory-episode-repository.ts";
import { R2EpisodeRepository } from "../infrastructure/r2/r2-episode-repository.ts";
import { PlaybackRuntimeConfigError } from "./runtime-config-error.ts";
import {
  createEpisodeRepository,
  createPlaybackControllers,
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

    // Then: Fake が選ばれる
    expect(got.kind).toBe("in-memory");
    if (got.kind === "in-memory") {
      expect(got.repository).toBeInstanceOf(InMemoryEpisodeRepository);
    }
  });

  it("throws_runtime_config_error_when_mode_is_missing", () => {
    // Given: mode 未指定
    // When / Then: Composition Root は設定不足を返さず throw する
    expect(() => createEpisodeRepository({})).toThrow(PlaybackRuntimeConfigError);
  });

  it("selects_r2_episode_repository_when_mode_is_r2_and_episodes_binding_exists", async () => {
    // Given: R2 binding と D1 binding を持つ env と明示的な r2 mode
    const { database } = await createFakeLocalD1Binding();
    const env = { EPISODES: emptyBucket, EPISODE_PROGRESS: database };

    // When: repository を組み立てる
    const got = createEpisodeRepository(env, { mode: r2Mode });

    // Then: R2 Adapter が選ばれる
    expect(got.kind).toBe("r2");
    if (got.kind === "r2") {
      expect(got.repository).toBeInstanceOf(R2EpisodeRepository);
    }
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

describe("createPlaybackControllers", () => {
  it("builds_controllers_from_env_when_mode_is_explicit_in_memory", () => {
    // Given: 空 env（意図的な Fake 利用）
    const env = {};

    // When: Controller 一式を組み立てる
    const got = createPlaybackControllers(env, { mode: localMode });

    // Then: ready として Controller 一式が返る
    expect(got.listEpisodesController).toBeDefined();
    expect(got.createProgressController).toBeDefined();
    expect(got.pullProgressController).toBeDefined();
  });

  it("throws_without_building_controllers_when_mode_is_missing", () => {
    // Given: mode 未指定の env
    const env = {};

    // When / Then: 設定不足を返さず、Controller も組み立てない
    expect(() => createPlaybackControllers(env)).toThrow(PlaybackRuntimeConfigError);
  });

  it("routes_progress_through_stub_when_in_memory_mode_has_no_override_and_no_d1_binding", async () => {
    // Given: override 無し・in-memory mode・D1 binding 無しの空 env（in-memory は progress を永続しない Stub）
    const got = createPlaybackControllers({}, { mode: localMode });

    // When: 一覧・音声・progress 経路を叩く
    const list = await got.listEpisodesController();
    const audio = got.getAudioController("missing");
    const write = await got.createProgressController("ep-1", {
      positionSec: 1,
      clientAt: "2026-09-22T10:00:00.000Z",
    });
    const update = got.updateProgressController("ep-1", {
      positionSec: 2,
      clientAt: "2026-09-22T10:01:00.000Z",
    });
    const complete = got.completeProgressController("ep-1", {
      positionSec: 57,
      clientAt: "2026-09-22T10:05:00.000Z",
    });
    const pull = await got.pullProgressController("2026-09-22T10:00:00.000Z");

    // Then: 空 repository を検証純関数が通し、一覧は空・音声は Domain 経由の External NotFound
    // Stub Progress は永続しないため create は merge 応答・update は行なし 404
    // complete は原稿無しで EpisodeContentError → External NotFound
    expect(list.episodes).toEqual([]);
    await expect(audio).rejects.toMatchObject({ name: "NotFoundError" });
    expect(write).toEqual({
      firstPlayedAt: "2026-09-22T10:00:00.000Z",
      firstCompletedAt: null,
    });
    await expect(update).rejects.toMatchObject({ name: "NotFoundError" });
    await expect(complete).rejects.toMatchObject({ name: "NotFoundError" });
    expect(pull).toEqual({ episodes: [] });
  });

  it("routes_progress_through_stub_when_in_memory_mode_even_if_env_has_d1_binding", async () => {
    // Given: D1 binding を持つ env と in-memory mode（in-memory は env の中身を見ない）
    const { database, calls } = await createFakeLocalD1Binding();
    const got = createPlaybackControllers({ EPISODE_PROGRESS: database }, { mode: localMode });

    // When: progress を作成する
    const write = await got.createProgressController("ep-1", {
      positionSec: 1,
      clientAt: "2026-10-01T00:00:00.000Z",
    });

    // Then: D1 へは SQL を渡さず、Stub は merge 結果の応答だけを返す
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

  it("throws_without_falling_back_to_stub_when_mode_is_r2_and_d1_binding_is_missing", () => {
    // Given: R2 binding だけを持つ env（進捗が保存できたように見えて消える構成）
    const env = { EPISODES: emptyBucket };

    // When / Then: 無言 fallback せず、D1 の未結線を runtime config error にする
    expect(() => createPlaybackControllers(env, { mode: r2Mode })).toThrow(
      PlaybackRuntimeConfigError,
    );
    expect(() => createPlaybackControllers(env, { mode: r2Mode })).toThrow("EPISODE_PROGRESS");
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
});
