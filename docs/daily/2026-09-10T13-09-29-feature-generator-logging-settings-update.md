---
name: generator の log / error policy を一元化し produce の段階進捗を Start/Done で追える形にした
date: 2026-09-10T13:09:29
session_id: none
branch: feature/generator-logging-settings-update
prev: なし
---

## 1. Summary

GHA run 34418654127 で本番 produce が `generator: kind=unknown` を出して落ちた（infra 401 が分類不能）。原因は `delivery.Format` の `classify` が infra 各 package の Error 型を instanceof で並べており、原稿 fallback の `geminiapi.Error` の case を持っていなかったこと。これを機に generator の Error / log まわりを規約準拠へ整理した。(1) error 分類契約 `Kinded`（`ErrorKind`/`ErrorOp`）と kind 語彙定数を `entities/errors` へ置き、`classify` は具体型を知らず interface だけで kind/op を採る。(2) infra Adapter 8 個が持っていた同一構造の Error 型を `internal/infrastructure/adaptererror.Error{Source,Op,Err}` 1 型へ集約（各 Adapter は source 定数のみ）。(3) 応答 body snippet の 3 コピーを `internal/infrastructure/httpdiag` へ集約。(4) `entities/errors/error.go` を `kind.go`（契約）/ `domain.go`（domain 実体）へ分割。(5) 生 log（`composition` の `log.Print`）を廃し `delivery.LogWriter` を CLI 唯一の log 出力面に。(6) produce の各段階を Start（呼び出し直前）/ Done（成功後のみ・結果サマリ付き）の 2 点で通知し、失敗時は Start だけ残す。application は `port.ProgressReporter`/`FallbackReporter` へ文字列を渡すだけ。(7) fallback log から vendor 具体名を落とし event 名 1 語へ。(8) geminiapi の read_body 失敗を 5xx と同じ一過性 retry へ揃え、`sameGeminiOp` に Source 比較を追加、body snippet を出す全経路に secret 非露出の why を対称に置いた。

## 2. Changes

- **skill 更新**: 既存 `architecture/backend` skill が infra Adapter 1 個前提（playback）で、Adapter 8 個の generator に規約が答えを持たなかった。`backend/SKILL.md` §2-1 新設（package root 直下は層 dir のみ、共有物は使う層 dir 直下の非分類区画へ）、`backend/entities.md` §2（errors/ = domain 実体 + 層横断の分類契約、Infra Error 実体は置かない）、`backend/infrastructure.md` §4（Adapter 複数で構造同一なら infra 層内の非platform package へ集約、layer root に置かない、分類契約は Entities 所有）、`architecture/error-taxonomy.md` §6 新設（Error 実体と分類契約の分離）。playback（infra Adapter は drive 1 個、error は colocate）は更新後 §4-3 でそのまま通ることを確認。
- **検証**: pre-commit hook が 2 commit とも full gate green（generator: depguard/errcheck/govet/gofmt/build 0 issues、unit coverage 91.7% >= 90%、Broad Integration 込みの `go test ./...`。playback: biome/tsc/dependency-cruiser/vitest 365 tests・coverage 100%）。
- **設計の作り直し**: 初期案 `entities/errors.InfraError` は `entities.md` §2 違反、`internal/httpdiag/`（root 直下）は §2-1 違反、`internal/infrastructure/responsediag/`（platform 名慣例破り）はいずれも reject。`Kinded` の `ErrorKind()` 重複を Go embedding で消す案は typed-nil で promoted method が panic するため revert（定数返却の method は手書き 4 箇所のまま。nil 安全の対称を優先）。
- **並行実行**: 追加の小変更 4 群（Kind embed→revert / Start-Done 2 点化 / sameGeminiOp+why+lane / 統合）を担当 file 排他で haiku subagent 並列。
- **lane 更新**: `generator-lane.md` D 表の「HTTP 応答 body snippet の共通化」を実装済み（`httpdiag` へ集約）へ、「Gemini fallback 発火の観測」の `logManuscriptSourceSwitched` 参照を `delivery.LogWriter.Fallback` へ書き換え（撤去済み識別子の陳腐化解消）。
- push / PR は create-pr flow で実施予定。

### Commits

- `612e3a4`
- `6758616`
