import type { EpisodeItem, ListEpisodesResponse } from "../../../../contracts/index.ts";

export type EpisodeListItem = ListEpisodesResponse["episodes"][number];
export type EpisodeManuscript = Omit<EpisodeItem, "audioRef" | "progress">;

/**
 * 取得したままの原稿 json 1 件。`stem` は取得元ファイル名の stem（= 期待 episodeId）。
 */
export type RawManuscriptEntry = {
  stem: string;
  json: unknown;
};

/**
 * 所定 object 空間の原稿 json / 音声を「取得したまま返す」Driven Port。schema 検証・stem 一致判定は
 * 一切しない（generator の `port.ItemSource` / `port.EpisodeWriter` の read 方向鏡像）。
 *
 * 実装（R2 / InMemory）は真の外部境界の I/O だけを担う。schema 適合・stem 一致・
 * 不正 JSON・音声欠落の判定は use-case（`application/use-cases/*`）が `application/manuscript` の
 * 純関数を使って行う。
 *
 * @invariant vendor 固有型・storage 固有 id を露出しない
 */
export interface EpisodeRepository {
  /**
   * 所定空間直下の原稿 json を取得したまま返す。
   *
   * @ensure 各要素の `json` は取得して decode しただけの生 payload。該当なしは空配列（null でない）。
   * @ensure storage I/O 自体の失敗（認証・network・非 2xx・応答形式不正）は Infrastructure Error を throw する。
   */
  listManuscripts(): Promise<RawManuscriptEntry[]>;

  /**
   * 対象 episodeId の音声 byte を取得したまま返す。
   *
   * @ensure 音声エントリまたは byte が無い時は `undefined` を返す（throw しない）。
   * @ensure storage I/O 自体の失敗は Infrastructure Error を throw する。
   */
  getAudio(episodeId: string): Promise<Uint8Array | undefined>;
}
