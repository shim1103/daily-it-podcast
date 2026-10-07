import type {
  ProgressWriteUseCaseInput,
  ProgressWriteUseCaseOutput,
} from "../application/progress/progress-write-use-case-io.ts";
import type {
  EpisodeWithProgress,
  ListEpisodesUseCaseOutput,
} from "../application/use-cases/list-episodes.ts";
import type {
  GetAudioUseCaseInput,
  GetAudioUseCaseOutput,
} from "../application/use-cases/get-audio.ts";
import type {
  PullProgressUseCaseInput,
  PullProgressUseCaseOutput,
} from "../application/use-cases/pull-progress.ts";
import { EpisodeContentError } from "../entities/errors/episode-content-error.ts";
import type { EpisodeManuscript } from "../entities/models/episode-manuscript.ts";
import { createFakeEpisodeAudioBytes } from "../test/fixtures/audio-bytes.ts";
import fakeEpisodesJson from "./fake-episodes.json" with { type: "json" };

type FakeEpisodeRecord = {
  episodeId: string;
  date: string;
  title: string;
  durationSec: number;
  body: EpisodeManuscript["body"];
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

function toEpisodeWithProgress(record: FakeEpisodeRecord): EpisodeWithProgress {
  return {
    episodeId: record.episodeId,
    date: record.date,
    title: record.title,
    durationSec: record.durationSec,
    body: record.body,
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

export const validListEpisodesUseCaseOutput: ListEpisodesUseCaseOutput = {
  episodes: fakeEpisodes.map(toEpisodeWithProgress),
};

export const validEpisodeWithProgress: EpisodeWithProgress = toEpisodeWithProgress(
  findFakeEpisode("ep-1"),
);

export function createFakeListEpisodesUseCase(
  impl?: () => Promise<ListEpisodesUseCaseOutput>,
): () => Promise<ListEpisodesUseCaseOutput> {
  return impl ?? (async () => validListEpisodesUseCaseOutput);
}

export function createFakeGetAudioUseCase(
  impl?: (input: GetAudioUseCaseInput) => Promise<GetAudioUseCaseOutput>,
): (input: GetAudioUseCaseInput) => Promise<GetAudioUseCaseOutput> {
  return (
    impl ??
    (async ({ episodeId }) => {
      return loadFakeEpisodeAudio(episodeId);
    })
  );
}

/** A 足場・Controller test 用の Write 成功出力。 */
export const validProgressWriteUseCaseOutput: ProgressWriteUseCaseOutput = {
  firstPlayedAt: "2026-09-22T10:00:00.000Z",
  firstCompletedAt: null,
};

/** A 足場・Controller test 用の pull 成功出力（差分なし）。 */
export const validPullProgressUseCaseOutput: PullProgressUseCaseOutput = {
  episodes: [],
};

export function createFakeProgressWriteUseCase(
  impl?: (input: ProgressWriteUseCaseInput) => Promise<ProgressWriteUseCaseOutput>,
): (input: ProgressWriteUseCaseInput) => Promise<ProgressWriteUseCaseOutput> {
  return impl ?? (async () => validProgressWriteUseCaseOutput);
}

export function createFakePullProgressUseCase(
  impl?: (input: PullProgressUseCaseInput) => Promise<PullProgressUseCaseOutput>,
): (input: PullProgressUseCaseInput) => Promise<PullProgressUseCaseOutput> {
  return impl ?? (async () => validPullProgressUseCaseOutput);
}
