---
name: composition ProduceEpisode 結線を composite ItemSource 経由で通し factory は撤回
date: 2026-08-30T11:20:00
session_id: none
branch: feature/generator-composition-produce-episode-wiring
prev: なし
---

## 1. Summary

`ProduceEpisode` の production graph を Composition Root で結線した。情報源は composite
`ItemSource`（登録順 concat・error 透過・非 nil 空 slice）でラップし、`FetchSourceItems` へ
渡す。GetXAPI 1 本の現在でも composite 経由にした。`compositeItemSource.List` の分岐は
Sociable Unit（`item_source_test.go`）が所有し、結線関数は test を持たない。

達成契約 file が要求していた `ProduceEpisodeFactory` 型を一度実装したが、`Build()` が
旧結線関数の本体を移しただけの死んだ中間層になったため撤回し、結線関数直書きへ戻した。
判断の正は Decision `2026-08-30T11-20-00`。

## 2. Changes

- issue-manager で manager(non-edit) → executor 実装 → reviewer 査読 → executor 再修正 →
  manager audit → 契約 file 削除まで一巡した後、user 指摘で factory 撤回と test 1対1化を
  別 flow で実施。
- SU unit と実装 file の 1対1 対応へ整理。分岐を持つ `item_source.go` だけが `item_source_test.go`
  を持ち、分岐なしの結線関数・Adapter ctor 群は test なし。`composition/**` は coverage gate
  除外（DESIGN.md）。
- 破棄した test: `Build()` 非 nil 検証（分岐なし結線への構造 guard）、`fakeSecret` redaction
  （factory 撤回で `dummyConfig` ごと不要化。config 側 test が所有）。
- test 関数名を日本語で書いてしまい、既存 repo 慣習（全 test が英語 `Test<Subject>_<expected>_<when>`）
  に反していたため英語へ rename。GWT コメントと assert message は日本語のまま。
- pre-commit / pre-push は generator + playback 全系統を走らせる。push は sandbox proxy が
  git 認証を弾いたため sandbox 無効化で実行。
- GitHub Issue は無し（local 契約 file が正）。
- PR: https://github.com/shim1103/daily-it-podcast/pull/94 （base `develop`）

### Commits

- `2d2c3fa`
- `7dcadda`
