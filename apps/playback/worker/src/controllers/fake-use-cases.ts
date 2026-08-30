import {
  episodeAudioPath,
  type GetEpisodeResponse,
  type ListEpisodesResponse,
} from "../../../contracts/index.ts";
import { createFakeEpisodeAudioBytes } from "../test/fixtures/audio-bytes.ts";
import { EpisodeContentError } from "../entities/errors/episode-content-error.ts";
import fakeEpisodesJson from "./fake-episodes.json" with { type: "json" };

type FakeEpisodeRecord = {
  episodeId: string;
  date: string;
  title: string;
  durationSec: number;
  body: GetEpisodeResponse["body"];
};

const fakeEpisodes = fakeEpisodesJson as FakeEpisodeRecord[];
const fakeAudioCache = new Map<string, Uint8Array>();

function loadFakeEpisodeAudio(episodeId: string): Uint8Array {
  const cached = fakeAudioCache.get(episodeId);
  if (cached !== undefined) {
    return cached;
  }
  const record = findFakeEpisode(episodeId);
  const bytes = createFakeEpisodeAudioBytes(record.durationSec);
  fakeAudioCache.set(episodeId, bytes);
  return bytes;
}

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
    throw new EpisodeContentError(`JSON エントリが無い: ${episodeId}`);
  }
  return found;
}

export const validListEpisodesResponse: ListEpisodesResponse = {
  episodes: fakeEpisodes.map(({ episodeId, date, title, durationSec, body }) => ({
    episodeId,
    date,
    title,
    durationSec,
    topics: body.topics.map((topic) => ({ title: topic.title })),
    audioRef: episodeAudioPath(episodeId),
  })),
};

export const validGetEpisodeResponse: GetEpisodeResponse = toGetEpisodeResponse(
  findFakeEpisode("ep-1"),
);

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
  return (
    impl ??
    (async (episodeId) => {
      return loadFakeEpisodeAudio(episodeId);
    })
  );
}
