import { episodeAudioPath } from "../../../../contracts/index.ts";
import { EpisodeContentError } from "../../entities/errors/episode-content-error.ts";
import type { EpisodeListItem, EpisodeManuscript } from "../ports/episode-repository.ts";
import { ManuscriptSchema } from "./manuscript-schema.ts";

/**
 * Driven Adapter が取得した生 payload を、schema 適合かつ stem 一致の検証済み原稿へ変換する。
 *
 * schema 適合・stem 一致・不正 JSON の判定「本体」。Google Drive / in-memory どちらの Adapter
 * からも呼ばれる純関数で、副作用も Port 依存も持たない。
 *
 * 複合失敗時の判定順: schema 適合を先に判定し、通過後に stem 一致を判定する。よって schema
 * 不適合と stem 不一致が同時に起きうる入力では「schema に不適合」を報告する。不正 JSON
 * （非 object）も schema 判定で弾き、専用種別は作らない。
 *
 * @require stem は取得元ファイル名の stem（= 期待 episodeId）
 * @ensure 適合時は検証済み `EpisodeManuscript` を返す。schema 不適合・stem 不一致は
 *   {@link EpisodeContentError} を throw し、理由は message で区別できる（種別クラスは細分しない）
 */
export function verifyManuscript(json: unknown, stem: string): EpisodeManuscript {
  const parsed = ManuscriptSchema.safeParse(json);
  if (!parsed.success) {
    throw new EpisodeContentError(`原稿 JSON が schema に不適合: ${stem}`);
  }
  if (parsed.data.episodeId !== stem) {
    throw new EpisodeContentError(`stem と JSON の episodeId が不一致: ${stem}`);
  }
  return parsed.data;
}

/**
 * `listEpisodes` の1 entry 分を、原稿全文付きの `EpisodeListItem` へ変換する。不適合な entry は
 * throw せず `undefined` を返す（除外は listEpisodes 自身の仕様）。
 *
 * @ensure schema 不適合・stem 不一致・不正 JSON の時は `undefined`。適合時は body 全文と audioRef
 *   を持つ `EpisodeListItem`
 */
export function selectValidListItem(json: unknown, stem: string): EpisodeListItem | undefined {
  const parsed = ManuscriptSchema.safeParse(json);
  if (!parsed.success || parsed.data.episodeId !== stem) {
    return undefined;
  }
  return {
    ...parsed.data,
    audioRef: episodeAudioPath(parsed.data.episodeId),
  };
}
