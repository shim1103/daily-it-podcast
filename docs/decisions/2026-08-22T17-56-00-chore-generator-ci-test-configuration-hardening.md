---
name: generator mutation testing は mutest を local 専用で使う
date: 2026-08-22T17:56:00
branch: chore/generator-ci-test-configuration-hardening
---

## 1. Decision

1. generator の mutation testing は `mutest` `v0.6.0` を使う
2. 実行は local 専用とし、hook、GitHub Actions、score threshold を置かない
3. 対象は比較・等価演算子が仕様を表す generator の boundary logic に限定する
4. survived mutant は assertion不足の候補として review し、仕様上等価な mutant は test 追加の対象にしない

## 2. Reason

1. `mutest` は比較・等価演算子に対象を絞るため、初回の mutation 実行で signal と実行 cost を抑えられる
2. mutation score は assertion品質の十分条件ではない。CI threshold にすると、score のための test を追加する誘因になる
3. local 実行なら開発 feedback を妨げず、survived mutant を個別に仕様と照合できる

## 3. Rejected

1. `gomutants` の広い mutator set を初回から使う案（tool と mutant の種類が増え、初回の原因分解が難しくなる）
2. mutation score を CI / hook の必須 threshold にする案（実行 cost と equivalent mutant を品質 gate に混ぜる）
3. survived mutant ごとに必ず test を追加する案（仕様変更または等価 mutant を誤って固定する）
