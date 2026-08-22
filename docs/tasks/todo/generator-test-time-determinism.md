## 1. Summary

このIssueでは、generator の X API Adapter Unit Test から current wall-clock への依存を除く。完了後、test input の時刻は固定され、実行時刻に関係なく同じ error 契約を検証できる。

## 2. Context

1. `GetXAPI` と `TwitterAPI.io` の vendor error test は `time.Now().UTC()` を `ItemSource.List` の入力に使う。
2. これらの case は vendor error 時に部分結果を返さない契約を検証する。current wall-clock は契約の入力値ではない。

## 3. Canonical Sources

1. `apps/generator/internal/infrastructure/x/getxapi/post_source_sociable_unit_test.go` — GetXAPI error test の current-time input
2. `apps/generator/internal/infrastructure/x/twitterapiio/post_source_sociable_unit_test.go` — TwitterAPI.io error test の current-time input
3. `scripts/generator/test-unit.sh` — generator Unit gate
4. `testing-strategy` skill — FIRST と Unit Test の規約

## 4. Scope

### In Scope

1. generator Unit Test 内の current wall-clock input を固定 UTC 時刻へ置換する
2. vendor error case が部分結果なしで error を返す既存契約を維持する

### Out of Scope

1. production code の時刻取得・時刻注入の追加
2. X API Adapter の HTTP / error contract 変更
3. fuzzing、mutation testing、condition coverage

## 5. Contract

vendor error response を返す adapter に固定 UTC 時刻を渡すと、`List` は nil result と error を返す。

## 6. Constraints

1. test input に `time.Now` を使わない。実行時刻で assertion または vendor request が変わらないこと
2. test 対象は Unit scope に留め、本番 credential と実 vendor API を使わない

## 7. Acceptance Criteria

1. [ ] AC-1: generator の Unit Test source に `time.Now` が存在しない
2. [ ] AC-2: GetXAPI の vendor error case は固定 UTC 時刻で nil result と error を検証する
3. [ ] AC-3: TwitterAPI.io の vendor error case は固定 UTC 時刻で nil result と error を検証する
4. [ ] AC-4: generator Unit gate が `-shuffle=on` と `-count=1` で pass する

## 8. Verification

```bash
rg 'time\.Now' apps/generator --glob '*_test.go'
./scripts/generator/test-unit.sh
```

最初のcommandは match なし、Unit gate は exit 0 を期待する。

## 9. Dependencies

なし。

## 10. Risks

固定時刻を request の期待値へ反映し忘れると、adapter の error 契約とは無関係な test failure になる。既存の vendor request assertion を確認する。

## 11. Notes

current wall-clock を必要とする production behavior はこのIssueの対象外。
