---
name: apps/generator の Goroutine 導入 A/B/C 確定と agent 環境 unstack 化
date: 2026-09-23T19:23:27
session_id: なし
branch: refactor/generator-go-performance
prev: なし
---

## 1. Summary

apps/generator における Goroutine 導入可能箇所の technical survey から、scope-split の A（境界契約）・B（Decision）・C（実装Issue）・D（release lane）分類を経て、fan-out 化5箇所と尺計算統合の境界契約を固定した。commit 中に発覚した agent 環境（.claude/.codex/.cursor）の track 方針との不整合も併せて解消した。

## 2. Changes

1. executor agent による apps/generator 全体調査を経て、Goroutine 化候補を fan-out 可否・rate limit 有無込みで洗い出した
2. scope-split で A（interface・定数・contract-docs）/ B（Decision）/ C（Issue）へ分類し、shim との往復で対象・concurrency 上限・型設計・Issue 分割単位を確定した
3. `errgroup` 採用、fan-out 対象5箇所（compositeItemSource・HackerNews story/comment・Lobsters・HasPair）、見送り3箇所を Decision として固定した
4. WAV 尺計算を `port.SpeechSynthesizer` 契約側（`models.SpeechAudio.DurationSec`）へ統合する方針を Decision として固定した
5. `scope-split` の D（未決lane）へ release 単位の Issue 依存関係記述を仕様追加した（`~/settings/agent-standards` 側、このrepoの管理外）
6. `docs/tasks/todo/lane.md` へ Release lane 区画を新設し、Issue2本を登録した
7. commit 中に `.claude/settings.json` / `.codex/config.toml` / `.cursor/cli.json` が `.gitignore` の ignore 方針と食い違い track されたままだったことが判明し、`rm --cached` で解消した

### Commits

- `0740740`
- `3a5a16d`
- `372cc59`
- `452bbc7`
- `660dcad`
- `aae12c7`
