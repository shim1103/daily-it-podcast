---
name: Drive の json+wav 書込は公開順とし補償は後回しにする
date: 2026-08-30T23:32:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. Drive への episode 書込は **`{episodeId}.json` → `{episodeId}.wav` の公開順**とする。multi-file の原子性は求めない。
2. 途中失敗時の **補償 delete** と **staging→rename** は今は採らない。不完全ペア（片方だけ）が残りうることを許容する。
3. 不完全ペアの読取側の扱いは `contracts/drive-layout.md` と Playback の既存方針（Get で渡せない件は隠す）に従う。
4. 補償・staging の再検討は実装・運用の後続（lane D）とする。本 Decision は「今は公開型」だけを固定する。

## 2. Reason

1. Google Drive に json+wav を 1 トランザクションで確定する API は無い。近似は補償か staging か残骸許容のどれかである。
2. 補償 delete は delete 自体の失敗で残骸が残り、「残骸ゼロ」を約束できない。Adapter に掃除責任も増える。
3. staging→rename は layout と手順が増え、今の主失敗因（Cursor / Gemini）に対する効果が薄い。
4. cron は 1 日 1 回・stem は UUID なので、不完全ペアの頻度は低く、Playback Get が隠す妥協で足りる。

## 3. Rejected

1. 失敗時に書いた file を必ず delete する補償案 — 補償失敗で状態がさらに汚れる。今の優先ではない。
2. 一時名で書いて両揃い後に最終名へ移す staging 案 — YAGNI。layout 契約が複雑になる。
3. 不完全ペア専用の外部 error code を増やす案 — Playback 先行 Decision で既に Rejected（Drive 内部事情の漏洩）。
