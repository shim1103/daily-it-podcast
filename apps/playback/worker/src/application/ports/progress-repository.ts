import type { EpisodeProgress } from "../../entities/models/episode-progress.ts";

/** pull で返す 1 件。更新があった episode の進捗。 */
export type ProgressUpdatedEntry = {
  readonly episodeId: string;
  readonly progress: EpisodeProgress;
};

/** 行の版。Application からは不透明で、同じ行の `replaceProgressIfVersion` の期待値にだけ使う。 */
export type ProgressVersion = number & { readonly __brand: "ProgressVersion" };

/** 差分取得の位置。非負整数。文字列との変換と形式の検証は Controller が行い、Port と adapter は変換しない。 */
export type ProgressCursor = number;

/** 進捗本体と、それを読んだ時点の版。 */
export type VersionedProgress = {
  readonly progress: EpisodeProgress;
  readonly version: ProgressVersion;
};

/** 進捗 1 行。`episodeId` + 進捗本体。merge せずそのまま永続する。 */
export type ProgressRow = { readonly episodeId: string } & EpisodeProgress;

/** 条件付き書込の結果。`conflict` は前提（行なし・期待した版）と行の現在の状態が食い違ったこと。 */
export type ProgressConditionalWriteResult = "written" | "conflict";

export type GetByEpisodeIdsArgs = { readonly episodeIds: readonly string[] };
export type GetByEpisodeIdsResult = ReadonlyMap<string, EpisodeProgress>;

export type GetProgressWithVersionArgs = { readonly episodeId: string };
export type GetProgressWithVersionResult = VersionedProgress | null;

export type InsertProgressIfAbsentArgs = { readonly row: ProgressRow };

export type ReplaceProgressIfVersionArgs = {
  readonly row: ProgressRow;
  readonly expectedVersion: ProgressVersion;
};

export type ListChangedAfterArgs = { readonly cursor: ProgressCursor };

/** 差分取得の結果。`cursor` は次回の `listChangedAfter` に渡す位置。 */
export type ListChangedAfterResult = {
  readonly cursor: ProgressCursor;
  readonly entries: readonly ProgressUpdatedEntry[];
};

/**
 * 再生進捗の永続 Port。1 episode = 1 行。鍵は `episodeId` のみ（user 分割なし）。
 * merge・冪等・完走判定は Application。本 Port は永続能力のみ（Decision 2026-10-01T18:54:14）。
 *
 * @invariant vendor / D1 固有型を露出しない
 */
export interface ProgressRepository {
  /**
   * 指定 episode の進捗行を一括取得する。
   *
   * @ensure 行が無い episodeId は Map に載せない（caller が `progress: null` にする）
   * @ensure storage I/O 失敗は Infrastructure Error を throw する
   */
  getByEpisodeIds(args: GetByEpisodeIdsArgs): Promise<GetByEpisodeIdsResult>;

  // todo: C が insertProgressIfAbsent・replaceProgressIfVersion へ置換したら、この面と全 adapter の実装・test を削除する
  /**
   * 進捗行をそのまま永続する（create / update の区別なし。merge しない）。
   *
   * @ensure 渡された row を episodeId 鍵で置き換える。応答を返さない
   * @ensure storage I/O 失敗は Infrastructure Error を throw する
   */
  upsertProgress(row: ProgressRow): Promise<void>;

  // todo: C が listChangedAfter へ置換したら、この面と全 adapter の実装・test を削除する
  /**
   * `since` より後に更新された進捗行を返す（HTTP pull の永続側）。
   *
   * @ensure `lastPlayedAt > since` の行だけを返す
   * @ensure 該当なしは空配列（null でない）
   * @ensure storage I/O 失敗は Infrastructure Error を throw する
   */
  listUpdatedSince(since: string): Promise<readonly ProgressUpdatedEntry[]>;

  /**
   * 指定 episode の進捗行を、その時点の版つきで取得する。
   *
   * @ensure 行が無い episodeId は null を返す（throw しない）
   * @ensure `version` は同じ行の `replaceProgressIfVersion` の期待値にだけ使う不透明な値
   * @ensure storage I/O 失敗は Infrastructure Error を throw する
   */
  getProgressWithVersion(args: GetProgressWithVersionArgs): Promise<GetProgressWithVersionResult>;

  /**
   * 行が無い時だけ、進捗行を書く（行なしと読んだ caller の書込）。
   *
   * @ensure 行が無い時だけ row を書いて `"written"` を返す。`written` のたびに行の版が新しくなり、`listChangedAfter` の cursor が進む
   * @ensure 既に行がある時は何も書かず `"conflict"` を返す（競合では throw しない）
   * @ensure adapter 内で再試行しない
   * @ensure storage I/O 失敗は Infrastructure Error を throw する
   */
  insertProgressIfAbsent(args: InsertProgressIfAbsentArgs): Promise<ProgressConditionalWriteResult>;

  /**
   * 期待した版と行の現在の版が一致する時だけ、進捗行を置き換える（行ありと読んだ caller の書込）。
   *
   * @ensure 版が一致した時だけ row を置き換えて `"written"` を返す。`written` のたびに行の版が新しくなり、`listChangedAfter` の cursor が進む
   * @ensure 版が一致しない時（行が無い時を含む）は何も書かず `"conflict"` を返す（競合では throw しない）
   * @ensure adapter 内で再試行しない
   * @ensure storage I/O 失敗は Infrastructure Error を throw する
   */
  replaceProgressIfVersion(
    args: ReplaceProgressIfVersionArgs,
  ): Promise<ProgressConditionalWriteResult>;

  /**
   * `cursor` より後に書かれた進捗行を、書込順で返す（HTTP pull の永続側）。
   *
   * @require cursor は非負整数で、契約の起点（`PROGRESS_PULL_ORIGIN_CURSOR`）を整数にした値か、本 Port が返した値
   * @ensure entries は cursor より後に書かれた行だけを書込順で返す。該当なしは空配列（null でない）
   * @ensure 返す cursor は次回の引数に使う値。entries があれば最後の行を指し、無ければ引数と同じ値
   * @ensure storage I/O 失敗は Infrastructure Error を throw する
   */
  listChangedAfter(args: ListChangedAfterArgs): Promise<ListChangedAfterResult>;
}
