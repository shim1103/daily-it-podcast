---
name: 進捗のmergeはD1内の1文upsertで行い、比較のため時刻をUTC固定幅ISOへ正規化して保存する
date: 2026-10-01T18:54:16
branch: feature/playback-progress-d1-local-peer
---

## 1. Decision

1. 進捗の書込（create・update・complete）は、read-then-writeでなく、勝ち側を `RETURNING` で受ける1文のupsert（`INSERT ... ON CONFLICT DO UPDATE`）でD1内に閉じる。どの字段が先勝ち・後勝ちかは `2026-09-22T19-13-35-feature-playback-progress.md` に従い、ここへ写さない。
2. 時刻列は、保存前にUTC固定幅のISO表記へ正規化して保存する。応答の時刻もUTCの `Z` 表記になる。pullの `since` も同じ正規化を通してから比較する。正規化は契約schemaを通った後の adapter 内で行い、契約が許す入力の形は変えない。

## 2. Reason

1. read-then-writeは、2端末が同じ行を読んでからそれぞれ書くと、後の書込が先に勝った側を上書きしうる。`first*` の先勝ちと、`lastPlayedAt` の後勝ちが壊れる。1文なら行ごとの単行原子性（`2026-09-22T19-23-39-feature-playback-progress.md`）に任せられ、競合のための補償や再読込を持たずに済む。
2. 書込結果として返す `firstPlayedAt`・`firstCompletedAt` は、merge後に勝った側の値である。`RETURNING` で同じ文から受ければ、書込の後に読み直す追加の読取が要らない。
3. 上の1文はD1のTEXT列を `MIN` と辞書順で比べる。契約は `Z` と `±hh:mm` の両方を許すため、offset混在では辞書順が時系列とずれる（`2026-10-01T09:00:00+09:00` は `2026-10-01T01:00:00Z` より後ろに並ぶが、時刻は前）。受理時にUTC固定幅へ揃えれば、文字列の大小が時系列と一致し、SQL側は素の比較で足りる。
4. 入口で `+09:00` 等を拒むと、契約が許す入力を狭める。契約を変えず、保存の形だけで一致させる。

## 3. Rejected

1. **offsetのまま保存し、TEXTの比較に任せる案** — 辞書順がoffset混在で時系列とずれ、先勝ち・後勝ちが誤る。
2. **SQLの時刻関数（`julianday` / `strftime`）で比較する案** — 保存を正規化すれば素の比較で足りる。D1のSQLiteの版と関数の許可は未確認で、確認していないものへ依存を増やさない。
3. **read-then-writeで、adapterがDateで比較して書き戻す案** — 読取と書込のあいだに別の書込が入り、勝ち側を上書きしうる。補償や再読込が要り、単行原子性に任せる方針（`2026-09-22T19-23-39-feature-playback-progress.md`）とも合わない。
