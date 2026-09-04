import type { ListEpisodesResponse } from "../../../contracts/index.ts";
import type { EpisodeRepository } from "../application/ports/episode-repository.ts";
import { getAudio } from "../application/use-cases/get-audio.ts";
import { listEpisodes } from "../application/use-cases/list-episodes.ts";
import type { GetAudioController } from "../controllers/get-audio-controller.ts";
import { createGetAudioController } from "../controllers/get-audio-controller.ts";
import type { ListEpisodesController } from "../controllers/list-episodes-controller.ts";
import { createListEpisodesController } from "../controllers/list-episodes-controller.ts";
import { GoogleDriveEpisodeRepository } from "../infrastructure/drive/google-drive-episode-repository.ts";
import { InMemoryEpisodeRepository } from "../infrastructure/drive/in-memory-episode-repository.ts";
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
};

/**
 * local development / unit test 用に use case 一式を丸ごと差し替える override。
 *
 * @invariant repository 解決（`createEpisodeRepository`）を経由しない。env の Drive 設定不足を無視する
 */
export type PlaybackUseCaseOverrides = {
  useCases: {
    listEpisodes: () => Promise<ListEpisodesResponse>;
    getAudio: (episodeId: string) => Promise<Uint8Array>;
  };
};

/**
 * env から `EpisodeRepository` を選ぶ結果。
 */
export type EpisodeRepositorySelection =
  | { kind: "drive"; repository: EpisodeRepository }
  | { kind: "in-memory"; repository: EpisodeRepository };

/**
 * env から `EpisodeRepository` を選ぶ。
 *
 * @require env は Cloudflare Workers native secrets/vars（`.env` は読まない）
 * @ensure OAuth 値と DRIVE_FOLDER_ID が全て揃う時は "drive"、明示的 local / unit test mode の時は
 *   "in-memory"。設定不足は runtime config module が throw する
 */
export function createEpisodeRepository(
  env: PlaybackEnv,
  options: PlaybackRepositoryOptions = {},
): EpisodeRepositorySelection {
  const validated = validatePlaybackEnv(env, options);

  if (validated.mode === "drive") {
    const {
      GOOGLE_OAUTH_CLIENT_ID: clientId,
      GOOGLE_OAUTH_CLIENT_SECRET: clientSecret,
      GOOGLE_OAUTH_REFRESH_TOKEN: refreshToken,
      DRIVE_FOLDER_ID: folderId,
    } = validated.env;

    return {
      kind: "drive",
      // why: Workers 実行時の global fetch は this 束縛が不要（narrow integration が globalThis.fetch
      //   直渡しで疎通確認済み）。ラッパ関数を挟むと、その 1 行が unit で到達不能な dead branch になる
      repository: new GoogleDriveEpisodeRepository({
        fetch,
        oauth: { clientId, clientSecret, refreshToken },
        folderId,
      }),
    };
  }

  return { kind: "in-memory", repository: new InMemoryEpisodeRepository() };
}

/**
 * env から Playback worker の Controller 一式を組み立てる。
 *
 * @require env は Cloudflare Workers native secrets/vars
 * @ensure useCaseOverrides がある時は repository 解決を経由せず、渡された use case を Controller
 *   へ直結する。無い時は従来通り repository を選べれば Controller 一式を返し、設定不足は throw する
 * @invariant useCaseOverrides は既存の Drive / in-memory 分岐（`createEpisodeRepository`）を変更しない
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
    };
  }

  const selection = createEpisodeRepository(env, options);

  const { repository } = selection;
  return {
    listEpisodesController: createListEpisodesController(() => listEpisodes(repository)),
    getAudioController: createGetAudioController((episodeId) => getAudio(repository, episodeId)),
  };
}
