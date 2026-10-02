/**
 * 現在時刻の Port。非決定な時刻を Application の外へ出す窓口。
 *
 * @ensure `now()` は呼出し時点の現在時刻を返す
 * @invariant Application は Clock 経由でのみ現在時刻を得る（`Date.now`・引数なしの `new Date()` を直接呼ばない）
 */
export interface Clock {
  now(): Date;
}
