import { z } from "zod";

export const playbackHttpErrorCodes = [
  "episode_not_found",
  "validation_error",
  "unavailable",
] as const;

export type PlaybackHttpErrorCode = (typeof playbackHttpErrorCodes)[number];

const episodeIdSchema = z.string().min(1);
const dateSchema = z.string().regex(/^[0-9]{4}-[0-9]{2}-[0-9]{2}$/);
const titleSchema = z.string().min(1);
const durationSecSchema = z.number().min(0);

const topicSchema = z.strictObject({
  title: titleSchema,
  preface: z.string(),
  detail: z.string(),
  startSec: z.number().min(0),
});

const bodySchema = z.strictObject({
  opening: z.string(),
  topics: z.array(topicSchema).min(1),
  closing: z.string(),
});

const episodeListItemSchema = z.strictObject({
  episodeId: episodeIdSchema,
  date: dateSchema,
  title: titleSchema,
  durationSec: durationSecSchema,
});

export const listEpisodesPath = "/episodes";

/**
 * 1件 JSON の HTTP path。音声バイトの path ではない。
 *
 * @require episodeId は空でない（GetEpisodeRequest.episodeId）
 * @ensure listEpisodesPath の後に path 段が 1 つだけ増える
 */
export function episodePath(episodeId: string): string {
  return `${listEpisodesPath}/${encodeURIComponent(episodeId)}`;
}

/**
 * 音声 GET の path。成功時の body は `audio/mpeg` のバイト列であり JSON ではない。
 *
 * @require episodeId は空でない（GetEpisodeRequest.episodeId）
 * @ensure episodePath の後に `audio` 段が 1 つ続く
 */
export function episodeAudioPath(episodeId: string): string {
  return `${episodePath(episodeId)}/audio`;
}

export const ListEpisodesResponseSchema = z.strictObject({
  episodes: z.array(episodeListItemSchema),
});

export const GetEpisodeRequestSchema = z.strictObject({
  episodeId: episodeIdSchema,
});

export const GetEpisodeResponseSchema = z.strictObject({
  episodeId: episodeIdSchema,
  date: dateSchema,
  title: titleSchema,
  durationSec: durationSecSchema,
  body: bodySchema,
  audioRef: z.string().min(1),
});

export const ErrorResponseSchema = z.strictObject({
  code: z.enum(playbackHttpErrorCodes),
});

export type ListEpisodesResponse = z.infer<typeof ListEpisodesResponseSchema>;
export type GetEpisodeRequest = z.infer<typeof GetEpisodeRequestSchema>;
export type GetEpisodeResponse = z.infer<typeof GetEpisodeResponseSchema>;
export type ErrorResponse = z.infer<typeof ErrorResponseSchema>;
