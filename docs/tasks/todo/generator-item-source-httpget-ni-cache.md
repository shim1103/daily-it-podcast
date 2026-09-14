# feat(generator): httpget helper・5 源実境界 NI・接続 cache suite

## 1. Summary

このIssueでは、5 情報源 Adapter の GET+retry / HTML 正規化を `infrastructure/httpget` へ抽出し、源ごと実境界 Narrow Integration を揃え、接続確認用 cache suite（CI 外）を置く。完了後、helper SU 1・源 NI 5・cache 1 file（5 case）の 7 test file が観測でき、`httptest` 合成 upstream NI は残らない。

## 2. Context

事実:
1. HackerNews / Lobsters / Publickey / TechCrunch / クラウド Watch の 5 Adapter に同型の GET+retry / HTML 正規化が重複している。
2. HN / Lobsters の現行 NI は `httptest` + DialTLS redirect（合成 upstream）。Publickey / TechCrunch / CloudWatch に同等の源 NI が無い。
3. 公開 feed / API は認証不要。接続確認に secret / `workflow_dispatch` は不要。
4. `.cache/` は `.gitignore` 済み。

仮定（作業を止めない）:
1. 接続 cache の CI 除外は既存の build tag パターン（System と同型）で足りる。
2. helper 公開 API の細名は実装時に A（code）へ固定する。

user 明示:
1. helper + 5 NI + 1 cache = 7 test file。1 Issue。
2. helper NI なし。cache は CI 除外。cache ≠ NI。
3. `httptest` NI は実境界 NI が吸収するため不要。
4. helper 置き場は `infra/httpget/`（`infra/rss/` にしない。各 source も `rss/` 配下へ移さない）。

## 3. Canonical Sources

1. Decision `docs/decisions/2026-09-14T13-06-11-feature-generator-textwriter-prompt-source-criteria-adapters.md` — helper / NI / cache の方針（B）
2. 1 媒体 1 Adapter・RSS facade 禁止 — Decision `2026-09-13T15-08-55` / `2026-09-02T14-41-00` / `14-41-01`
3. 各 Adapter — `apps/generator/internal/infrastructure/{hackernews,lobsters,publickey,techcrunch,cloudwatch}/`
4. 既存 NI（置換対象）— `apps/generator/test/{hackernews,lobsters}_narrow_integration_test.go`
5. Integration gate 入口 — `scripts/generator/test-integration.sh`
6. test 方針 — `skills/1:terms/testing-strategy`（再掲しない）
7. 層・Port — `skills/1:terms/architecture`（再掲しない）

## 4. Scope

### In Scope

1. `infrastructure/httpget` の導入と 1 Sociable Unit
2. 5 Adapter から重複 GET+retry / HTML 正規化を helper へ委譲（写像・URL・`SourceID`・方言 parse は各 Adapter に残す）
3. 5 源の実境界 Narrow Integration（HN/Lobsters の `httptest` NI を廃して置換。Publickey / TechCrunch / CloudWatch を新設）
4. 接続 cache suite 1 file・5 case（正常系のみ・実 HTTP・`.cache/` 保存）と CI / Integration gate からの除外

### Out of Scope

1. RSS 汎用 facade / `infrastructure/rss/` / 各 source の `rss/` 配下移設
2. `httpget` の Narrow Integration
3. `infra/feedcache` 等の production cache 層
4. TextWriter prompt の P1/P2、Application 先 fetch、Adapter HTML scrape
5. `MaxStoriesScanned` 値の再設計（別判断済みの契約値は触らない）
6. Broad / System / E2E の再設計

## 5. Contract

1. `httpget` は媒体非依存の GET+retry（および共有する HTML 正規化）だけを公開する。`ItemSource` / `SourceID` を知らない。
2. 各 Adapter は引き続き `port.ItemSource` を満たす。外向き契約（空 slice・feed 失敗は List ごと Error 等）は既存 Decision / Port を変えない。
3. 源 NI は実境界到達の I/O 契約を観測する。SU が所有する写像 exact・retry 表・件数上限を再 assert しない。
4. 接続 cache suite は各源 `List` 正常系の到達と cache 保存だけを観測する。Integration gate の既定実行対象に含まれない。

## 6. Constraints

1. 1 媒体 1 Adapter。RSS facade 禁止。
2. helper NI を置かない。
3. 接続 cache を NI file に混ぜない。production Adapter は cache を知らない。
4. 接続 cache を CI / `test-integration.sh` の既定収集に載せない。
5. 契約値・feed URL を Issue / Decision / DESIGN へ写して正本にしない。

## 7. Acceptance Criteria

- [ ] AC-1: `infrastructure/httpget` が存在し、専用 Sociable Unit が緑である
- [ ] AC-2: 5 Adapter が重複 GET+retry / HTML 正規化を helper 経由にし、媒体固有写像は各 package に残る
- [ ] AC-3: 5 源それぞれに実境界 Narrow Integration file があり、`httptest` 合成 upstream NI が残っていない
- [ ] AC-4: `httpget` 用 Narrow Integration file が存在しない
- [ ] AC-5: 接続 cache が 1 file・5 case（源ごと正常系）で、実 HTTP 結果を `.cache/` へ保存する
- [ ] AC-6: Integration gate（`scripts/generator/test-integration.sh` および同等 CI）の既定実行に接続 cache suite が含まれない
- [ ] AC-7: helper SU + 5 源 NI を含む関連 verification が緑である

## 8. Verification

```bash
cd apps/generator
go test ./internal/infrastructure/httpget/... \
  ./internal/infrastructure/hackernews/... \
  ./internal/infrastructure/lobsters/... \
  ./internal/infrastructure/publickey/... \
  ./internal/infrastructure/techcrunch/... \
  ./internal/infrastructure/cloudwatch/...

# Integration gate（接続 cache を含まないこと）
../../scripts/generator/test-integration.sh

# 接続 cache は local 明示実行のみ（build tag / 収集除外の正本は実装後の code・script）
```

coverage gate がある場合は既存 `scripts/generator/test-unit.sh` も pass。

## 9. Dependencies

blocked by: なし（5 源 `List` 本実装済み）

blocks: なし（TextWriter prompt P1/P2 は別 D）

## 10. Risks

1. 実境界 NI が外部 schema drift で赤くなる → 観測は I/O 契約に限り、写像詳細は SU に残す
2. 接続 cache が gate に混入する → build tag または script 収集除外を Verification で確認する

## 11. Notes

1. follow-up（本 Issue 外）: cache refresh の運用口の命名・手順文書化
2. HTML 正規化を `httpget` 同居か隣接 package かは、呼び出し頻度を見て実装時に A へ固定する
