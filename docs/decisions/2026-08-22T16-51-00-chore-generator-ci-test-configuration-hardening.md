---
name: Go CI の再利用可能な判断は platform skill、generator の timing は project docs が所有する
date: 2026-08-22T16:51:00
branch: chore/generator-ci-test-configuration-hardening
---

## 1. Decision

1. Go toolchain、`go test` option、standard coverage の能力境界は `2:platform/go/` に置く
2. `2:platform/go/SKILL.md` は導線と責務を持ち、詳細は `ci-checks.md` に置く
3. generator の Go version、coverage threshold、hook / GitHub Actions timing、script path は project の `DESIGN.md`、`README.md`、script、workflow が所有する
4. scope-split の B は `logging` の分割基準で記録先を決め、再利用可能な判断を Decision Record として保存する

## 2. Reason

1. Go 標準 tool の能力・制約は複数 project で再利用でき、既存 platform skill に Go の同種 SSoT はなかった
2. platform skill に project 固有設定を混ぜると、変更理由と保守者が異なる知識を同じ file に置くことになる
3. `SKILL.md` を入口に絞り、`ci-checks.md` を詳細の SSoT にすると導線と本文を重複しない
4. 実装後の判断を Decision Record に残すと、次の agent が code だけから復元できない選択理由と Rejected を参照できる

## 3. Rejected

1. Go CI の知識を generator の project doc だけに置く案（Go 固有で再利用可能な判断を project に閉じ込める）
2. generator の timing と threshold を platform skill に置く案（project 固有設定が shared knowledge に混入する）
3. B を implementation 後の doc 同期だけとして扱う案（再利用可能な判断と Rejected が記録されない）
