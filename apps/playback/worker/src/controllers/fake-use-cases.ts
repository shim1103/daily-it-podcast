import {
  episodeAudioPath,
  type GetEpisodeResponse,
  type ListEpisodesResponse,
} from "../../../contracts/index.ts";
import { EpisodeNotFoundError } from "../entities/errors/episode-not-found-error.ts";
import fakeEpisodesJson from "./fake-episodes.json" with { type: "json" };

type FakeEpisodeRecord = {
  episodeId: string;
  date: string;
  title: string;
  durationSec: number;
  body: GetEpisodeResponse["body"];
};

const fakeEpisodes = fakeEpisodesJson as FakeEpisodeRecord[];

function toGetEpisodeResponse(record: FakeEpisodeRecord): GetEpisodeResponse {
  return {
    episodeId: record.episodeId,
    date: record.date,
    title: record.title,
    durationSec: record.durationSec,
    body: record.body,
    audioRef: episodeAudioPath(record.episodeId),
  };
}

function findFakeEpisode(episodeId: string): FakeEpisodeRecord {
  const found = fakeEpisodes.find((episode) => episode.episodeId === episodeId);
  if (!found) {
    throw new EpisodeNotFoundError(`JSON エントリが無い: ${episodeId}`);
  }
  return found;
}

export const validListEpisodesResponse: ListEpisodesResponse = {
  episodes: fakeEpisodes.map(({ episodeId, date, title, durationSec }) => ({
    episodeId,
    date,
    title,
    durationSec,
  })),
};

export const validGetEpisodeResponse: GetEpisodeResponse = toGetEpisodeResponse(
  findFakeEpisode("ep-1"),
);

export const validAudioBytes = new Uint8Array([
  0x52, 0x49, 0x46, 0x46, 0x04, 0x00, 0x00, 0x00, 0x57, 0x41, 0x56, 0x45,
]);

export function createFakeListEpisodesUseCase(
  impl?: () => Promise<ListEpisodesResponse>,
): () => Promise<ListEpisodesResponse> {
  return impl ?? (async () => validListEpisodesResponse);
}

export function createFakeGetEpisodeUseCase(
  impl?: (episodeId: string) => Promise<GetEpisodeResponse>,
): (episodeId: string) => Promise<GetEpisodeResponse> {
  return (
    impl ??
    (async (episodeId) => {
      return toGetEpisodeResponse(findFakeEpisode(episodeId));
    })
  );
}

export function createFakeGetEpisodeAudioUseCase(
  impl?: (episodeId: string) => Promise<Uint8Array>,
): (episodeId: string) => Promise<Uint8Array> {
  return impl ?? (async () => validAudioBytes);
}
