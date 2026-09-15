# feature(generator): R2 へ CompletedEpisodeLookup を本実装する

## 1. Summary

このIssueでは generator の R2（S3 互換）`CompletedEpisodeLookup` を本実装し、Composition が R2 照会へ結線できる状態にする。完了後、test / 結線口から表示 date に対する完成ペア有無を Port 契約どおり返せる。**NI の gate 正 peer は experimental local S3**（Decision `2026-09-15T12-02-48`）。**本番正本の切替は列 6（Writer と同着）**。

## 2. Context

1. 事実: A が `infrastructure/r2/lookup.go` stub・足場 test・`composition.newR2CompletedEpisodeLookup` を固定済み。
2. 事実: Port `CompletedEpisodeLookup.HasPair` と Drive 実装は既存。配置は `contracts/episode-layout.md`。
3. 事実: peer/gate の正は `2026-09-15T12-02-48`（`19-29-08` の任意条を supersede）。本番口 `11-04-30`。error `11-19-21`。
4. 事実: 現行本番 runtime は当面 Drive。本 PR 単独で本番正本を切替えない。
5. 事実: C1（`generator-r2-test-peer-scope`）が local S3 gate infra を持つ。本 Issue は Lookup **振る舞い**。
6. 運用: 1 Issue = 1 PR。

## 3. Canonical Sources

1. peer / gate: `docs/decisions/2026-09-15T12-02-48-feature-generator-r2-write-adapter.md`
2. 本番口: `docs/decisions/2026-09-14T11-04-30-docs-generator-r2-write-details.md`
3. error / retry: `docs/decisions/2026-09-14T11-19-21-docs-generator-r2-write-details.md`
4. 契約: `contracts/episode-layout.md` / `port/completed_episode_lookup.go` / `infrastructure/r2`（Writer・Lookup stub）
5. Drive 鏡像: `infrastructure/drive/gdrive/lookup.go`（挙動の参考。契約の正本ではない）
6. C1 infra: `docs/tasks/todo/generator-r2-test-peer-scope.md`
7. test 方針: testing-strategy（再掲しない）

## 4. Scope

### In Scope

1. R2 `CompletedEpisodeLookup` 本実装（A 足場 test を behavior test へ置換）
2. Composition 結線口の維持・必要なら整備（本番正本切替は列 6）
3. Sociable Unit: 分岐表（ペア / date / retry / 4xx / secret 非露出）
4. Narrow: **gate 正は experimental local S3**（C1 infra）。`httptest` 併存可
5. C1 完了後、Lookup NI を Writer NI と同じ local S3 gate に載せる

### Out of Scope

1. 本番 `newProduceEpisode` の Lookup/Writer を R2 へ切替（列 6）
2. local S3 / `getPlatformProxy` **infra 自体**の新規導入（C1）
3. Workers binding（playback）
4. playback 読取振る舞い（列 5）
5. System/E2E・OAuth 削除（列 7）
6. cache
7. Port signature 変更
8. BI への wrangler 必須化

## 5. Contract

1. `HasPair` 成功時の真偽は Port 契約どおり（同一 stem の json+mp3 があり json の `date` が一致するとき true。片方・無し・不一致は false）。
2. network / 5xx / 429 は有限 retry。その他 4xx は fail-fast（`11-19-21`）。
3. Application の error 写像は変えない。
4. Error message / log に bucket・key・Account ID・Access Key・secret 実値を載せない。
5. production Lookup に delete を公開しない。
6. credential / endpoint 形は Writer と同型。
7. Lookup NI の gate 正 peer は Writer NI と同じ experimental local S3。

## 6. Constraints

1. Application は R2 / S3 SDK を import しない。
2. Port を List/Get 単位に下げない。
3. 本 PR だけで本番正本を R2 にしない。
4. SU に wrangler を起動しない。BI に wrangler を必須化しない。
5. List 戦略は Drive 鏡像（全 key List → stem 集合 → 候補 json Get）を既定とし、layout 変更はしない。

## 7. Acceptance Criteria

1. [ ] A 足場 test が behavior test に置換されている
2. [ ] Sociable Unit で完成ペア true / 不完全・date 不一致 false / retry・4xx / secret 非露出が観測できる
3. [ ] Narrow が experimental local S3（C1 gate）で緑（C1 未完了なら依存でブロックし、黙って skip しない）
4. [ ] `httptest` 併存を捨てていない（残す場合）
5. [ ] Error message に bucket / key / secret 実値が含まれない
6. [ ] generator unit / static gate が緑
7. [ ] 本番正本がまだ Drive（Lookup/Writer とも未切替）である旨が Notes に明示されている

## 8. Verification

1. `scripts/generator/check-static.sh`
2. `scripts/generator/test-unit.sh`
3. C1 が定義する local S3 integration（Lookup NI）

## 9. Dependencies

1. 先行: Writer 本実装済み・A Lookup stub
2. 並行/先行推奨: `generator-r2-test-peer-scope`（local S3 gate infra）。Lookup NI の gate 緑は C1 に依存
3. 対: `playback-r2-read-adapter`（列 5）
4. 次: `r2-smoke-migrate-cutover`（列 6・Writer と同着結線）

## 10. Risks

1. List 全件＋逐次 Get は object 増で遅延する → 日次件数前提を維持し、早期最適化は YAGNI。
2. C1 より先に本 Issue だけ閉じると Lookup NI gate が未充足 → C1 依存を AC で明示する。

## 11. Notes

本番同着切替は列 6。OAuth 削除は列 7。peer/gate の正は `2026-09-15T12-02-48`。
