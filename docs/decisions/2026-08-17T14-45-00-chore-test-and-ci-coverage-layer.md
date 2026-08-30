---
name: generator Unit に statement coverage 90% と depguard 層依存 block を入れる
date: 2026-08-17T14:45:00
branch: chore/test-and-ci
---

## 1. Decision

1. generator の Unit gate（`scripts/test-unit.sh` / pre-commit）に statement coverage **90%** を入れる。Branch Coverage は Go 標準 `go test -cover` が取らないため **暫定 statement**（skill の Branch 基準は将来見直し）
2. 除外は Composition Root（`internal/composition/**`）と TwitterAPI.io の薄い Error method（`internal/infrastructure/x/twitterapiio/error.go`）
3. 層依存 block は `golangci-lint` + `depguard`（depguard のみ enable、`list-mode: strict` の allow）。Infrastructure の Application 側許可は `application/port` prefix のみ。original 層 checker script は作らない
4. Integration / GHA / playback には coverage・layer lint を載せない（既存 Pyramid と docs-split を維持）
5. 本 decision は `2026-08-15T23-17-00-chore-test-and-ci` の「coverage を入れない」を上書きする

## 2. Reason

1. testing-strategy の Unit threshold と Scope 分離に合わせる
2. architecture backend の import 規則を機械 enforce する
3. Go 選定の merit（stdlib test + ecosystem linter）を使い、独自解析 script を増やさない（Least Power）

## 3. Rejected

1. cover / layer を Integration や GHA に載せる案（gate 二重化）
2. 層検査の original shell script（ecosystem を捨てて再発明する）
3. golangci で多数 linter を一括 enable する案（YAGNI）
