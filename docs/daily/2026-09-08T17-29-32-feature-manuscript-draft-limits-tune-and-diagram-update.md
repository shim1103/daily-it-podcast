---
name: ManuscriptDraft 尺定数の実運用調整と runtime 構成図の実態追随
date: 2026-09-08T17:29:32
session_id: none
branch: feature/manuscript-draft-limits-tune-and-diagram-update
prev: なし
---

## 1. Summary

ManuscriptDraft の topic 数と全体尺を実運用感に寄せて調整した。既存の尺モデル（秒正本 + `CharsPerSecond` 畳み込み）と contract test の 7 不変条件は変えず、値だけを動かした。全体 target は round な 16 分に固定し、topic 構成の端数を intro / closingSummary の target で吸収して「target 構成合計 == 全体 target」を厳密一致させた。

あわせて runtime 構成図を現状のスタックと経路へ更新した。情報源 3 サイトの icon 分割、原稿の Cursor 第一 / Gemini fallback 併記、Vite / React / TypeScript の個別 node 化、Hono と Workers の分離、Hono route の RPC / 音声 HTTP GET 2 経路化、OAuth 2.0 の明記。

## 2. Changes

1. 尺定数の contract test は数値をハードコードしない設計のため無改修で新値でも 7 不変条件 pass。`internal/application` / `internal/application/build` の fixture 2 本は新 total 下限に追随する機械的調整のみ。
2. `docs/tasks/todo/generator-lane.md` の D 表 2 行から旧レンジ（topic 3〜7・全体 8〜12 分）の再掲を除き、定数 file 参照へ置換（DRY）。
3. 構成図の icon catalog へ 5 件追加（github-actions / hackernews(Y Combinator) / lobsters / itmedia-rss(RSS) / vite）。Simple Icons を公式ブランドカラー注入で cache。
4. Hono の route が RPC 単一ではなく `/episodes`（RPC JSON）と `/episodes/:id/audio`（素の HTTP GET・Range）の 2 経路であることを code で確認し、図のエッジを 2 本へ分離。
5. PR: #（未作成、本 log の次工程）
6. `docs/lessons/index.md` へ 7 件追記（数式検算 / non-scope と明示指示の優先 / 図の icon 差し替えと識別情報 / エッジラベルと route 実装確認 / 言語 node の対称性 / 同系色 icon と階層表現）。

### Commits

- `92240941c5d8aeab7f145ef72ae44c4e406fb273`
- `b6046f0ea8569e17f3372fc0e126783a73ffa39d`
