import {
  type EpisodeItem,
  episodeAudioPath,
  type ListEpisodesResponse,
} from "../../../contracts/index.ts";
import type {
  EpisodeWithProgress,
  ListEpisodesUseCaseOutput,
} from "../application/use-cases/list-episodes.ts";
import { mapInternalErrorToExternal } from "./map-internal-error.ts";

export type ListEpisodesUseCase = () => Promise<ListEpisodesUseCaseOutput>;

export type ListEpisodesController = () => Promise<ListEpisodesResponse>;

function toEpisodeItem(episode: EpisodeWithProgress): EpisodeItem {
  return {
    episodeId: episode.episodeId,
    date: episode.date,
    title: episode.title,
    durationSec: episode.durationSec,
    body: {
      opening: {
        text: episode.body.opening.text,
        startSec: episode.body.opening.startSec,
      },
      topics: episode.body.topics.map((topic) => ({
        title: topic.title,
        preface: topic.preface,
        detail: topic.detail,
        startSec: topic.startSec,
      })),
      ending: {
        text: episode.body.ending.text,
        startSec: episode.body.ending.startSec,
      },
    },
    audioRef: episodeAudioPath(episode.episodeId),
    progress:
      episode.progress === null
        ? null
        : {
            positionSec: episode.progress.positionSec,
            firstPlayedAt: episode.progress.firstPlayedAt,
            firstCompletedAt: episode.progress.firstCompletedAt,
            lastPlayedAt: episode.progress.lastPlayedAt,
          },
  };
}

function toListEpisodesResponse(output: ListEpisodesUseCaseOutput): ListEpisodesResponse {
  return { episodes: output.episodes.map(toEpisodeItem) };
}

/**
 * 一覧 JSON を返す Controller を組み立てる。
 *
 * @require useCase は原稿に進捗を足した一覧の ListEpisodesUseCaseOutput を返す
 * @ensure 戻り関数は契約の ListEpisodesResponse を返し、各 episode へ音声の path を足す。Internal は External に変換して throw する
 * @invariant HTTP status と Response object を作らない
 */
export function createListEpisodesController(useCase: ListEpisodesUseCase): ListEpisodesController {
  return async function listEpisodesController(): Promise<ListEpisodesResponse> {
    try {
      return toListEpisodesResponse(await useCase());
    } catch (error) {
      throw mapInternalErrorToExternal(error);
    }
  };
}
