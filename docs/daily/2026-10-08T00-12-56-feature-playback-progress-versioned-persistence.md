---
name: Issue #218 の版つき永続（D1 adapter・In-memory）を実装し、Port・cursor・UseCase 境界の規約を実装へ揃えた
date: 2026-10-08T00:12:56
session_id: 3ac7cc4e-cd0c-4e65-ac14-de5549825b99
branch: feature/playback-progress-versioned-persistence
prev: 2026-10-02T19-25-54-feature-playback-progress.md
---

## 1. Summary

Issue #218「進捗の版つき永続（D1 adapter・In-memory）」を `fulfill` で実装した。査読と shim の指摘を受けて、条件付き書込の 2 面化、Port の引数・戻り値の型の規約、cursor の型、UseCase 境界の規約への移行まで広げた。pull の契約は `since` のままで、cursor 版への切替（#220）は含めない。

## 2. Changes

1. D1 adapter の NI は実 local D1 で 59 tests が緑だった。同じ期待値の並行 10 本で勝者は 1 だった。
2. sandbox では miniflare が `~/Library/Preferences/.wrangler` に書いて EPERM になった。`XDG_CONFIG_HOME`・`WRANGLER_LOG_PATH`・`WRANGLER_REGISTRY_PATH` を `$TMPDIR` 配下へ向けると sandbox 内で通った。
3. pre-commit が generator の golangci-lint で落ちた。別 worktree（`tab1`）を指す古い cache が原因で、`golangci-lint cache clean` で解消した。変更の欠陥ではない。
4. shim の指摘で次を変えた（理由と Rejected は Decision が正）。
   1. 条件付き書込を、期待値 null で INSERT と UPDATE を兼ねる 1 面から、`insertProgressIfAbsent` と `replaceProgressIfVersion` の 2 面へ分けた（Decision `2026-10-01T18-54-16`）。
   2. Port の全 method は引数を `XxxArgs`、戻り値を `XxxResult` で export する規約へ揃えた。skill 側の更新は未適用。skill の正本 `~/settings/agent-standards` は sandbox 外で読めず、`.claude/skills` は生成物のため、更新案を scratchpad の `coding-terms-update.md` に置いた。shim が正本へ適用する必要がある。
   3. cursor は contract の schema が 10 進文字列（15 桁以下）を固定しており、不透明ではなかった。「10 進非負整数の文字列」と書き、Port 以下は number にした。文字列から数値への変換と省略時の起点は route の zod が担う（Decision `2026-10-01T23-32-40`、`2026-08-25T18-42-00-chore-playback-worker-web-layer`）。
   4. UseCase 境界が規約に沿っていなかった（`UseCaseInput` が 0 件、application の 11 file が `apps/playback/contracts` を import、`audioRef` が application に入っていた）。全 use case を規約へ移行した。dependency-cruiser の `worker-application-no-http-contracts` で本番 code の import を機械的に禁止した。test は層 lint の対象外にした（shim の判断。Decision `2026-08-25T18-42-00-chore-playback-worker-web-layer`）。
5. D1 NI の「完走済み→未完走で置換」は、adapter が merge しないことを検出するための意図的に不自然な入力で、自然な入力の test と 2 つに分けた。
6. 未確認として残した点は wiki #192 の Q10・B19・B20 と B16 の更新へ移した。旧面 `upsertProgress` の行が remote D1 にあるかは未確認。
7. Issue #218・#219・#220 の本文は更新済み。
8. `6273888`（Decision 更新）と `dccd15f`（コード一式）はこの記録の時点で未 push。

### Commits

- `625f7f0`
- `0bb98aa`
- `75b6d16`
- `e807294`
- `6273888`
- `dccd15f`
