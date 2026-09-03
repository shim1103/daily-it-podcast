import type { ApiSuccessData } from "../api/api-result.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";

type ListEpisodesData = ApiSuccessData<PlaybackApiClient["listEpisodes"]>;

export type EpisodeData = ListEpisodesData["episodes"][number];

export type CatalogStatus = { status: "loading" } | { status: "success" } | { status: "error" };

export type SelectionState = { selected: false } | { selected: true; episode: EpisodeData };

/**
 * audio 取得に失敗した理由。現状は 1 値のみ。将来 decode 失敗等を足す受け皿として union で持つ。
 */
export type AudioFailureReason = "audio-load-failed";

/**
 * 再生中 episode の phase。`error` 枝のみ失敗理由（`reason`）を持ち、他 phase は追加データを持たない。
 * `lib/audio-element.ts` の `AudioLifecyclePhase`（4 値）に `loading` を足した 5 枝。
 */
export type PlaybackPhase =
  | { phase: "loading" }
  | { phase: "playing" }
  | { phase: "paused" }
  | { phase: "ended" }
  | { phase: "error"; reason: AudioFailureReason };

/**
 * 再生 state。`idle` は再生対象がなく付随データを持たない。`active` は再生対象があり、
 * `phase` に関わらず現在位置（`positionSec`）と長さ（`durationSec`、metadata 未取得なら null）を
 * 必ず持つ。停止中でも位置は `positionSec` に残る（B Decision §1-1/§1-2）。
 */
export type PlaybackState =
  | { kind: "idle" }
  | {
      kind: "active";
      episodeId: string;
      phase: PlaybackPhase;
      positionSec: number;
      durationSec: number | null;
    };

/** `PlaybackState` の `active` 枝。再生対象があり位置・長さを必ず持つ。 */
export type ActivePlayback = Extract<PlaybackState, { kind: "active" }>;

/**
 * page 全体の振る舞いを決める blocking 判定のみを持つ型。non-blocking（audio 失敗）は
 * `PlaybackState` の `phase:"error"` 枝が持ち、この型からは分離する。
 */
export type PageStatus =
  | { kind: "loading" }
  | { kind: "unavailable"; reason: "catalog-load-failed" }
  | { kind: "ready" };

/**
 * component が 1 row に必要とする投影。表示整形は Row component 側が `episode` から行う。
 * `episode` を持つことで、呼び出し側が別配列との index 結合をせず 1 row を組み立てられる。
 * `episodeId` は key・identity 参照用の冗長 field（`episode.episodeId` と同値）。
 * `isPlayed` は再生対象であること（phase 不問）、`isPlaying` は phase が playing であることを表す。
 */
export type EpisodeRowViewModel = {
  episode: EpisodeData;
  episodeId: string;
  isSelected: boolean;
  isPlayed: boolean;
  isPlaying: boolean;
};

/**
 * 一覧 cache と再生 union から再生中 episode を導出する。
 *
 * @ensure `playback.kind` が `idle`、または `episodeId` が一覧に無いとき null
 */
export function derivePlayedEpisode(
  episodes: readonly EpisodeData[],
  playback: PlaybackState,
): EpisodeData | null {
  if (playback.kind === "idle") {
    return null;
  }
  return episodes.find((episode) => episode.episodeId === playback.episodeId) ?? null;
}

/**
 * catalog の取得状態から page 全体の振る舞い（`PageStatus`）を導出する。
 *
 * @ensure catalog error は unavailable/catalog-load-failed、catalog loading は loading、
 *   catalog success は ready。選択・再生の異常は blocking 判定に影響しない
 */
export function derivePageStatus(catalogStatus: CatalogStatus): PageStatus {
  const status = catalogStatus.status;
  switch (status) {
    case "error":
      return { kind: "unavailable", reason: "catalog-load-failed" };
    case "loading":
      return { kind: "loading" };
    case "success":
      return { kind: "ready" };
    /* v8 ignore next 6 -- CatalogStatus は 3 値の union 型で、型検査上この分岐へ実行が到達しない。将来値が増えた時に tsc が検知するための exhaustiveness check（defensive-design.md §6）。網羅性ガードの never 代入と到達時 fallback は別責務のため両方置く */
    default: {
      const exhaustive: never = status;
      void exhaustive;
      // why: 「使えない」を返すのが loading より安全側（本体を描かせない）
      return { kind: "unavailable", reason: "catalog-load-failed" };
    }
  }
}

/**
 * 一覧と選択・再生 union から、component が 1 row に必要とする投影の配列を導出する。
 *
 * @ensure 各 row は入力 episode の実体を order 通りに持つ。選択中なら isSelected=true、
 *   `kind:"active"` の対象 episode（phase 不問）なら isPlayed=true、さらに `phase:"playing"` なら
 *   isPlaying=true。それ以外は false。`active` の episodeId が一覧に無い row は、等値比較が
 *   一致しないため isPlayed も自然に false になる
 */
export function deriveEpisodeRows(
  episodes: readonly EpisodeData[],
  context: { selection: SelectionState; playback: PlaybackState },
): EpisodeRowViewModel[] {
  const selectedEpisodeId = context.selection.selected ? context.selection.episode.episodeId : null;
  const playedEpisodeId = context.playback.kind === "active" ? context.playback.episodeId : null;
  const playingEpisodeId =
    context.playback.kind === "active" && context.playback.phase.phase === "playing"
      ? context.playback.episodeId
      : null;
  return episodes.map((episode) => ({
    episode,
    episodeId: episode.episodeId,
    isSelected: episode.episodeId === selectedEpisodeId,
    isPlayed: episode.episodeId === playedEpisodeId,
    isPlaying: episode.episodeId === playingEpisodeId,
  }));
}
