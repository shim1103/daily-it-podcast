---
name: Feature Component の JSX 化と Page 一時橋
date: 2026-08-27T12:36:26
session_id: none
branch: feature/playback-web-feature-component-jsx
prev: なし
---

## 1. Summary

issue-manager で `apps/playback/web/src/components/feature/` 配下5 component（episode-list-item・episode-list・episode-manuscript・episode-player・episode-topic）を DOM API 直書きから JSX 関数コンポーネントへ書き換え、review 指摘（Fault Isolation 境界侵害、skill 記述の DRY 違反）を反映した。Page 層未 JSX 化に伴う波及、sandbox 起因の複数の実行時制約を解消したうえで commit・push まで進めた。

## 2. Changes

1. 5 component を `.tsx` 化し、暫定橋渡し `mount-labeled-text.ts` を削除して `LabeledText` を直接 JSX で使う設計へ揃えた。対応する5本の `*.sociable_unit.test.ts` を `@testing-library/react` の `render()` ベースへ書き換えた。
2. `episode-list.tsx` が `ReactElement` を返すようになったことで Page 層 `episode-list-page.ts`（Out of Scope）の呼び出しが型不整合になる問題が発覚し、`mount-episode-list-view-model.ts` と同型の一時橋 `mount-episode-list.ts` を新設した。`createRoot` した container を取り出さず1度だけ挿入する設計にしないと `onClick` の synthetic event が機能しないことを確認し、その設計で実装した。
3. review 指摘（`episode-list.sociable_unit.test.ts` が協調先 `EpisodePlayer` の URL 組み立て結果まで再 assert していた Fault Isolation 境界侵害）を修正した。
4. shim との対話で `EpisodeListItem` を `<article><button></button></article>` 構造へ変更し `useKeyWithClickEvents` の biome-ignore を解消、`EpisodePlayer` の biome-ignore コメントを `EpisodeManuscript` との対描画契約に基づく実質根拠へ更新した。
5. `architecture/frontend/page-route.md` に unit test 観点セクションを追加し、`testing-strategy/coverage.md` に「coverage 数値は結果指標であり目標ではない」原則を追加した。shim の指摘によりpage-route.md の`§`参照過多（DRY 違反）を修正し、他 component skill と同じ参照密度へ揃えた。
6. sandbox 内で `generator/test-unit.sh` の `mktemp` が `TMPDIR` を無視し `mkstemp failed` で pre-commit hook が落ちる問題を発見し、テンプレート引数を `${TMPDIR:-/tmp}/prefix.XXXXXX` の形へ明示的に組み立てる修正を行った。
7. sandbox 内での golangci-lint cache 破損（cache 削除が sandbox の write 許可で拒否される）、SSH `git push` が sandbox の `ALL_PROXY` 経由で失敗する問題を切り分け、shim の判断（`sandbox.network.allowLocalBinding: true` 設定、`dangerouslyDisableSandbox` での push 実行）で解消した。
8. Verification（typecheck / test:unit 36 files・200 tests / lint / lint:layers）全て pass を確認し、2 commit（feat・fix）を push した。

### Commits

- `ba18c42`
- `dbba95d`
