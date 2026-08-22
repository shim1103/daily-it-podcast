---
name: generator CI は build・順序非依存 Unit・race を Go 1.26.6 で gate する
date: 2026-08-22T16:50:00
branch: chore/generator-ci-test-configuration-hardening
---

## 1. Decision

1. generator module と GitHub Actions runner は Go `1.26.6` を使う
2. generator static gate は既存 linter に加え `go build ./...` を実行する
3. generator Unit gate は statement coverage 90% と `-covermode=atomic` を維持し、`-shuffle=on` と `-count=1` を使う
4. generator Unit package の `go test -race` は GitHub Actions の Unit gate 後に実行する。local pre-commit / pre-push には追加しない
5. hook と GitHub Actions は project script を caller とし、toolchain command を重複して持たない

## 2. Reason

1. Go version を module、local、remote で揃えると language feature、standard library、静的解析、module 解決の差異を防げる
2. `go build` は test と別に production package の build を確認する
3. randomized order と cache 無効化は、通常の Unit 実行では隠れうる順序依存と stale result を検出する
4. race detector は実行 cost が高いため、短い local feedback を保ったまま clean remote runner で検査する
5. script を唯一の実行入口にすると、local と remote の品質契約を DRY に維持できる

## 3. Rejected

1. Go `1.22` を CI のまま維持する案（local Go `1.26.6` と実行環境が不一致になる）
2. `go build` を開発者の手動確認だけにする案（commit / CI gate で production build を保証できない）
3. race detector を pre-commit / pre-push に置く案（通常の local feedback の cost を不必要に増やす）
4. GitHub Actions YAML に toolchain command を直接書く案（project script との手順重複になる）
