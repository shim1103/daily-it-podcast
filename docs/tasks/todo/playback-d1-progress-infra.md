# feat(playback): 進捗D1 local peer本実装・NI・Fake・adapter・Root結線

## 1. Summary

このIssueでは、`createLocalD1Binding` を実 `getPlatformProxy` 本実装にし到達NIを緑化し、SU用Fakeを足し、`ProgressRepository` のD1 adapterとComposition Root切替まで完了する。Application merge／list embed／webは別Issue。

## 2. Context

1. 事実: Aは `LocalD1BindingHandle` と `createLocalD1Binding` stub（zero・proxy未起動）まで。
2. 事実: FakeはAではない（test時だけ要る）。本IssueのSU supportとして作る。
3. 事実: native binding adapterはHTTP NIを持たず、Fake／localのsociableが正（Decision `2026-09-15T14-28-12`）。
4. 事実: 旧planの「local peer」と「adapter」は同一完了境界にまとめる（NIはadapter検証にも要る）。

## 3. Canonical Sources

1. A: `apps/playback/test/support/create-local-d1-binding.ts`（型・stub signature）
2. A: `apps/playback/worker/src/infrastructure/d1/d1-database-binding.ts`・`progress-d1-constants.ts`・`ProgressRepository` Port
3. B: `docs/decisions/2026-09-22T19-29-15-feature-playback-progress.md`（adapter内retryしない）
4. B: `docs/decisions/2026-09-23T08-06-58-feature-playback-progress.md`（D1定数・binding面の置き場）
5. B: `docs/decisions/2026-09-15T14-28-12-feature-playback-r2-read-adapter.md`
6. B: `docs/decisions/2026-09-16T00-20-08-feature-generator-r2-test-peer-scope.md`（playback local peer置き場）
7. R2同型: `create-local-r2-binding.ts` / `local_r2_binding.narrow_integration.test.ts`
8. test方針: `testing-strategy`

## 4. Scope

### In Scope

1. `createLocalD1Binding` 本実装（`getPlatformProxy`・`remoteBindings: false`）。stubのzero戻りと申し送り`todo:`を除去
2. 到達NI（prepare／bind／run／firstが観測できる）
3. `createFakeLocalD1Binding`（ただのFake・proxyなし）とSU
4. D1 `ProgressRepository` adapter、RootのStub→本実装切替（bindingあり）
5. A足場testをbehaviorへ置換（該当分）
6. adapter／Rootまわりの申し送りcomment削除

### Out of Scope

1. 一過性dispatch・常設smoke拡張
2. create／update／complete／pull の merge／冪等本文（Application Issue）
3. list embed合成、web sync／UI
4. instance再登録

## 5. Contract

既存Port／`D1DatabaseBinding`／binding名を満たす。未固定の公開型を新設しない。

## 6. Constraints

1. adapter内でD1一時失敗をN回retryしない（B）。
2. PortにD1固有型を露出しない。
3. Fakeを本番treeの契約stubと混同しない（test/support）。
4. local peer起動失敗を黙ってFakeに落とさない。

## 7. Acceptance Criteria

1. [ ] `createLocalD1Binding` が実proxyで `EPISODE_PROGRESS` を返す
2. [ ] local D1到達NIが緑
3. [ ] FakeがproxyなしでSUに使える
4. [ ] D1 adapterがPort契約を満たし、Rootがbindingありで本実装を選ぶ
5. [ ] 該当A足場がbehavior testに置換され、申し送り`todo:`が消えている
6. [ ] adapter内retryをしていない

## 8. Verification

1. unit: Fake SU・adapter sociable（Fake／local）
2. integration: `local_d1_binding` 到達NI（R2同型でunitからexclude）
3. typecheck／既存playback unitが緑

## 9. Dependencies

1. blocked by: なし（A stub済み）。dispatch疎通（`playback-d1-dispatch-smoke`）は並列可。remote Verification強化に使えるが完了必須ではない。

## 10. Risks

1. local D1とremote TESTの差分 → NIはlocal、remoteはdispatch／release側

## 11. Notes

1. 誤ってAに置いたFakeがあれば本Issue所有へ移すか作り直す。
2. Application／embed／webは別達成契約。
