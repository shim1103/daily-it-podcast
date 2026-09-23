import { describe, expect, it } from "vitest";
import {
  createFakeGetAudioUseCase,
  createFakeListEpisodesUseCase,
  createFakeProgressWriteUseCase,
  createFakePullProgressUseCase,
  validListEpisodesResponse,
  validProgressPullResponse,
  validProgressWriteResponse,
} from "../controllers/fake-use-cases.ts";
import { StubProgressRepository } from "../application/ports/progress-repository.ts";
import { InMemoryEpisodeRepository } from "../infrastructure/in-memory/in-memory-episode-repository.ts";
import { R2EpisodeRepository } from "../infrastructure/r2/r2-episode-repository.ts";
import { PlaybackRuntimeConfigError } from "./runtime-config-error.ts";
import {
  createEpisodeRepository,
  createPlaybackControllers,
  createProgressRepository,
  type PlaybackRepositoryMode,
} from "./root.ts";

const localMode: PlaybackRepositoryMode = "in-memory";
const r2Mode: PlaybackRepositoryMode = "r2";

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

  it("明示的 r2 mode で EPISODES binding がある時、R2EpisodeRepository を選ぶ", () => {
    // Given: R2 binding を持つ env と明示的な r2 mode
    const bucket = { get: async () => null, list: async () => ({ objects: [] }) };
    const env = { EPISODES: bucket };

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

describe("createProgressRepository", () => {
  it("A では常に StubProgressRepository を返す", () => {
    // Given: D1 binding の有無に依らない
    // When: ProgressRepository を選ぶ
    const got = createProgressRepository({});

    // Then: Stub（D1 adapter 差し替えは C）
    expect(got).toBeInstanceOf(StubProgressRepository);
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

  it("override 無し in-memory mode の Controller は repository → use-case → 検証純関数を通る", async () => {
    // Given: override 無し・in-memory mode（repository は空で組み立てられる）
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
    // progress は Stub の zero / 空
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
