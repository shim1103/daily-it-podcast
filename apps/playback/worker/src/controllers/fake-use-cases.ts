import {
  episodeAudioPath,
  type GetEpisodeResponse,
  type ListEpisodesResponse,
} from "../../../contracts/index.ts";

export const validListEpisodesResponse: ListEpisodesResponse = {
  episodes: [
    {
      episodeId: "ep-1",
      date: "2026-08-17",
      title: "題",
      durationSec: 60,
    },
  ],
};

export const validGetEpisodeResponse: GetEpisodeResponse = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [
      {
        title: "題",
        preface: "前置き",
        detail: "詳細",
        startSec: 0,
      },
    ],
    closing: "終了",
  },
  audioRef: episodeAudioPath("ep-1"),
};

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
  return impl ?? (async () => validGetEpisodeResponse);
}

export function createFakeGetEpisodeAudioUseCase(
  impl?: (episodeId: string) => Promise<Uint8Array>,
): (episodeId: string) => Promise<Uint8Array> {
  return impl ?? (async () => validAudioBytes);
}
