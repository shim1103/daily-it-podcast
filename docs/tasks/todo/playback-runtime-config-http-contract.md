## 1. Summary

このIssueでは、Playback Workerのruntime config不備を外部service一時不能の`unavailable`から分離し、専用のHTTP error contractとして扱う。
設定不備と外部依存障害をClientが区別できる状態にする。

## 2. Context

1. 現在のPlayback Workerは、runtime config不備を`UnavailableError`へ変換し、`503`と`unavailable`を返す。
2. HTTP boundaryは`External Error`をstatusとerror codeへ変換する。
3. HTTP statusの原則では、外部service一時不能は`503`、server自身の設定不備は`500`に分類する。

## 3. Canonical Sources

1. `docs/decisions/2026-08-20T12-57-00-playback-runtime-config-validation.md` — runtime config Errorの内部境界
2. `apps/playback/contracts/external-errors.ts` — External Errorの契約
3. `apps/playback/contracts/http.ts` — HTTP error codeの契約
4. `apps/playback/worker/src/routes/http-error-response.ts` — External ErrorからHTTP responseへの変換
5. `apps/playback/worker/src/composition/runtime-config.ts` — runtime configの内部Error
6. `/architecture` と `/http-boundary` — Error境界とHTTP status分類

## 4. Scope

### In Scope

1. runtime config不備を表すExternal ErrorとHTTP error codeを定義する。
2. runtime config不備を`500`系の専用error codeへmappingする。
3. `unavailable`を外部service一時不能専用として維持する。
4. WorkerとPlayback Web Clientのcontract testを更新する。

### Out of Scope

1. runtime configのkey定義・completeness判定の変更。
2. Cloudflare secret / var bindingの本番登録。
3. Google OAuth tokenの有効性確認やDrive API実通信。
4. Generator・Playback Webのruntime config共有schema。
5. UIのError表示。

## 5. Contract

1. runtime config不備は`configuration_error`として外部へ返す。
2. runtime config不備のHTTP statusは`500`とする。
3. 外部service一時不能は`503`と`unavailable`のまま維持する。
4. response bodyは`{ code }`だけを返し、内部messageやsecret値を含めない。
5. server logには内部Errorのcause chainとdiagnostic messageを残す。

## 6. Constraints

1. `PlaybackRuntimeConfigError`はruntime config内部Errorとして維持する。
2. HTTP contract変更は`apps/playback/contracts/`をSSOTにし、READMEへ重複定義しない。
3. Error mappingはHTTP boundaryへ集約し、ControllerとComposition RootへHTTP status判断を追加しない。
4. secretの値をsource code、test fixture、response、logへ含めない。

## 7. Acceptance Criteria

1. [ ] `configuration_error`がHTTP error code schemaへ追加されている。
2. [ ] runtime config不備が`500`と`configuration_error`へ変換される。
3. [ ] Drive API一時不能が`503`と`unavailable`のまま変換される。
4. [ ] WorkerとClientのcontract testが両方passする。
5. [ ] response bodyに内部messageとsecret値が含まれない。

## 8. Verification

```bash
./scripts/playback/test-unit.sh
./scripts/playback/test-integration.sh
./scripts/playback/check-static.sh
```

## 9. Dependencies

1. `playback-runtime-config-boundary.md` — runtime config内部Error境界（完了）
2. `apps/playback/contracts/` — HTTP error contract
3. `playback-web-api-client.md` — Client側のerror code解釈

## 10. Risks

1. Clientが新codeを未対応のまま扱うrisk。WorkerとClientのcontract testを同時に更新して防ぐ。
2. 設定不備を外部service障害と誤認するrisk。statusとcodeを分離して防ぐ。

## 11. Notes

1. 現在の`unavailable`互換を維持するか、migration期間を設けるかは実装時に確定する。
