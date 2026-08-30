---
name: Playback 安定 E2E fixture は ProduceEpisode 経路の validation（TextWriter wire 含む）を通る完成成果物とする
date: 2026-08-31T00:22:00
branch: feature/playback-e2e-deploy
---

## 1. Decision

1. 安定 E2E fixture（本番 `DRIVE_FOLDER_ID` へ人手配置する `{episodeId}.json` + `{episodeId}.wav`）は、**`ProduceEpisode` が書込前に通す validation を満たした完成成果物**とする。少なくとも TextWriter 出力の Domain Rule（`ManuscriptDraftFromWriterOutput` / `manuscript_draft_limits`）と、完成稿の `WriteEpisode`（`manuscript.schema.json` + stem 一致 + 非空音声）を通る。
2. 本 Decision は先行 Decision（`2026-08-31T00-12-01-feature-playback-e2e-deploy.md`）のうち「fixture の原稿は WriteEpisode の検証を通る」という範囲を **上書き**する。同 file の他項（≥1 正常系・本番 folder・専用 folder 非作成・日次増分の受け入れ）は維持する。
3. fixture の具体値・生成手順の正本は repo 内 artifact（`apps/playback/test/e2e/fixtures/stable-episode/`）とする。本 Decision に UUID・本文・秒数を写さない。

## 2. Reason

1. Drive 上の完成稿だけが `WriteEpisode` を通っても、TextWriter wire の文字数・句点・topic 数などを満たさない「本番では書けない稿」になり得る。E2E が観測する稿は、日次 produce が出しうる形であるべきである。
2. Generator CLI 入口を毎回回す必要はない。Application の parse / assembly / WriteEpisode と同じ規則を満たす bytes を置けば、経路上の Gate と整合する。
3. 値の正本を Decision に写すと artifact と二重になる。

## 3. Rejected

1. `WriteEpisode`（schema）だけ通せば足りる案 — TextWriter Domain Rule を満たさない稿が本番 fixture になり、produce 経路と乖離する。
2. 毎回 `ProduceEpisode` CLI / 実 Cursor・Gemini で fixture を作る案 — E2E 準備が vendor・secret・非決定出力に縛られる。
3. Decision 本文に draft 文字数や greeting 文言を正本化する案 — constants / artifact と二重になる。
