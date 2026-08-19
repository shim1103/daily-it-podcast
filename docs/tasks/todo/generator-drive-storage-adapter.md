Issue draft。`create-issue` で正式化する前の一時置き場。

## 1. Summary

この Issue では、generator Infrastructure に Google Drive 保存 Adapter を実装する。`EpisodeWriter.Write` が `{episodeId}.json` と `{episodeId}.wav` を `DRIVE_FOLDER_ID` folder 直下へ put できる。schema 検証と OAuth refresh は含まない。

## 2. Context

- 現 branch に monolith 実装（schema + OAuth + put 混在）がある。本 Issue で slim 化する。
- 層分担は `docs/decisions/2026-08-19T15-00-00-feature-generator-drive-adapter-layer-split.md`。
- Generator の `contracts/` 読み手は Application。保存 Adapter は `drive-layout.md` の配置だけを実装する。

## 3. Canonical Sources

- `docs/decisions/2026-08-19T15-00-00-feature-generator-drive-adapter-layer-split.md` — generator Drive 書込の層分担
- `contracts/drive-layout.md` — 配置・命名
- `apps/generator/internal/application/port/episode_writer.go` — Port 境界
- `apps/generator/internal/infrastructure/drive/gdrive/writer.go` — 現 monolith。分離元
- `apps/generator/internal/infrastructure/agentsecrets/proxy.go` — Body/Form inject の使い方
- `apps/generator/internal/infrastructure/secretnames/names.go` — `DRIVE_FOLDER_ID` の名前
- `architecture/backend/infrastructure` — Driven Adapter の責務
- `architecture/ports-adapters` — Application Port 所有 / Infrastructure 実装
- test 方針 — testing-strategy skill

## 4. Scope

### In Scope

- Drive REST（list / create / upload）と `EpisodeWriter` Port 実装
- infra 内 `TokenSource` interface + test fake
- Composition 結線、Adapter 隣 sociable unit（Drive HTTP stub）
- 既存 monolith から schema 検証と OAuth refresh を取り除く refactor

### Out of Scope

- `manuscript.schema.json` 検証（別 Issue）
- OAuth refresh 実装（別 Issue）
- Application UseCase / cmd / GHA

## 5. Contract

- Port は `EpisodeWriter.Write` を維持（`PutFile` Port に下げない）
- 成功時: `DRIVE_FOLDER_ID` folder 直下に `{episodeID}.json` と `{episodeID}.wav`（`drive-layout.md`）
- Port `@require` の schema 条項は caller 責務。Adapter は enforce しない

## 6. Constraints

- 秘密値を code に書かない（`README.md` の変数名のみ）
- Application Port を増やさない（`TokenSource` は infra 内）
- `EpisodeWriter` は `{episodeId}.json` / `.wav` の命名を持ってよいが、`manuscript.schema.json` の field 意味は持たない
- `client` は name / MIME / byte だけを扱い、episode 語彙を持たない

## 7. Acceptance Criteria

- [ ] AC-1: token stub 注入で json + wav が put される
- [ ] AC-2: Drive HTTP 失敗は Infrastructure Error
- [ ] AC-3: generator unit gate / depguard が pass

## 8. Verification

```bash
cd apps/generator && go test ./internal/infrastructure/drive/... ./internal/composition/...
./scripts/test-unit.sh
./scripts/generator/check-static.sh
```

## 9. Dependencies

- 先行: `port.EpisodeWriter`、AgentSecrets Form/Body inject
- 後続: Google OAuth Adapter Issue、Application 原稿検証 Issue

## 10. Risks

- list が folder を絞らない実装は同名 file 衝突に弱い（layout 契約上 create は `parents` で folder 指定）

## 11. Notes

- HTTP client は name / MIME / byte のみ。`EpisodeWriter` 実装が `drive-layout.md` の命名を map する。
- 実装開始時の読む順: layer-split decision → `drive-layout.md` → Port → 現 `writer.go` → AgentSecrets proxy test
