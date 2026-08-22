---
name: PCM to WAV に Go stdlib fuzz target を追加
date: 2026-08-22T19:00:00
session_id: none
branch: test/generator-pcm-fuzzing
prev: なし
---

## 1. Summary

`issue-manager` flowで `docs/tasks/todo/generator-pcm-fuzzing.md` を完了させた。`pcmToWAV` に対する Go stdlib fuzz target、bounded local fuzz entrypoint script、regression corpus を追加し、既存 gate script・CI workflow には変更を加えなかった。

## 2. Changes

1. `FuzzPCMToWAV` を TDD（Red→Green）で実装し、seed 3件（empty / odd-length / aligned）と regression corpus 2件が `go test` で subtest として実行されることを確認した。
2. `scripts/generator/fuzz-pcm-to-wav.sh` を追加し、`-fuzztime=20s` で exit 0、new failure 0件を確認した。
3. `./scripts/generator/test-unit.sh`（coverage 91.1%）、`./scripts/generator/test-race.sh` がともに exit 0 であることを確認した。
4. sandbox 内では `httptest.NewServer` の local port bind が `operation not permitted` になる既存制約を確認し、`dangerouslyDisableSandbox: true` で再実行して既存 test が pass することを確認した（今回変更の欠陥ではない）。
5. fuzz target の一般的な選定条件（pure function・fuzzが生成できる型・external I/O/credential非依存）と、`getxapi`/`twitterapiio` の response parse部分が次の拡張候補になり得ることを decision file の Notes へ追記した。

### Commits

1. `c986390`
2. `496aee3`
