import type {
  ListEpisodesResponse,
  ProgressPullResponse,
  ProgressWriteRequest,
  ProgressWriteResponse,
} from "../../../contracts/index.ts";
import type { EpisodeRepository } from "../application/ports/episode-repository.ts";
import type { ProgressRepository } from "../application/ports/progress-repository.ts";
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
import { D1ProgressRepository } from "../infrastructure/d1/d1-progress-repository.ts";
import { InMemoryEpisodeRepository } from "../infrastructure/in-memory/in-memory-episode-repository.ts";
import { InMemoryProgressRepository } from "../infrastructure/in-memory/in-memory-progress-repository.ts";
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
 * env から `ProgressRepository` を選ぶ。
 *
 * @require env は Cloudflare Workers native secrets/vars。mode は呼び出し側が常に明示する
 * @ensure 明示的 `options.mode === "r2"` の時は `EPISODE_PROGRESS`（D1 binding）の `D1ProgressRepository`、
 *   明示的 `options.mode === "in-memory"` の時は env の中身を見ず、行を保持する新しい
 *   `InMemoryProgressRepository` を返す。mode 未指定、および r2 での D1 binding 欠落は
 *   runtime config module が throw する
 * @invariant r2 mode の D1 欠落を `InMemoryProgressRepository` へ無言で逃がさない
 */
export function createProgressRepository(
  env: PlaybackEnv,
  options: PlaybackRepositoryOptions = {},
): ProgressRepository {
  const validated = validatePlaybackEnv(env, options);

  if (validated.mode === "r2") {
    return new D1ProgressRepository({ database: validated.progressDatabase });
  }

  return new InMemoryProgressRepository();
}

/**
 * env から Playback worker の Controller 一式を組み立てる。
 *
 * @require env は Cloudflare Workers native secrets/vars
 * @ensure useCaseOverrides がある時は repository 解決を経由せず、渡された use case を Controller
 *   へ直結する。無い時は episode と progress の repository を選べれば Controller 一式を返し、
 *   設定不足（r2 mode の R2・D1 binding 欠落を含む）は throw する
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
  const progressRepository = createProgressRepository(env, options);

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
      completeProgress(repository, progressRepository, { episodeId, ...body }),
    ),
    pullProgressController: createPullProgressController((since) =>
      pullProgress(progressRepository, since),
    ),
  };
}
