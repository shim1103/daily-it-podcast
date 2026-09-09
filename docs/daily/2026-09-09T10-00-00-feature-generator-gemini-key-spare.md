---
name: 原稿 fallback の Gemini key を TTS と分離し system e2e で実証
date: 2026-09-09T10:00:00
session_id: none
branch: feature/generator-gemini-key-spare
prev: なし
---

## 1. Summary

原稿 fallback（Gemini generateContent）が TTS と同じ `GEMINI_API_KEY` を共有して本番 produce が HTTP 429 で落ちていた（run 34209712652）。`GeminiConfig` へ `SpareAPIKey` を足し、config 契約の必須 env に `SPARE_GEMINI_API_KEY` を加え、`newGeminiTextWriter` だけをそれへ向けた（TTS は `APIKey` 据え置き）。`generator-produce-episode.yml` / `generator-system.yml` へ env 注入、System test の credential 欠落 Skip 判定 key にも追加。`DEPLOY.md` の env 表と `workflow_dispatch` 前提記述（`--ref` で feature branch の tree が回る旨）を更新。判断は Decision `2026-09-09T10-00-00`。本日は本番 Gemini 枠が 429 中のため produce は回さず（shim 指示）、`generator-system.yml` を `--ref feature/generator-gemini-key-spare` で dispatch して検証した。

## 2. Changes

- **config 変更の実証**: `generator-system.yml` run 34298458327 が PASS（所要 508.5s、episodeId `6feba01d-879f-470a-b264-d5074840b94a`、Drive 実到達）。log に `SPARE_GEMINI_API_KEY: ***` の注入と `manuscript text writer switched: from=cursor to=gemini reason=source_exhausted` が出ており、今回は Cursor 枠枯渇 → Gemini fallback の実切替を伴う e2e が完走した。`config.Load` が `SPARE_GEMINI_API_KEY` を必須に加えても壊れず、fallback が `SpareAPIKey` で原稿を生成することを本番相当経路で確認。
- **dispatch 手段の確認**: yml file 名が master にあれば `gh workflow run <yml> --ref <feature-branch>` で feature branch 側の tree（yml の env 変更込み）が実行され、master への merge / checkout は不要。この事実を `DEPLOY.md` §5 の従来記述（「master に yml が無いと 404」だけ）から file 名と tree の branch を分けた形へ修正。
- **DRY 推敲**: 初回 commit で config.go invariant / DEPLOY.md の両方へ書いた「key を分ける理由」を、config.go を正として DEPLOY.md 側を圧縮（`9be8692`）。
- push: `origin/feature/generator-gemini-key-spare`（SSH auth のため sandbox 無効で実行）。develop / master への feature PR は create-pr flow で作成予定。
- 本日は本番 produce の workflow_dispatch を実施せず。次の Gemini 枠回復日に回す（`generator-lane.md` D 表）。

### Commits

- `5ffd171`
- `451e512`
- `a12f6ea`
- `9be8692`
