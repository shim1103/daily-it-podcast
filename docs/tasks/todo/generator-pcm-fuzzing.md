## 1. Summary

このIssueでは、raw PCM から WAV byte への変換に Go stdlib fuzzing を追加する。任意の byte input で panic せず、failure input を決定的な regression corpus として残せる状態にする。

## 2. Context

1. PCM to WAV conversion は external I/O と credential を持たない byte boundary である。
2. empty PCM と 16-bit sample alignment は既存 Unit Test で error contract を持つ。

## 3. Canonical Sources

1. `docs/decisions/2026-08-22T17-57-00-chore-generator-ci-test-configuration-hardening.md` — fuzzing tool、target、実行場所、corpus の決定
2. `apps/generator/internal/infrastructure/speech/gemini/pcm_to_wav.go` — PCM to WAV contract
3. `testing-strategy` skill — Unit Test と regression の規約

## 4. Scope

### In Scope

1. PCM to WAV conversion の `testing.F` fuzz target
2. empty、odd-length、aligned PCM の seed corpus
3. bounded local fuzz entrypoint
4. fuzz failure input を repository corpus へ昇格する運用

### Out of Scope

1. hook、GitHub Actions、scheduled CI での fuzzing
2. HTTP、OAuth、vendor API、credential 境界の fuzzing
3. PCM to WAV 以外の fuzz target

## 5. Contract

1. 任意の PCM byte input は panic しない
2. empty または odd-length PCM は error を返す
3. non-empty aligned PCM が成功するとき、WAV header、data length、payload は input と整合する

## 6. Constraints

1. fuzz input の最大 size を制限し、local resource を無制限に消費しない
2. fuzzing は bounded duration で終了する
3. failure input は内容を確認してから repository corpus へ保存する

## 7. Acceptance Criteria

1. [ ] AC-1: PCM to WAV conversion に stdlib `testing.F` fuzz target がある
2. [ ] AC-2: empty、odd-length、aligned PCM の seed がある
3. [ ] AC-3: bounded local fuzz entrypoint が exit 0 で完了する
4. [ ] AC-4: fuzz target は input size 上限を超える input を処理しない
5. [ ] AC-5: fuzz failure input を corpus へ保存すると通常 Unit Test で再実行される

## 8. Verification

1. bounded fuzz entrypoint が exit 0
2. generator Unit gate が exit 0
3. 保存済み corpus が通常 Unit Test で実行される

## 9. Dependencies

なし。

## 10. Risks

fuzzing は CPU と memory を使う。bounded duration と input size 上限を守り、CI へ追加しない。

## 11. Notes

fuzz failure は最小化後も仕様 bug かを確認し、再現 input だけを corpus に保存する。
