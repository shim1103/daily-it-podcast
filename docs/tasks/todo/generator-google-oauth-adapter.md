Issue draft。`create-issue` で正式化する前の一時置き場。

## 1. Summary

この Issue では、Google OAuth refresh から access token を取得し、Drive 保存 Adapter へ infra 内 `TokenSource` として供給する。

## 2. Context

- Drive 保存 Adapter は token を外部から受け取る前提（`generator-drive-storage-adapter.md`）。
- 秘密キー名は `README.md`（`GOOGLE_OAUTH_*`）。AgentSecrets Form inject は済。
- refresh token の vendor 契約は Drive 保存と別 mechanism。layer-split decision に従い切り出す。

## 3. Canonical Sources

- `docs/decisions/2026-08-19T15-00-00-feature-generator-drive-adapter-layer-split.md` — OAuth を別トラックへ分離する判断
- `README.md` — 秘密の名前
- `apps/generator/internal/infrastructure/agentsecrets/` — Form inject
- `apps/generator/internal/infrastructure/secretnames/names.go` — OAuth secret name 定数
- `apps/generator/internal/infrastructure/drive/gdrive/writer.go` — 現 refresh 実装の分離元
- `architecture/backend/infrastructure` — vendor 認証 client の責務
- test 方針 — testing-strategy skill

## 4. Scope

### In Scope

- `infrastructure/google/oauth` — refresh 実装
- infra 内 `TokenSource` の本番実装
- sociable unit（token endpoint stub）
- `composition/gdrive.go` — stub を本番 TokenSource へ差し替え
- 現 `gdrive` 内 refresh 処理の移送 / 削除

### Out of Scope

- Drive list / create / upload
- schema 検証 / UseCase
- playback worker OAuth

## 5. Contract

- 成功時: 非空 access token を返す
- 失敗時: Infrastructure Error

## 6. Constraints

- Application Port を増やさない
- 秘密値を code に保持しない
- token endpoint の request shape は secret 名 inject に閉じ、caller が form field を組み立てない

## 7. Acceptance Criteria

- [ ] AC-1: token endpoint stub で Bearer が返る
- [ ] AC-2: 401 / 空 token は Infrastructure Error
- [ ] AC-3: generator unit gate / depguard が pass

## 8. Verification

```bash
cd apps/generator && go test ./internal/infrastructure/google/oauth/... ./internal/composition/...
./scripts/test-unit.sh
./scripts/generator/check-static.sh
```

## 9. Dependencies

- blocked by: Drive 保存 Adapter Issue（`TokenSource` 注入点）
- blocks: 本番 credential での Drive E2E（別 follow-up）

## 10. Risks

- token refresh の失敗理由が Drive upload failure と混ざる risk があるため、OAuth 側で operation を分けて Error 化する


## 11. Notes

- 実装開始時の読む順: layer-split decision → `README.md` の secret 名 → `secretnames/names.go` → `agentsecrets/proxy.go` → 現 `gdrive/writer.go`

