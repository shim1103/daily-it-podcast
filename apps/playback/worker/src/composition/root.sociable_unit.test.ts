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
import { episodeProgressColumns } from "../infrastructure/d1/progress-d1-constants.ts";
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
  it("明示的な in-memory mode の時、InMemoryEpisodeRepository を選ぶ", () => {
    // Given: 空 env と明示的な local / unit test mode

    // When: repository を組み立てる
    const got = createEpisodeRepository({}, { mode: localMode });

    // Then: Fake が選ばれる
    expect(got.kind).toBe("in-memory");
    if (got.kind === "in-memory") {
      expect(got.repository).toBeInstanceOf(InMemoryEpisodeRepository);
    }
  });

  it("mode が無い時、runtime config error を throw する", () => {
    // Given: mode 未指定
    // When / Then: Composition Root は設定不足を返さず throw する
    expect(() => createEpisodeRepository({})).toThrow(PlaybackRuntimeConfigError);
  });

  it("明示的 r2 mode で EPISODES binding がある時、R2EpisodeRepository を選ぶ", async () => {
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

  it("明示的 r2 mode で EPISODES binding が無い時、throw する", () => {
    // Given: R2 binding が無い env と明示的な r2 mode
    const env = {};

    // When / Then: 未結線を無言 fallback しない
    expect(() => createEpisodeRepository(env, { mode: r2Mode })).toThrow(
      PlaybackRuntimeConfigError,
    );
  });
});

describe("createPlaybackControllers", () => {
  it("明示的な in-memory mode の時、その env に基づく Controller 一式を組み立てる", () => {
    // Given: 空 env（意図的な Fake 利用）
    const env = {};

    // When: Controller 一式を組み立てる
    const got = createPlaybackControllers(env, { mode: localMode });

    // Then: ready として Controller 一式が返る
    expect(got.listEpisodesController).toBeDefined();
    expect(got.createProgressController).toBeDefined();
    expect(got.pullProgressController).toBeDefined();
  });

  it("mode が無い時、throw して Controller を組み立てない", () => {
    // Given: mode 未指定の env
    const env = {};

    // When / Then: 設定不足を返さず、Controller も組み立てない
    expect(() => createPlaybackControllers(env)).toThrow(PlaybackRuntimeConfigError);
  });

  it("override 無し・D1 binding 無しの in-memory mode の Controller は、Stub の progress を通る", async () => {
    // Given: override 無し・in-memory mode・D1 binding 無しの空 env（in-memory は progress を永続しない Stub）
    const got = createPlaybackControllers({}, { mode: localMode });

    // When: 一覧・音声・progress 経路を叩く
    const list = await got.listEpisodesController();
    const audio = got.getAudioController("missing");
    const write = await got.createProgressController("ep-1", {
      positionSec: 1,
      clientAt: "2026-09-22T10:00:00.000Z",
    });
    const update = await got.updateProgressController("ep-1", {
      positionSec: 2,
      clientAt: "2026-09-22T10:01:00.000Z",
    });
    const complete = await got.completeProgressController("ep-1", {
      positionSec: 57,
      clientAt: "2026-09-22T10:05:00.000Z",
    });
    const pull = await got.pullProgressController("2026-09-22T10:00:00.000Z");

    // Then: 空 repository を検証純関数が通し、一覧は空・音声は Domain 経由の External NotFound
    // progress は永続しない Stub が返す zero / 空
    expect(list.episodes).toEqual([]);
    await expect(audio).rejects.toMatchObject({ name: "NotFoundError" });
    expect(write).toEqual({
      firstPlayedAt: "1970-01-01T00:00:00.000Z",
      firstCompletedAt: null,
    });
    expect(update).toEqual({
      firstPlayedAt: "1970-01-01T00:00:00.000Z",
      firstCompletedAt: null,
    });
    expect(complete).toEqual({
      firstPlayedAt: "1970-01-01T00:00:00.000Z",
      firstCompletedAt: null,
    });
    expect(pull).toEqual({ episodes: [] });
  });

  it("明示的な in-memory mode の時、env に D1 binding があっても progress は Stub を通る", async () => {
    // Given: D1 binding を持つ env と in-memory mode（in-memory は env の中身を見ない）
    const { database, calls } = await createFakeLocalD1Binding({
      first: async () => ({
        [episodeProgressColumns.firstPlayedAt]: "2026-10-01T00:00:00.000Z",
        [episodeProgressColumns.firstCompletedAt]: null,
      }),
    });
    const got = createPlaybackControllers({ EPISODE_PROGRESS: database }, { mode: localMode });

    // When: progress を作成する
    const write = await got.createProgressController("ep-1", {
      positionSec: 1,
      clientAt: "2026-10-01T00:00:00.000Z",
    });

    // Then: D1 へは SQL を渡さず、Stub の zero 値が返る
    expect(write).toEqual({ firstPlayedAt: "1970-01-01T00:00:00.000Z", firstCompletedAt: null });
    expect(calls).toHaveLength(0);
  });

  it("明示的 r2 mode で D1 binding がある時、progress Controller は D1 binding を介して応答を返す", async () => {
    // Given: 書込の勝ち側として first* 行を返す D1 binding と、R2 binding を持つ env
    const { database, calls } = await createFakeLocalD1Binding({
      first: async () => ({
        [episodeProgressColumns.firstPlayedAt]: "2026-10-01T00:00:00.000Z",
        [episodeProgressColumns.firstCompletedAt]: null,
      }),
    });
    const got = createPlaybackControllers(
      { EPISODES: emptyBucket, EPISODE_PROGRESS: database },
      { mode: r2Mode },
    );

    // When: progress を作成する
    const write = await got.createProgressController("ep-1", {
      positionSec: 1,
      clientAt: "2026-10-01T00:00:00.000Z",
    });

    // Then: Stub の zero 値ではなく D1 binding の応答が返り、D1 へ SQL が渡る
    expect(write).toEqual({ firstPlayedAt: "2026-10-01T00:00:00.000Z", firstCompletedAt: null });
    expect(calls).toHaveLength(1);
  });

  it("明示的 r2 mode で D1 binding が無い時、Stub へ落とさず throw して Controller を組み立てない", () => {
    // Given: R2 binding だけを持つ env（進捗が保存できたように見えて消える構成）
    const env = { EPISODES: emptyBucket };

    // When / Then: 無言 fallback せず、D1 の未結線を runtime config error にする
    expect(() => createPlaybackControllers(env, { mode: r2Mode })).toThrow(
      PlaybackRuntimeConfigError,
    );
    expect(() => createPlaybackControllers(env, { mode: r2Mode })).toThrow("EPISODE_PROGRESS");
  });

  it("useCases override がある時、mode 未指定を無視して stub use case を使う", async () => {
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
