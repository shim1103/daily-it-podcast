import { z } from "zod";

export const playbackHttpErrorCodes = [
  "episode_not_found",
  "validation_error",
  "configuration_error",
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

const openingSchema = z.strictObject({
  text: z.string(),
  startSec: z.number().min(0),
});

const closingSchema = z.strictObject({
  summary: z.string(),
  startSec: z.number().min(0),
});

const bodySchema = z.strictObject({
  opening: openingSchema,
  topics: z.array(topicSchema).min(1),
  closing: closingSchema,
});

export const episodeItemSchema = z.strictObject({
  episodeId: episodeIdSchema,
  date: dateSchema,
  title: titleSchema,
  durationSec: durationSecSchema,
  body: bodySchema,
  audioRef: z.string().min(1),
});

export const listEpisodesPath = "/episodes" as const;

/** Hono route template。音声 GET の path パラメータ付き（`:episodeId` 段）。 */
export const episodeRoutePath = `${listEpisodesPath}/:episodeId` as const;

/** Hono route template。音声 GET の path パラメータ付き。 */
export const episodeAudioRoutePath = `${episodeRoutePath}/audio` as const;

/**
 * episodeId を含む path 段。音声 GET の親 path として使う。
 *
 * @require episodeId は空でない
 * @ensure listEpisodesPath の後に path 段が 1 つだけ増える
 */
export function episodePath(episodeId: string): string {
  return `${listEpisodesPath}/${encodeURIComponent(episodeId)}`;
}

/**
 * 音声 GET の path。成功時の body は `audio/wav` のバイト列であり JSON ではない。
 *
 * @require episodeId は空でない
 * @ensure episodePath の後に `audio` 段が 1 つ続く
 */
export function episodeAudioPath(episodeId: string): string {
  return `${episodePath(episodeId)}/audio`;
}

/** 音声 GET 成功時の `Content-Type`。Drive 上の `{episodeId}.wav` に対応する。 */
export const episodeAudioContentType = "audio/wav";

export const ListEpisodesResponseSchema = z.strictObject({
  episodes: z.array(episodeItemSchema),
});

/** 音声 GET 等、path パラメータ episodeId の入力契約。 */
export const EpisodeIdRequestSchema = z.strictObject({
  episodeId: episodeIdSchema,
});

export const ErrorResponseSchema = z.strictObject({
  code: z.enum(playbackHttpErrorCodes),
});

export type EpisodeItem = z.infer<typeof episodeItemSchema>;
export type ListEpisodesResponse = z.infer<typeof ListEpisodesResponseSchema>;
export type EpisodeIdRequest = z.infer<typeof EpisodeIdRequestSchema>;
export type ErrorResponse = z.infer<typeof ErrorResponseSchema>;
