---
name: ProduceEpisode Broad Integration を secret なしで gate に載せる
date: 2026-08-30T15:43:58
session_id: none
branch: feature/generator-broad-integration-produce-episode
prev: なし
---

## 1. Summary

secret なし Broad Integration で `ProduceEpisode` 合成経路（GetXAPI → Cursor → Gemini → OAuth+gdrive）の配線・状態伝播・error 伝播を self-validate した。真外部は multi-host TLS redirect と fake Cursor agent。`NewProduceEpisodeFromEnv` は通さず production Adapter 同型の手組み。Narrow 二重 helper は中立 support へ寄せた。issue-manager で AC 全 pass 後に達成契約 file を削除し、lane を完了反映した。

## 2. Changes

- issue-manager: manager plan → executor 実装 → code-reviewer → 採択修正 → Verification 独立再実行 → issue file 削除。途中 Task 中断あり、成果物確認後に reviewer から再開。
- Broad 4 case: 成功書込 1 組、0 件 `no_source_items`、TextWriter 失敗、Synthesize 失敗。Authorization / path / schema は再 assert しない。
- shim 確認: Broad に `NewProduceEpisodeFromEnv` は含めない。一般論としても Broad = Adapter 合成、FromEnv / CLI 入口は System。Decision に固定。
- Verification: `check-static` / `test-unit`（coverage 92.6%）/ `test-integration` いずれも exit 0。pre-commit / pre-push 通過後 push。
- GitHub Issue なし（local 達成契約が正）。

### Commits

- `1c9f163`
- `17f7411`
