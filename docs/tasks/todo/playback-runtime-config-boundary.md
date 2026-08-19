## 1. Summary

このIssueでは、Playback WorkerのDrive runtime config検証と、Playback Web ClientからWorker endpointを呼ぶ境界を確定する。
完了後は、secretを必要なruntimeへだけ注入し、設定漏れをInMemoryへの暗黙fallbackとして隠さない状態にする。

## 2. Context

1. Playback WorkerはCloudflare Workersの`env`を`fetch(request, env)`で受け取り、Composition Rootでrepositoryを選択する。
2. Workerが必要とするDrive configはOAuth 3 keyと`DRIVE_FOLDER_ID`の4 keyである。
3. GeneratorとPlayback WorkerとPlayback Webは、runtime・言語・secret注入経路が異なる。
4. `InMemoryEpisodeRepository`を本番Workerのsecret未注入時に暗黙選択すると、設定障害を空データの成功応答として隠す。

## 3. Canonical Sources

1. `docs/decisions/2026-08-19T17-37-00-playback-runtime-secret-boundary.md` — runtime別secret責務
2. `docs/decisions/2026-08-19T17-38-00-playback-repository-selection.md` — repository選択とfallback方針
3. `apps/playback/worker/src/composition/root.ts` — 現在の`PlaybackEnv`とrepository選択
4. `apps/playback/contracts/` — Worker HTTPのpath・schema・error契約
5. `docs/tasks/todo/playback-web-api-client.md` — Playback Web ClientのHTTP境界
6. `README.md` — secret名の運用inventory
7. `/architecture` — 層と依存方向
8. `/philosophy` — DRY・KISS・Design by Contract

## 4. Scope

### In Scope

1. Playback Worker内に必要なDrive key定義と`PlaybackEnv`の責務境界を確定する。
2. `undefined`と空文字を含むconfig completenessの判定を確定する。
3. Production Workerとlocal / unit testのrepository選択経路を分離する。
4. Composition Rootがconfig completenessを検証し、Route HandlerとControllerへsecret責務を漏らさない。
5. Playback Web ClientがWorkerの`baseUrl`を使ってHTTPを呼び、Drive secretを受け取らない境界を確定する。
6. Worker設定不足、明示的なInMemory選択、Drive選択、Clientのendpoint呼び出しを検証する。

### Out of Scope

1. GeneratorのOAuth実装変更。
2. Cloudflare本番secretの実値登録。
3. Google OAuth tokenの有効性確認やDrive APIの実通信。
4. Playback WebのUI、ViewModel、page実装。
5. repo全体で共有するsecret schemaの新設。

## 5. Contract

1. Playback Workerは必要な4 keyを自身のruntime configとして扱う。
2. Production Workerで必要keyが不足する場合、repositoryをInMemoryへ暗黙fallbackせず、外向きには`503`系の設定エラーを返す。
3. InMemoryは明示的なlocal development / unit test経路からだけ選択できる。
4. Playback Web ClientはWorker endpointの`baseUrl`だけを入力に取り、Drive secretを保持・送信しない。
5. Generatorのsecret定義とPlayback Workerのsecret定義は別runtimeの契約として独立する。

## 6. Constraints

1. Composition Root・Controller・Use Case・Infrastructureの依存方向を`/architecture`に従わせる。
2. Route Handlerへsecret名やrepository選択の判断を追加しない。
3. `README.md`のsecret一覧をruntimeの実行時検証へ流用しない。
4. secretの値をsource code・test fixtureの実値・client bundleへ含めない。
5. `playback-web-api-client.md`と責務を重複させず、Client実装の詳細は同Issueを参照する。

## 7. Acceptance Criteria

1. [ ] Workerが必要とする4 keyと、その注入元がWorker内の契約として明記されている。
2. [ ] `undefined`・空文字・一部欠落が設定不足として判定される。
3. [ ] Production Workerの設定不足がInMemoryの正常応答へfallbackしない。
4. [ ] InMemory選択がlocal development / unit testの明示経路に限定されている。
5. [ ] OAuth tokenの有効性確認がComposition RootではなくAdapter / Integration境界にある。
6. [ ] Playback Web ClientがDrive secretなしでWorker endpointを呼べる。
7. [ ] Generator・Playback Worker・Playback Webのruntime別secret責務が重複定義されていない。
8. [ ] Worker unit test、Playback Web Client unit test、typecheckがpassする。

## 8. Verification

```bash
./scripts/playback/test-unit.sh
./scripts/playback/check-static.sh
```

Playback Web Clientのtest commandは、実装済みtoolchainに合わせて確定する。

## 9. Dependencies

1. `playback-web-api-client.md` — ClientのHTTP契約
2. Cloudflare Workersのenv注入設定 — 本番確認時に必要
3. Google OAuth / Drive AdapterのIntegration test — token有効性の確認に必要

## 10. Risks

1. 本番設定不足を早期に検出できず、空データを正常応答するrisk。設定不足を明示的な`503`へ変換して防ぐ。
2. Drive secretがPlayback Webへ流出するrisk。Clientの入力を`baseUrl`に限定して防ぐ。
3. runtimeごとのsecret定義が重複して不整合になるrisk。各runtimeのSSOTを分離し、READMEはinventoryに限定する。

## 11. Notes

1. secret nameの全runtime共通SSOTは作らない。
2. `PlaybackEnv`はPlayback WorkerのComposition Root境界の型として扱う。
3. `docs/decisions/2026-08-19T17-37-00-playback-runtime-secret-boundary.md` と `docs/decisions/2026-08-19T17-38-00-playback-repository-selection.md` を判断のSSOTとする。
