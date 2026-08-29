import type { GetEpisodeResponse, ListEpisodesResponse } from "../../../../contracts/index.ts";

export type EpisodeListItem = ListEpisodesResponse["episodes"][number];
export type EpisodeManuscript = Omit<GetEpisodeResponse, "audioRef">;

/**
 * 取得したままの原稿 json 1 件。`stem` は取得元ファイル名の stem（= 期待 episodeId）。
 */
export type RawManuscriptEntry = {
  stem: string;
  json: unknown;
};

/**
 * 対象 episodeId の原稿 json と wav の有無。`json` は download して decode しただけの生 payload。
 */
export type RawManuscriptDetail = {
  json: unknown;
  hasAudio: boolean;
};

/**
 * 所定フォルダの原稿 json / wav を「取得したまま返す」Driven Port。schema 検証・stem 一致判定は
 * 一切しない（generator の `port.ItemSource` / `port.EpisodeWriter` の read 方向鏡像）。
 *
 * 実装（`GoogleDriveEpisodeRepository` / `InMemoryEpisodeRepository`）は真の外部境界の I/O
 * （token 取得・files.list・bytes download / Map 出し入れ）だけを担う。schema 適合・stem 一致・
 * 不正 JSON・wav 欠落の判定は use-case（`application/use-cases/*`）が `application/manuscript` の
 * 純関数を使って行う。
 *
 * @invariant vendor 固有型・Drive file id・フォルダ id を露出しない
 */
export interface EpisodeRepository {
  /**
   * 所定フォルダ直下の原稿 json を取得したまま返す。
   *
   * @ensure 各要素の `json` は download して decode しただけの生 payload。該当なしは空配列（null でない）。
   * @ensure Drive HTTP 自体の失敗（token 取得・network・非 2xx・応答形式不正）は Infrastructure Error
   *   （`DriveError`）を throw する。
   */
  listManuscripts(): Promise<RawManuscriptEntry[]>;

  /**
   * 対象 episodeId の原稿 json と wav 有無を取得したまま返す。
   *
   * @ensure `json` は download して decode しただけの生 payload。schema 検証・stem 一致判定はしない。
   * @ensure 対象 episodeId の json エントリ自体が Drive / メモリに無い時は `undefined` を返す
   *   （取得対象が存在しないことを戻り値で表現する。throw しない）。
   * @ensure schema 不適合・stem 不一致・wav 欠落では throw しない。wav の有無は `hasAudio` で返す。
   * @ensure Drive HTTP 自体の失敗は Infrastructure Error（`DriveError`）を throw する。
   */
  getManuscript(episodeId: string): Promise<RawManuscriptDetail | undefined>;

  /**
   * 対象 episodeId の wav byte を取得したまま返す。
   *
   * @ensure wav エントリまたは byte が無い時は `undefined` を返す（throw しない）。
   * @ensure Drive HTTP 自体の失敗は Infrastructure Error（`DriveError`）を throw する。
   */
  getEpisodeAudio(episodeId: string): Promise<Uint8Array | undefined>;
}
