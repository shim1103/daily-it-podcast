import { z } from "zod";

export const playbackHttpErrorCodes = [
  "episode_not_found",
  "validation_error",
  "configuration_error",
  "unavailable",
] as const;

export type PlaybackHttpErrorCode = (typeof playbackHttpErrorCodes)[number];

/**
 * 進捗 Write（create / update / complete）の最大試行回数（初回含む）。
 * browser の有限 retry 上限。session またぎキューは持たない。
 */
export const PROGRESS_WRITE_MAX_ATTEMPTS = 3 as const;

/**
 * Write を HTTP 契約 code として再試行してよいもの。
 * validation / not_found は再送しても結果が変わらないので含めない。
 */
export const progressWriteRetryableHttpErrorCodes = [
  "unavailable",
] as const satisfies readonly PlaybackHttpErrorCode[];

const progressInstantLowerExclusiveMs = Date.UTC(1970, 0, 1);
const progressInstantUpperExclusiveMs = Date.UTC(2100, 0, 1);

/**
 * 進捗の時刻（`clientAt`・`since`・`first*`・`lastPlayedAt`）が共有する schema。
 * 入力は offset 付き（`Z` または `±hh:mm`）の ISO-8601 を受け、instant が範囲外なら拒否する。
 * 出力は同じ instant の UTC 固定幅（`YYYY-MM-DDTHH:mm:ss.sssZ`）で、文字列順が時系列順と一致する。
 */
const progressInstantSchema = z.iso
  .datetime({ offset: true })
  .refine((value) => {
    const instantMs = Date.parse(value);
    return (
      instantMs > progressInstantLowerExclusiveMs && instantMs < progressInstantUpperExclusiveMs
    );
  })
  .transform((value) => new Date(value).toISOString());

/**
 * 進捗 pull の cursor。10 進の非負整数の文字列（15 桁以下、先頭 0 は `"0"` のみ）。
 * 形式の正本はこの schema で、形式外は Route の検証が 400 にする。桁数の上限は、整数へ変換しても精度が落ちない範囲に収まる。
 * client は中身を解釈せず、応答で受け取った値をそのまま次回の要求へ渡す。文字列から整数への変換は route の zod が行う設計
 * （Decision 2026-10-01T23-32-40）だが、この schema はまだ変換せず、文字列の形式だけを検証する。
 */
const progressCursorSchema = z.string().regex(/^(0|[1-9][0-9]{0,14})$/);

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

const endingSchema = z.strictObject({
  text: z.string(),
  startSec: z.number().min(0),
});

const bodySchema = z.strictObject({
  opening: openingSchema,
  topics: z.array(topicSchema).min(1),
  ending: endingSchema,
});

/**
 * list embed / 永続行が共有する再生進捗。未playは episode 側で `progress: null`（行なし）。
 * 行があるとき `firstPlayedAt` / `lastPlayedAt` は必須。`firstCompletedAt` のみ未完走で null。
 * 時刻の形式・範囲・正規化は {@link progressInstantSchema} に従う（parse 後は UTC 固定幅）。
 */
export const episodeProgressSchema = z.strictObject({
  positionSec: z.number().min(0),
  firstPlayedAt: progressInstantSchema,
  firstCompletedAt: progressInstantSchema.nullable(),
  lastPlayedAt: progressInstantSchema,
});

export const episodeItemSchema = z.strictObject({
  episodeId: episodeIdSchema,
  date: dateSchema,
  title: titleSchema,
  durationSec: durationSecSchema,
  body: bodySchema,
  audioRef: z.string().min(1),
  progress: episodeProgressSchema.nullable(),
});

export const listEpisodesPath = "/episodes" as const;

/** Hono route template。音声 GET の path パラメータ付き（`:episodeId` 段）。 */
export const episodeRoutePath = `${listEpisodesPath}/:episodeId` as const;

/** Hono route template。音声 GET の path パラメータ付き。 */
export const episodeAudioRoutePath = `${episodeRoutePath}/audio` as const;

/** Hono route template。進捗 create/update の path パラメータ付き。 */
export const episodeProgressRoutePath = `${episodeRoutePath}/progress` as const;

/** Hono route template。進捗 complete の path パラメータ付き。 */
export const episodeProgressCompleteRoutePath = `${episodeProgressRoutePath}/complete` as const;

/**
 * session 中の進捗 pull（差分 Get）。list embed とは別 URL。
 * query の形は {@link ProgressPullQuerySchema}。
 */
export const progressPullPath = "/progress" as const;

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
 * 音声 GET の path。成功時の body は `audio/mpeg` のバイト列であり JSON ではない。
 *
 * @require episodeId は空でない
 * @ensure episodePath の後に `audio` 段が 1 つ続く
 */
export function episodeAudioPath(episodeId: string): string {
  return `${episodePath(episodeId)}/audio`;
}

/**
 * 進捗 create（POST）/ update（PATCH）の path。
 *
 * @require episodeId は空でない
 * @ensure episodePath の後に `progress` 段が 1 つ続く
 */
export function episodeProgressPath(episodeId: string): string {
  return `${episodePath(episodeId)}/progress`;
}

/**
 * 進捗 complete（POST）の path。
 *
 * @require episodeId は空でない
 * @ensure episodeProgressPath の後に `complete` 段が 1 つ続く
 */
export function episodeProgressCompletePath(episodeId: string): string {
  return `${episodeProgressPath(episodeId)}/complete`;
}

/** Drive 上の音声 file 拡張子。`{episodeId}.mp3` に対応する。 */
export const episodeAudioFileExtension = ".mp3";

/** 音声 GET 成功時の `Content-Type`。Drive 上の `{episodeId}.mp3` に対応する。 */
export const episodeAudioContentType = "audio/mpeg";

/** episodes は date 降順（新しい日付が先頭）で返る契約。並び替えは worker 側 use-case が持つ。 */
export const ListEpisodesResponseSchema = z.strictObject({
  episodes: z.array(episodeItemSchema),
});

/** 音声 GET 等、path パラメータ episodeId の入力契約。 */
export const EpisodeIdRequestSchema = z.strictObject({
  episodeId: episodeIdSchema,
});

/**
 * 進捗 create（POST）/ update（PATCH）/ complete（POST .../complete）共通の JSON body。
 * 操作の違いは HTTP method と path で表す。字段が同じため schema は共有する。
 * `clientAt` は merge 判定用の client 時刻。形式・範囲・正規化は {@link progressInstantSchema} に従い、
 * 入力は offset 付きでも受け、parse 後は UTC 固定幅になる。
 * 行なし update → 404、重複 create → 冪等 200 の意味は Decision を正とし、ここへ写さない。
 */
export const ProgressWriteRequestSchema = z.strictObject({
  positionSec: z.number().min(0),
  clientAt: progressInstantSchema,
});

/**
 * 進捗 Write 成功応答。勝ち側の first* だけを返す（`positionSec` は載せない）。
 * 時刻は {@link progressInstantSchema} に従う。
 */
export const ProgressWriteResponseSchema = z.strictObject({
  firstPlayedAt: progressInstantSchema,
  firstCompletedAt: progressInstantSchema.nullable(),
});

/**
 * DB が空の初回 pull が返す cursor。「どの行より前」を指す起点で、
 * cursor を省略した要求は、route の zod が既定値としてこの値にする設計（Decision 2026-10-01T23-32-40）。
 * cursor 版へ切り替えるまで、既定値の設定は未実装。
 */
export const PROGRESS_PULL_ORIGIN_CURSOR = "0" as const;

/** 進捗 pull 応答の 1 件。更新があった episode の進捗（`progress` は常に object）。 */
const progressPullEntrySchema = z.strictObject({
  episodeId: episodeIdSchema,
  progress: episodeProgressSchema,
});

// todo: C が route と UseCase を ProgressCursorPull* へ切り替えたら、この schema と型を削除する
/**
 * 進捗 pull の query。`since` より後に更新された行だけを返す。
 * `since` は {@link progressInstantSchema} に従い、parse 後は UTC 固定幅になる。
 */
export const ProgressPullQuerySchema = z.strictObject({
  since: progressInstantSchema,
});

// todo: C が route と UseCase を ProgressCursorPull* へ切り替えたら、この schema と型を削除する
/** 進捗 pull 応答。更新があった episode だけ（`progress` は常に object）。 */
export const ProgressPullResponseSchema = z.strictObject({
  episodes: z.array(progressPullEntrySchema),
});

/**
 * 進捗 pull（cursor 版）の query。`cursor` を省略すると全件を返す。
 * 以後は応答の `cursor` をそのまま渡す。形式は {@link progressCursorSchema} に従う。
 *
 * @require cursor は省略するか、応答で受け取った値をそのまま渡す
 * @ensure cursor の形式が不正なら 400（`validation_error`）
 */
export const ProgressCursorPullQuerySchema = z.strictObject({
  cursor: progressCursorSchema.optional(),
});

/**
 * 進捗 pull（cursor 版）の応答。`cursor` は最上位にだけ出し、行には載せない。
 *
 * @ensure 変更が無いときは空の `episodes` と、要求と同じ cursor を返す
 * @ensure DB が空の初回は {@link PROGRESS_PULL_ORIGIN_CURSOR} を返す
 */
export const ProgressCursorPullResponseSchema = z.strictObject({
  cursor: progressCursorSchema,
  episodes: z.array(progressPullEntrySchema),
});

export const ErrorResponseSchema = z.strictObject({
  code: z.enum(playbackHttpErrorCodes),
});

export type EpisodeProgress = z.infer<typeof episodeProgressSchema>;
export type EpisodeItem = z.infer<typeof episodeItemSchema>;
export type ListEpisodesResponse = z.infer<typeof ListEpisodesResponseSchema>;
export type EpisodeIdRequest = z.infer<typeof EpisodeIdRequestSchema>;
export type ProgressWriteRequest = z.infer<typeof ProgressWriteRequestSchema>;
export type ProgressWriteResponse = z.infer<typeof ProgressWriteResponseSchema>;
// todo: C が route と UseCase を ProgressCursorPull* へ切り替えたら、この型を削除する
export type ProgressPullQuery = z.infer<typeof ProgressPullQuerySchema>;
// todo: C が route と UseCase を ProgressCursorPull* へ切り替えたら、この型を削除する
export type ProgressPullResponse = z.infer<typeof ProgressPullResponseSchema>;
export type ProgressCursorPullQuery = z.infer<typeof ProgressCursorPullQuerySchema>;
export type ProgressCursorPullResponse = z.infer<typeof ProgressCursorPullResponseSchema>;
export type ErrorResponse = z.infer<typeof ErrorResponseSchema>;
