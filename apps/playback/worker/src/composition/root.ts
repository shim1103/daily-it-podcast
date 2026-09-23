import type {
  ListEpisodesResponse,
  ProgressPullResponse,
  ProgressWriteRequest,
  ProgressWriteResponse,
} from "../../../contracts/index.ts";
import type { EpisodeRepository } from "../application/ports/episode-repository.ts";
import {
  StubProgressRepository,
  type ProgressRepository,
} from "../application/ports/progress-repository.ts";
import { completeProgress } from "../application/use-cases/complete-progress.ts";
import { createProgress } from "../application/use-cases/create-progress.ts";
import { getAudio } from "../application/use-cases/get-audio.ts";
import { listEpisodes } from "../application/use-cases/list-episodes.ts";
import { pullProgress } from "../application/use-cases/pull-progress.ts";
import { updateProgress } from "../application/use-cases/update-progress.ts";
import type { GetAudioController } from "../controllers/get-audio-controller.ts";
import { createGetAudioController } from "../controllers/get-audio-controller.ts";
import type { ListEpisodesController } from "../controllers/list-episodes-controller.ts";
import { createListEpisodesController } from "../controllers/list-episodes-controller.ts";
import type { ProgressWriteController } from "../controllers/progress-write-controller.ts";
import { createProgressWriteController } from "../controllers/progress-write-controller.ts";
import type { PullProgressController } from "../controllers/pull-progress-controller.ts";
import { createPullProgressController } from "../controllers/pull-progress-controller.ts";
import { InMemoryEpisodeRepository } from "../infrastructure/in-memory/in-memory-episode-repository.ts";
import { R2EpisodeRepository } from "../infrastructure/r2/r2-episode-repository.ts";
import { validatePlaybackEnv, type PlaybackRepositoryOptions } from "./runtime-config.ts";
import type { PlaybackEnv } from "./runtime-config-bindings.ts";

export type {
  PlaybackRepositoryMode,
  PlaybackRepositoryOptions,
} from "./runtime-config.ts";
export type { PlaybackEnv } from "./runtime-config-bindings.ts";
export { PlaybackRuntimeConfigError } from "./runtime-config-error.ts";

export type PlaybackControllers = {
  listEpisodesController: ListEpisodesController;
  getAudioController: GetAudioController;
  createProgressController: ProgressWriteController;
  updateProgressController: ProgressWriteController;
  completeProgressController: ProgressWriteController;
  pullProgressController: PullProgressController;
};

/**
 * local development / unit test 用に use case 一式を丸ごと差し替える override。
 *
 * @invariant repository 解決（`createEpisodeRepository`）を経由しない。mode 未指定を無視する
 */
export type PlaybackUseCaseOverrides = {
  useCases: {
    listEpisodes: () => Promise<ListEpisodesResponse>;
    getAudio: (episodeId: string) => Promise<Uint8Array>;
    createProgress: (
      episodeId: string,
      body: ProgressWriteRequest,
    ) => Promise<ProgressWriteResponse>;
    updateProgress: (
      episodeId: string,
      body: ProgressWriteRequest,
    ) => Promise<ProgressWriteResponse>;
    completeProgress: (
      episodeId: string,
      body: ProgressWriteRequest,
    ) => Promise<ProgressWriteResponse>;
    pullProgress: (since: string) => Promise<ProgressPullResponse>;
  };
};

/**
 * env から `EpisodeRepository` を選ぶ結果。
 */
export type EpisodeRepositorySelection =
  | { kind: "in-memory"; repository: EpisodeRepository }
  | { kind: "r2"; repository: EpisodeRepository };

/**
 * env から `EpisodeRepository` を選ぶ。
 *
 * @require env は Cloudflare Workers native secrets/vars（`.env` は読まない）。mode は呼び出し側が
 *   常に明示する
 * @ensure 明示的 `options.mode === "in-memory"` の時は "in-memory"、明示的
 *   `options.mode === "r2"` の時は "r2"。mode 未指定・設定不足は runtime config module が throw する
 * @invariant mode の明示指定が必須で、env の中身から暗黙に解決しない
 */
export function createEpisodeRepository(
  env: PlaybackEnv,
  options: PlaybackRepositoryOptions = {},
): EpisodeRepositorySelection {
  const validated = validatePlaybackEnv(env, options);

  if (validated.mode === "r2") {
    return { kind: "r2", repository: new R2EpisodeRepository({ bucket: validated.bucket }) };
  }

  return { kind: "in-memory", repository: new InMemoryEpisodeRepository() };
}

/**
 * A: D1 adapter 未実装のため常に Stub ProgressRepository。
 * binding（`EPISODE_PROGRESS`）の有無検証と D1 adapter 差し替えは C。
 *
 * @ensure 常に ProgressRepository を返す（無言で null にしない）
 */
// todo: D1 adapter（C）で env.EPISODE_PROGRESS を検証し具象へ差し替え。それまで Stub 固定
export function createProgressRepository(_env: PlaybackEnv): ProgressRepository {
  return new StubProgressRepository();
}

/**
 * env から Playback worker の Controller 一式を組み立てる。
 *
 * @require env は Cloudflare Workers native secrets/vars
 * @ensure useCaseOverrides がある時は repository 解決を経由せず、渡された use case を Controller
 *   へ直結する。無い時は従来通り repository を選べれば Controller 一式を返し、設定不足は throw する
 * @invariant useCaseOverrides は既存の in-memory / r2 分岐（`createEpisodeRepository`）を変更しない
 */
export function createPlaybackControllers(
  env: PlaybackEnv,
  options: PlaybackRepositoryOptions = {},
  useCaseOverrides?: PlaybackUseCaseOverrides,
): PlaybackControllers {
  if (useCaseOverrides) {
    const { useCases } = useCaseOverrides;
    return {
      listEpisodesController: createListEpisodesController(useCases.listEpisodes),
      getAudioController: createGetAudioController(useCases.getAudio),
      createProgressController: createProgressWriteController(useCases.createProgress),
      updateProgressController: createProgressWriteController(useCases.updateProgress),
      completeProgressController: createProgressWriteController(useCases.completeProgress),
      pullProgressController: createPullProgressController(useCases.pullProgress),
    };
  }

  const selection = createEpisodeRepository(env, options);
  const progressRepository = createProgressRepository(env);

  const { repository } = selection;
  return {
    listEpisodesController: createListEpisodesController(() =>
      listEpisodes(repository, progressRepository),
    ),
    getAudioController: createGetAudioController((episodeId) => getAudio(repository, episodeId)),
    createProgressController: createProgressWriteController((episodeId, body) =>
      createProgress(progressRepository, { episodeId, ...body }),
    ),
    updateProgressController: createProgressWriteController((episodeId, body) =>
      updateProgress(progressRepository, { episodeId, ...body }),
    ),
    completeProgressController: createProgressWriteController((episodeId, body) =>
      completeProgress(progressRepository, { episodeId, ...body }),
    ),
    pullProgressController: createPullProgressController((since) =>
      pullProgress(progressRepository, since),
    ),
  };
}
