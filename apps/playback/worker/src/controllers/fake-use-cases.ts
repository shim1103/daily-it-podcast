import {
  episodeAudioPath,
  type EpisodeItem,
  type ListEpisodesResponse,
  type ProgressPullResponse,
  type ProgressWriteRequest,
  type ProgressWriteResponse,
} from "../../../contracts/index.ts";
import { createFakeEpisodeAudioBytes } from "../test/fixtures/audio-bytes.ts";
import { EpisodeContentError } from "../entities/errors/episode-content-error.ts";
import fakeEpisodesJson from "./fake-episodes.json" with { type: "json" };

type FakeEpisodeRecord = {
  episodeId: string;
  date: string;
  title: string;
  durationSec: number;
  body: EpisodeItem["body"];
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

function toEpisodeItem(record: FakeEpisodeRecord): EpisodeItem {
  return {
    episodeId: record.episodeId,
    date: record.date,
    title: record.title,
    durationSec: record.durationSec,
    body: record.body,
    audioRef: episodeAudioPath(record.episodeId),
    progress: null,
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
  episodes: fakeEpisodes.map(toEpisodeItem),
};

export const validEpisodeItem: EpisodeItem = toEpisodeItem(findFakeEpisode("ep-1"));

export function createFakeListEpisodesUseCase(
  impl?: () => Promise<ListEpisodesResponse>,
): () => Promise<ListEpisodesResponse> {
  return impl ?? (async () => validListEpisodesResponse);
}

export function createFakeGetAudioUseCase(
  impl?: (episodeId: string) => Promise<Uint8Array>,
): (episodeId: string) => Promise<Uint8Array> {
  return (
    impl ??
    (async (episodeId) => {
      return loadFakeEpisodeAudio(episodeId);
    })
  );
}

/** A 足場・Controller test 用の Write 成功応答。 */
export const validProgressWriteResponse: ProgressWriteResponse = {
  firstPlayedAt: "2026-09-22T10:00:00.000Z",
  firstCompletedAt: null,
};

/** A 足場・Controller test 用の pull 成功応答（差分なし）。 */
export const validProgressPullResponse: ProgressPullResponse = {
  episodes: [],
};

export function createFakeProgressWriteUseCase(
  impl?: (episodeId: string, body: ProgressWriteRequest) => Promise<ProgressWriteResponse>,
): (episodeId: string, body: ProgressWriteRequest) => Promise<ProgressWriteResponse> {
  return impl ?? (async () => validProgressWriteResponse);
}

export function createFakePullProgressUseCase(
  impl?: (since: string) => Promise<ProgressPullResponse>,
): (since: string) => Promise<ProgressPullResponse> {
  return impl ?? (async () => validProgressPullResponse);
}
