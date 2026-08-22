---
name: generator fuzzing は stdlib で PCM to WAV の local bounded fuzz から始める
date: 2026-08-22T17:57:00
branch: chore/generator-ci-test-configuration-hardening
---

## 1. Decision

1. generator fuzzing は Go stdlib の `testing.F` と `go test -fuzz` を使う
2. 初回 target は raw PCM を WAV byte へ変換する pure function とする
3. fuzzing は bounded local 実行だけにし、hook、pull request CI、scheduled CI には載せない
4. fuzz failure input は repository の seed corpus に保存し、通常の Unit Test の regression として実行する

## 2. Reason

1. Go stdlib fuzzing は dependency を増やさず、coverage-guided input generation と failure minimization を提供する
2. PCM to WAV conversion は byte input、失敗条件、output invariant が明確で、外部 I/O や credential を必要としない
3. unrestricted fuzzing は時間と CPU を消費する。初回は local bounded 実行で target と corpus の安定性を確認する
4. failure input を seed corpus に保存すると、偶発的に見つかった bug を決定的な regression test に変換できる

## 3. Rejected

1. random input generator を独自実装する案（stdlib fuzzing と coverage guidance を再実装する）
2. HTTP / credential 境界を初回 fuzz target にする案（external state が failure の再現性を壊す）
3. pull request CI で無制限 fuzzing を実行する案（feedback time と resource consumption を制御できない）
