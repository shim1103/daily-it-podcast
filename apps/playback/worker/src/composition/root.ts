import type { Clock } from "../application/ports/clock.ts";
import type { EpisodeRepository } from "../application/ports/episode-repository.ts";
import type { ProgressRepository } from "../application/ports/progress-repository.ts";
import { completeProgress } from "../application/use-cases/complete-progress.ts";
import { createProgress } from "../application/use-cases/create-progress.ts";
import { getAudio } from "../application/use-cases/get-audio.ts";
import { listEpisodes } from "../application/use-cases/list-episodes.ts";
import { pullProgress } from "../application/use-cases/pull-progress.ts";
import { updateProgress } from "../application/use-cases/update-progress.ts";
import type { GetAudioController, GetAudioUseCase } from "../controllers/get-audio-controller.ts";
import { createGetAudioController } from "../controllers/get-audio-controller.ts";
import type {
  ListEpisodesController,
  ListEpisodesUseCase,
} from "../controllers/list-episodes-controller.ts";
import { createListEpisodesController } from "../controllers/list-episodes-controller.ts";
import type {
  ProgressWriteController,
  ProgressWriteUseCase,
} from "../controllers/progress-write-controller.ts";
import { createProgressWriteController } from "../controllers/progress-write-controller.ts";
import type {
  PullProgressController,
  PullProgressUseCase,
} from "../controllers/pull-progress-controller.ts";
import { createPullProgressController } from "../controllers/pull-progress-controller.ts";
import { D1ProgressRepository } from "../infrastructure/d1/d1-progress-repository.ts";
import { InMemoryEpisodeRepository } from "../infrastructure/in-memory/in-memory-episode-repository.ts";
import { InMemoryProgressRepository } from "../infrastructure/in-memory/in-memory-progress-repository.ts";
import { R2EpisodeRepository } from "../infrastructure/r2/r2-episode-repository.ts";
import { SystemClock } from "../infrastructure/system/system-clock.ts";
import {
  validateEpisodeEnv,
  validateProgressEnv,
  type PlaybackRepositoryOptions,
} from "./runtime-config.ts";
import type { PlaybackEnv } from "./runtime-config-bindings.ts";

export type {
  PlaybackProgressRepositoryMode,
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
 * @invariant repository 解決（`createEpisodeRepository`・`createProgressRepository`）を経由しない。
 *   episode の mode・進捗 mode が未指定でも throw しない
 */
export type PlaybackUseCaseOverrides = {
  useCases: {
    listEpisodes: ListEpisodesUseCase;
    getAudio: GetAudioUseCase;
    createProgress: ProgressWriteUseCase;
    updateProgress: ProgressWriteUseCase;
    completeProgress: ProgressWriteUseCase;
    pullProgress: PullProgressUseCase;
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
 * @require env は Cloudflare Workers native secrets/vars（`.env` は読まない）。`options.mode` は
 *   呼び出し側が常に明示する
 * @ensure 明示的 `options.mode === "in-memory"` の時は "in-memory"、明示的
 *   `options.mode === "r2"` の時は "r2"。`options.mode` 未指定、および r2 での `EPISODES` 欠落は
 *   runtime config module が throw する
 * @invariant 進捗の設定（`options.progressMode`・`EPISODE_PROGRESS`）に依存しない。mode の明示指定が
 *   必須で、env の中身から暗黙に解決しない
 */
export function createEpisodeRepository(
  env: PlaybackEnv,
  options: PlaybackRepositoryOptions = {},
): EpisodeRepositorySelection {
  const validated = validateEpisodeEnv(env, options);

  if (validated.mode === "r2") {
    return { kind: "r2", repository: new R2EpisodeRepository({ bucket: validated.bucket }) };
  }

  return { kind: "in-memory", repository: new InMemoryEpisodeRepository() };
}

/**
 * env から `ProgressRepository` を選ぶ。
 *
 * @require env は Cloudflare Workers native secrets/vars。`options.progressMode` は呼び出し側が
 *   常に明示する
 * @ensure 明示的 `options.progressMode === "d1"` の時は `EPISODE_PROGRESS`（D1 binding）の
 *   `D1ProgressRepository`、明示的 `options.progressMode === "in-memory"` の時は env の中身を見ず、
 *   行を保持する新しい `InMemoryProgressRepository` を返す。`options.progressMode` 未指定、および d1 での
 *   D1 binding 欠落は runtime config module が throw する
 * @invariant episode の設定（`options.mode`・`EPISODES`）に依存しない。d1 mode の D1 欠落を
 *   `InMemoryProgressRepository` へ無言で逃がさない
 */
export function createProgressRepository(
  env: PlaybackEnv,
  options: PlaybackRepositoryOptions = {},
): ProgressRepository {
  const validated = validateProgressEnv(env, options);

  if (validated.mode === "d1") {
    return new D1ProgressRepository({ database: validated.progressDatabase });
  }

  return new InMemoryProgressRepository();
}

/**
 * 現在時刻の `Clock` を返す。
 *
 * @ensure env・mode に依らず、システム時計を返す `SystemClock` を返す
 */
export function createClock(): Clock {
  // todo: C が Write UseCase へ clock を注入したら、この todo を消す（`createPlaybackControllers` はまだ呼ばない）
  return new SystemClock();
}

/**
 * env から Playback worker の Controller 一式を組み立てる。
 *
 * @require env は Cloudflare Workers native secrets/vars。useCaseOverrides が無い時、
 *   `options.mode`（episode）と `options.progressMode`（進捗）は呼び出し側がどちらも明示する
 * @ensure useCaseOverrides がある時は repository 解決を経由せず、渡された use case を Controller
 *   へ直結する。無い時は episode と進捗の repository をそれぞれの mode で選べれば Controller 一式を返し、
 *   設定不足（どちらかの mode の未指定、r2 mode の `EPISODES` 欠落、d1 mode の D1 binding 欠落）は
 *   throw する
 * @invariant episode の永続先と進捗の永続先は独立に選ぶ（組合せは自由）。useCaseOverrides は
 *   `createEpisodeRepository`・`createProgressRepository` の分岐を変更しない
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
    getAudioController: createGetAudioController((input) => getAudio(repository, input)),
    createProgressController: createProgressWriteController((input) =>
      createProgress(progressRepository, input),
    ),
    updateProgressController: createProgressWriteController((input) =>
      updateProgress(progressRepository, input),
    ),
    completeProgressController: createProgressWriteController((input) =>
      completeProgress(repository, progressRepository, input),
    ),
    pullProgressController: createPullProgressController((input) =>
      pullProgress(progressRepository, input),
    ),
  };
}
