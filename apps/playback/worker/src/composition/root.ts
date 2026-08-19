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

/**
 * Cloudflare Workers native secrets/vars から受け取る、Drive 接続に必要な env の形。
 *
 * why: 本番 Adapter（GoogleDriveEpisodeRepository）を選択するには OAuth 値と
 * DRIVE_FOLDER_ID が要る。全て揃う時は本番 Adapter、全て未設定の時はローカル開発・test 用の
 * 意図的な Fake 利用とみなす。一部だけ設定されている状態は設定漏れとみなし区別する
 * （`createEpisodeRepository` を参照）。
 */
export type PlaybackEnv = {
  GOOGLE_OAUTH_CLIENT_ID?: string;
  GOOGLE_OAUTH_CLIENT_SECRET?: string;
  GOOGLE_OAUTH_REFRESH_TOKEN?: string;
  DRIVE_FOLDER_ID?: string;
};

type PlaybackControllers = {
  listEpisodesController: ListEpisodesController;
  getEpisodeController: GetEpisodeController;
  getEpisodeAudioController: GetEpisodeAudioController;
};

/**
 * env から `EpisodeRepository` を選ぶ判定結果。
 *
 * why: OAuth 値の一部だけが欠ける状態（本番相当の環境での設定漏れ）と、
 * 全て空の状態（ローカル開発・test で意図的に Fake を使う状態）を区別して呼び出し元へ渡す。
 * 判定結果を返すだけに留め、判定自体の合否判断（throw するか等）は呼び出し元に委ねる。
 */
export type EpisodeRepositorySelection =
  | { kind: "drive"; repository: EpisodeRepository }
  | { kind: "in-memory"; repository: EpisodeRepository }
  | { kind: "misconfigured"; missing: readonly string[] };

export type PlaybackControllersResult =
  | { kind: "ready"; controllers: PlaybackControllers }
  | { kind: "misconfigured"; missing: readonly string[] };

const driveEnvKeys = [
  "GOOGLE_OAUTH_CLIENT_ID",
  "GOOGLE_OAUTH_CLIENT_SECRET",
  "GOOGLE_OAUTH_REFRESH_TOKEN",
  "DRIVE_FOLDER_ID",
] as const satisfies readonly (keyof PlaybackEnv)[];

/**
 * env から `EpisodeRepository` を選ぶ。
 *
 * @require env は Cloudflare Workers native secrets/vars（`.env` は読まない）
 * @ensure OAuth 値と DRIVE_FOLDER_ID が全て揃う時は "drive"、全て未設定の時は "in-memory"、
 *   一部だけ設定されている時は "misconfigured"（missing に欠落 key 名一覧）を返す
 */
export function createEpisodeRepository(env: PlaybackEnv): EpisodeRepositorySelection {
  const missing = driveEnvKeys.filter((key) => env[key] === undefined);

  if (missing.length === 0) {
    const {
      GOOGLE_OAUTH_CLIENT_ID: clientId,
      GOOGLE_OAUTH_CLIENT_SECRET: clientSecret,
      GOOGLE_OAUTH_REFRESH_TOKEN: refreshToken,
      DRIVE_FOLDER_ID: folderId,
    } = env as Required<PlaybackEnv>;

    return {
      kind: "drive",
      repository: new GoogleDriveEpisodeRepository({
        fetch: (input, init) => fetch(input, init),
        oauth: { clientId, clientSecret, refreshToken },
        folderId,
      }),
    };
  }

  if (missing.length === driveEnvKeys.length) {
    return { kind: "in-memory", repository: new InMemoryEpisodeRepository() };
  }

  return { kind: "misconfigured", missing };
}

/**
 * env から Playback worker の Controller 一式を組み立てる。
 *
 * @require env は Cloudflare Workers native secrets/vars
 * @ensure repository 選定が "drive" または "in-memory" の時は "ready"（Controller 一式）を返す。
 *   "misconfigured" の時は Controller を組み立てず "misconfigured"（missing に欠落 key 名一覧）を返す
 */
export function createPlaybackControllers(env: PlaybackEnv): PlaybackControllersResult {
  const selection = createEpisodeRepository(env);

  if (selection.kind === "misconfigured") {
    return { kind: "misconfigured", missing: selection.missing };
  }

  const { repository } = selection;
  return {
    kind: "ready",
    controllers: {
      listEpisodesController: createListEpisodesController(() =>
        listEpisodes(repository),
      ),
      getEpisodeController: createGetEpisodeController((episodeId) =>
        getEpisode(repository, episodeId),
      ),
      getEpisodeAudioController: createGetEpisodeAudioController((episodeId) =>
        getEpisodeAudio(repository, episodeId),
      ),
    },
  };
}
