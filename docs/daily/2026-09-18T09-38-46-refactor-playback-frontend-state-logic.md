---
name: playback web の catalog error伝播とview-model公開面の整理
date: 2026-09-18T09:38:46
session_id: none
branch: refactor/playback-frontend-state-logic
prev: なし
---

## 1. Summary

`use-episode-list-page.ts`のcode review依頼を起点に、view-model層の公開面整理（未使用field削除・contract-docs圧縮）、catalog取得失敗のerror伝播（HTTP error 7種の分類とUI表示・retry導線）、および関連するcode-docs/decisionのDRY整理までを一連のrefactoringとして進めた。

途中、non-edit方針のturnで誤ってfileをeditしてしまう事故が発生し、hookのworktree破棄禁止policyとcommit無断禁止policyが衝突して詰まったため、後始末はuser（Cursor経由）に委ねた。

## 2. Changes

1. `useEpisodeListPage`の公開面から未消費field（`selectedEpisode`）を削除し、`playback`（生union）は残しつつ理由をwhyコメントに明記した。JSDocの実装walk-throughを責務宣言のみへ圧縮し、1回しか使わないconst中間変数をinline化した
2. `useEpisodeCatalog`から不要な`apiClientRef`（latest-ref pattern）を削除した。呼び出し元がmodule scopeで1度だけ`apiClient`を生成しrender間で同一参照を保つ前提のため、`useCallback`のdepsへ直接渡す形へ単純化した
3. `readJsonResult`を`resolveApiResult`へ改名した。関数はHTTP status検証・schema検証を含む3責務を持つが、旧名は「json読み取り」しか示唆せずstatus検証を含む事実を隠していた
4. `CatalogStatus.error`に`PlaybackApiErrorCode`（7種）を保持させ、`playback-state.ts`に宣言的mapping（`catalogErrorPresentation`）を追加してUI向けmessage・retry可否を導出するようにした。catalogは画面唯一のprimary contentのため、7種いずれも表示形態は`unavailable`（全体を覆う）へ収束し、retryable（`network_error`/`unavailable`のみtrue）だけが分岐軸になるという設計判断をDecisionへ記録した
5. `retryable=true`の時だけmessage下に出す小さい青系retry button（`.page-error-retry`）を実装し、`useEpisodeListPage.retry()`で`catalog.load`を配線した。broad integration testにbutton描画・click後の再取得・非retryable時の非表示を追加した
6. `EpisodeRowViewModel.episodeId`（`episode.episodeId`と同値の冗長field）を削除し、`key`・`seek`呼び出しを`row.episode.episodeId`経由へ統一した
7. decisionへ記録済みの設計判断をcode commentへ全文転記しない規約を`coding-style/comments.md`へ追加し、「失敗対象が画面唯一のprimary contentの場合は表示形態が性質分類に関わらず全体を覆うへ収束する」という判定基準を`error-handling/defensive-design.md`§8へ追加した
8. Decision本文とcode実態の食い違い（`client_error`のcomment記述先）を修正し、`PlaybackState`非正規化設計の再検討事項を`docs/tasks/todo/lane.md`へ追記した

### Commits

- `601de94`
- `a470153`
- `a427450`
- `bbf622b`
- `aa583d2`
- `e18bbba`
- `386ccc3`
- `aaf2003`
- `794c8bd`
