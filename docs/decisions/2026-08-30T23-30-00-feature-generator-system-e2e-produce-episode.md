---
name: 同日の完成 episode があれば Fetch 前に skip して成功とする
date: 2026-08-30T23:30:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. 表示用 `date`（JST 暦日、`YYYY-MM-DD`。正は先行 Decision `2026-08-29T14-13-00` / `2026-08-15T16-23-07`）につき、Drive 上に **完成ペア**（同一 stem の `{stem}.json` と `{stem}.wav` が両方ある）が既にあるとき、`ProduceEpisode.Run` は **Fetch より前**に打ち切り、**成功**（error なし）とする。
2. 判定に使う鍵は **暦日 `date`** である。原稿 byte の同一性は見ない。
3. 同日に **片方だけ**（json のみ / wav のみ）があるときは skip しない。通常どおり Produce を続行する（新 UUID で完成ペアを書く）。Domain Error にも同 stem upsert にもしない。
4. 同日に完成が無いときは通常どおり続行する。
5. 規則の所有は **Application**。Drive 照会の手段は **Port** とし、Infrastructure 実装を Composition が DI する。`WriteEpisode` 内の直前再チェックは今は置かない。
6. stem を日付文字列にする案は採らない（先行 Decision `2026-08-29T14-12-00` の opaque UUID を維持する）。

## 2. Reason

1. 本番 cron は 1 日 1 回だが、手動 `workflow_dispatch` や成功後の再実行で同日に 2 本目が積める。GetX / Cursor / Gemini はいずれも token・quota を消費するため、完成が既にあるなら Fetch より前に止めるのが最小コストである。
2. 「1 暦日 1 配信可能 episode」を業務規則にするなら、運用頻度だけに賭けるより Application で enforce した方が再発時の答えが固定する。
3. 鍵を content hash にすると Adapter が本文を知る。鍵を stem UUID だけにすると「同日」が表現できない。表示 `date` が既に JSON にあるので、同日判定の正本として足りる。
4. 片方だけの残骸を Domain Error にすると、公開型書込の途中失敗が翌朝の定時を赤にする。upsert（既存 stem 再利用）は UUID 直前発行の流れと衝突し、設計面が増える。完成が無いなら新規 Produce で足り、不完全は Playback Get が隠す。
5. `WriteEpisode` 直前の二度目チェックは、単一 process の cron では効きが薄く、今は YAGNI である。

## 3. Rejected

1. stem を `YYYY-MM-DD` にする案 — id と表示 date が混ざる（先行 Decision `2026-08-15T16-23-07` / `2026-08-29T14-12-00` Rejected と同型）。1 日 2 本にしたいときの変更も重い。
2. 同日があれば中身を update し直す案 — 毎回 Cursor / Gemini を焼き、token 節約の目的に反する。
3. 同日の片方残骸を Domain Error にする案 — 残骸があると定時が失敗し、公開型書込の妥協と両立しない。
4. 同日の片方残骸へ既存 stem で upsert する案 — ID 回収と部分上書きが要り、今の「UUID を Write 直前に 1 回発行」と噛み合わない。
5. `WriteEpisode` 内だけで同日判定する案 — Fetch / Cursor / Gemini の後では token を既に消費している。
6. content 同一性で dedup する案 — Application が原稿本文比較を抱え、同日規則より重い。
