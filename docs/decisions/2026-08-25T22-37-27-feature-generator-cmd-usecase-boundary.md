---
name: 完成原稿の構築は ProduceEpisode、検査と書込は WriteEpisode（Builder / Gate）
date: 2026-08-25T22:37:27
branch: feature/generator-cmd-usecase-boundary
---

## 1. Decision

1. 完成原稿（`contracts/manuscript.schema.json` に適合する成果）の**構築**は `ProduceEpisode` が行う（Builder）。
2. 完成原稿の**検査**と Drive への書込は既存の `WriteEpisode` が行う（Gate）。定型結合・TTS 呼び出し順・尺の書き込みは `WriteEpisode` に移さない。
3. 原則の正は architecture `backend/application.md` の Builder / Gate。本 Decision は generator への適用先だけを固定する。

## 2. Reason

1. 尺（`startSec` / `durationSec`）を埋める側は完成形の field 配置をすでに知る。検査だけを別 UseCase に残すと「contracts を知るのは Gate だけ」と言い切れず、構築知と検査知が対立して見える。役割を Builder（作る）と Gate（验す+書く）に分けると、両方の関与が矛盾しなくなる。
2. `WriteEpisode` に Intro / 定型 / TTS 順まで寄せると、名前どおりの書込ゲートが台本生成方針まで抱え、定型文の変更理由と schema 検証の変更理由が同居する（SRP 崩れ）。既存の検証ゲートを壊さず Builder を別に置く方が、lesson 109（検証迂回禁止）とも両立する。
3. 抽象規則を architecture に置き、本 file に path や Port 署名を再掲しないことで、同じ問いの再発時は architecture → 本 Decision の順で辿れる。

## 3. Rejected

1. 組み立て・尺・完成 JSON 構築を `WriteEpisode` に寄せ、`ProduceEpisode` を Fetch 後の順送りだけにする案 — Gate が Builder になり、書込 UseCase の変更理由が複数になる。
2. `WriteEpisode` を廃止し、検証なしの `EpisodeWriter` を Composition が公開する案 — 検証を呼び出し側が忘れられる（lesson 109 に反する）。
3. Builder と Gate の両方で完成 schema を Validate する案 — 検査手続きの重複（DRY 割れ）。
