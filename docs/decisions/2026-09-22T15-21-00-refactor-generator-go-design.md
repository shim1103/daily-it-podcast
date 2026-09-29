---
name: cmd は runCLI で defer を完了させてから os.Exit を一度だけ呼ぶ（成功時も Exit）
date: 2026-09-22T15-21-00
branch: refactor/generator-go-design
---

## 1. Decision

1. `apps/generator/cmd/generator` の本体は `runCLI() int` に置き、`signal.NotifyContext` の `stop` を `defer` する。
2. `main` は `os.Exit(runCLI())` のみとする。成功（0）も失敗（非0）も **Exit はここ一回**。
3. 成否の正本が OS exit code と stderr であることは先行 Decision（`2026-08-26T14-42-16-feature-generator-cmd-entrypoint.md`）を維持する。本 Decision はその実装形を、defer と両立する形へ更新する。

## 2. Reason

1. Go の `os.Exit` は呼び出しスタックを戻らず、**その関数に積んだ `defer` を実行しない**。失敗経路で途中 `os.Exit(1)` すると、signal 監視解除（`stop`）が走らない。
2. 先行 Decision は「成功時は `main` 自然終了、失敗時だけ `os.Exit`」と書いた。これは exit 観測としては足りるが、失敗時に defer が飛ぶ点を構造で潰せていない。本体を `return code` にし、`main` で一度 Exit すれば、成功・失敗どちらでも defer が先に走る。
3. 成功時に `os.Exit(0)` 相当になることは、OS 観測面を壊さない。変えるのは runtime の「自然終了」へのこだわりではなく、**後始末と exit の順序**である。

## 3. Rejected

1. **失敗時だけ途中で `os.Exit`、成功は `main` 自然終了のまま**（先行 Decision の実装形） — 失敗時に `defer stop()` が実行されない。
2. **`atexit` 相当や signal 専用の別 cleanup 経路を増やす案** — 入口を厚くし、Go の `defer` 慣例から外れる。
3. **exit code を細分化する案** — 先行 Decision Rejected と同旨。入口を厚くする。
