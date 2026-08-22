---
name: Generator の秘密 transport 選択を Composition へ固定する
date: 2026-08-22T18:35:00
branch: feature/generator-infras-all-narrow-integration
---

## 1. Decision

1. Generator の HTTP 秘密注入と command 起動は、それぞれ `secrettransport` と `commandlaunch` の契約へ依存する。Adapter は AgentSecrets の concrete 実装、秘密名、project dir、child env allowlist、runtime 選択を所有しない。
2. Composition は opaque な秘密参照と秘密名の binding、Cursor 用 AgentSecrets project、child env allowlist、runtime implementation の選択を所有する。A の固定契約は `apps/generator/internal/infrastructure/secrettransport/contract.go`、`apps/generator/internal/infrastructure/commandlaunch/contract.go`、`apps/generator/internal/composition/secret_bindings.go` を正とする。
3. HTTP request と command invocation は秘密値を持たない。外部境界の error は秘密値、request body、stdin、child stderr 本文を含めない。
4. local の AgentSecrets proxy / exec wrapper と future process-env runtime は、同じ契約の implementation として差し替える。

## 2. Reason

1. vendor Adapter が concrete runtime と秘密名を持つと、runtime を追加するたびに Adapter 自身の責務が増える。選択を Composition へ集約すると、vendor I/O と credential runtime の変更理由を分離できる（`philosophy` §2-1、§3-1）。
2. opaque な参照を transport へ渡し、秘密名の解決を Composition と runtime implementation の境界へ閉じることで、Adapter が秘密名・値のいずれにも依存しない。これは Least Privilege に従う（`philosophy` §4-3）。
3. HTTP body と command stdin は外部境界へ渡る data であり、error へ転写すると secret と結合して漏洩経路になる。error の観測境界から除外することで、実装ごとの sanitize 忘れを防ぐ。
4. local と production の実行方法は異なるが、Adapter が必要とするのは「秘密参照付き HTTP」と「秘密を渡さない command 起動」だけである。同一契約へ実装を差し替えることで、runtime 差分を vendor Adapter へ波及させない。

## 3. Rejected

1. Adapter が `agentsecrets.Client`、AgentSecrets project dir、秘密名を直接所有する案。runtime 選択と vendor I/O が同居し、future runtime 追加時に複数 Adapter を変更する。
2. 秘密名または秘密値を command argv、stdin、error へ載せる案。argv / stdin / error の観測範囲を広げ、秘密境界を壊す。
3. runtime ごとに vendor Adapter を複製する案。HTTP protocol の同じ知識が runtime 数だけ重複し、Adapter 間の振る舞いが乖離する。
