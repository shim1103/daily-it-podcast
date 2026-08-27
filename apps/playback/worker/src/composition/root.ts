import type { GetEpisodeResponse, ListEpisodesResponse } from "../../../contracts/index.ts";
import type { EpisodeRepository } from "../application/ports/episode-repository.ts";
import { getEpisode } from "../application/use-cases/get-episode.ts";
import { getEpisodeAudio } from "../application/use-cases/get-episode-audio.ts";
import { listEpisodes } from "../application/use-cases/list-episodes.ts";
import type { GetEpisodeAudioController } from "../controllers/get-episode-audio-controller.ts";
import { createGetEpisodeAudioController } from "../controllers/get-episode-audio-controller.ts";
import type { GetEpisodeController } from "../controllers/get-episode-controller.ts";
import { createGetEpisodeController } from "../controllers/get-episode-controller.ts";
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
  getEpisodeController: GetEpisodeController;
  getEpisodeAudioController: GetEpisodeAudioController;
};

/**
 * local development / unit test 用に use case 一式を丸ごと差し替える override。
 *
 * @invariant repository 解決（`createEpisodeRepository`）を経由しない。env の Drive 設定不足を無視する
 */
export type PlaybackUseCaseOverrides = {
  useCases: {
    listEpisodes: () => Promise<ListEpisodesResponse>;
    getEpisode: (episodeId: string) => Promise<GetEpisodeResponse>;
    getEpisodeAudio: (episodeId: string) => Promise<Uint8Array>;
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
      repository: new GoogleDriveEpisodeRepository({
        fetch: (input, init) => fetch(input, init),
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
      getEpisodeController: createGetEpisodeController(useCases.getEpisode),
      getEpisodeAudioController: createGetEpisodeAudioController(useCases.getEpisodeAudio),
    };
  }

  const selection = createEpisodeRepository(env, options);

  const { repository } = selection;
  return {
    listEpisodesController: createListEpisodesController(() => listEpisodes(repository)),
    getEpisodeController: createGetEpisodeController((episodeId) =>
      getEpisode(repository, episodeId),
    ),
    getEpisodeAudioController: createGetEpisodeAudioController((episodeId) =>
      getEpisodeAudio(repository, episodeId),
    ),
  };
}
