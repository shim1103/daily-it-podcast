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
import {
  validatePlaybackEnv,
  type PlaybackEnv,
  type PlaybackRepositoryOptions,
} from "./runtime-config.ts";

export type {
  PlaybackEnv,
  PlaybackRepositoryMode,
  PlaybackRepositoryOptions,
} from "./runtime-config.ts";
export { PlaybackRuntimeConfigError } from "./runtime-config-error.ts";

export type PlaybackControllers = {
  listEpisodesController: ListEpisodesController;
  getEpisodeController: GetEpisodeController;
  getEpisodeAudioController: GetEpisodeAudioController;
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
 * @ensure repository を選べる時は "ready"（Controller 一式）を返す。設定不足は throw する
 */
export function createPlaybackControllers(
  env: PlaybackEnv,
  options: PlaybackRepositoryOptions = {},
): PlaybackControllers {
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
