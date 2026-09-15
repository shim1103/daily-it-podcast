# feature(generator): R2 NI gate peer（experimental local S3）と playback local binding infra を導入する

## 1. Summary

このIssueでは Decision `2026-09-15T12-02-48` に従い、**generator の NI gate peer として wrangler experimental local S3 を導入・gate 化**し、**playback の local binding test 基盤（`getPlatformProxy`）を C1=infra として用意**する。完了後、該当 integration gate が local peer 経由で緑になる。振る舞い本実装（Lookup / 列 5 Adapter）は別 Issue。

## 2. Context

1. 事実: 本番口は `2026-09-14T11-04-30`（generator=S3、playback=binding）。
2. 事実: 先行 `2026-09-14T19-29-08` は local S3 を任意・非 gate とした。本 Issue の正は supersede 後の `2026-09-15T12-02-48`。
3. 事実: Writer 本実装・`httptest` NI は既存。Lookup 本実装は `generator-r2-completed-episode-lookup`。列 5 振る舞いは `playback-r2-read-adapter`。
4. 事実: C1=infra、列 5=behavior（Decision §1-7）。
5. 運用: 1 Issue = 1 PR。Verification は Issue 内で赤から緑（先出し専用 smoke Issue なし）。

## 3. Canonical Sources

1. peer / gate: `docs/decisions/2026-09-15T12-02-48-feature-generator-r2-write-adapter.md`
2. 旧 peer（一部 superseded）: `docs/decisions/2026-09-14T19-29-08-feature-generator-r2-write-adapter.md`
3. 本番口: `docs/decisions/2026-09-14T11-04-30-docs-generator-r2-write-details.md`
4. 実施順: `docs/decisions/2026-09-14T12-49-26-docs-generator-r2-write-details.md`
5. A stub（infra 入口）: `apps/generator/test/r2locals3/`（`Start` / `wrangler.local-s3.stub.jsonc`）
6. A stub（playback local binding 入口）: `apps/playback/test/support/create-local-r2-binding.ts`
7. test 方針: testing-strategy（再掲しない）
8. 対 Issue: `generator-r2-completed-episode-lookup` / `playback-r2-read-adapter`

## 4. Scope

### In Scope

1. generator: wrangler experimental local S3 を Writer NI の **gate 正 peer** として導入する（別 integration job / script）
2. `httptest` NI の併存維持（捨てない）
3. Lookup NI を同じ local S3 peer へ載せる準備、または Lookup Issue 完了後に載せる AC の固定（最終形は Writer+Lookup NI）
4. playback: `getPlatformProxy` による local binding **infra**（起動・注入・gate 配線）。列 5 が刺せる状態にする
5. lane / 必要なら薄い動線を新 Decision へ更新

### Out of Scope

1. R2 `CompletedEpisodeLookup` 振る舞い本実装（Lookup Issue）
2. playback `EpisodeRepository` 振る舞い本実装・本番正本切替（列 5 / 列 6）
3. BI への wrangler / local S3 必須化
4. SU への wrangler / `getPlatformProxy` 必須化
5. 本番 bucket 実到達

## 5. Contract

1. generator の **integration gate** は experimental local S3 を正 peer とし、Writer NI（および Lookup 本実装後は Lookup NI）がそれで緑になる。
2. generator Sociable Unit は wrangler を起動しない。
3. Broad は wrangler local S3 を必須 peer にしない。
4. playback local binding infra の正手段は `getPlatformProxy`（同等埋め込み proxy 可）。
5. 列 5 は本 Issue の infra の上で振る舞いを実装する（本 Issue は list/get 契約本体を満たさない）。

## 6. Constraints

1. Application / Worker Application は storage vendor SDK を不必要に増やさない（既存層規則に従う）。
2. experimental 欠如を黙って skip して gate 緑にしない（通すか、失敗を見える化する）。
3. `test-unit` に wrangler 起動を混ぜない。
4. 本 Issue だけで本番正本を R2 にしない。

## 7. Acceptance Criteria

1. [ ] Decision `2026-09-15T12-02-48` が peer/gate の正として lane から辿れる
2. [ ] generator に experimental local S3 を使う **別 integration** Verification があり、Writer NI がそれで緑
3. [ ] `httptest` 経路が併存している（削除していない）
4. [ ] Lookup NI を同 peer に載せる契約が Issue / Lookup Issue から辿れる（実装は Lookup 完了後で可）
5. [ ] playback 側に `getPlatformProxy` local binding infra があり、列 5 が利用可能な入口になっている
6. [ ] SU に wrangler / platform proxy 必須起動が無い
7. [ ] BI に wrangler local S3 必須化が無い
8. [ ] 既存 `check-static` / `test-unit`（または playback 相当 unit）が緑を維持

## 8. Verification

1. generator: 別 integration job/script（local S3）+ 既存 `scripts/generator/check-static.sh` / `scripts/generator/test-unit.sh`
2. playback: local binding infra を使う最小 Verification（列 5 前でも infra 単体で観測可能ならそれ）
3. Issue 内で赤から緑（専用先行 smoke Issue なし）

## 9. Dependencies

1. 先行: Writer 本実装、Decision `12-02-48`
2. 後続: `generator-r2-completed-episode-lookup`（振る舞い。NI peer は本 Decision）
3. 後続: `playback-r2-read-adapter`（列 5 振る舞い。infra は本 Issue）
4. 次: 列 6 `r2-smoke-migrate-cutover`

## 10. Risks

1. experimental local S3 が path-style Put/List を拒む → Issue 内で失敗が見える。黙って `httptest` だけ緑に戻して閉じない。
2. `getPlatformProxy` の wrangler version 差 → version を Verification で固定可能な範囲に閉じる。

## 11. Notes

旧 `19-29-08` の「任意・非 gate」は superseded。本番口・SU/BI 禁止・generator 非 binding・Lookup 同着は維持。
