Issue draft。`create-issue` で正式化する前の一時置き場。

## 1. Summary

この Issue では、Drive 書込前に `contracts/manuscript.schema.json`・stem 一致・WAV 非空を Application で enforce し、検証後に `EpisodeWriter` を呼ぶ UseCase を実装する。

## 2. Context

- generator の `contracts/` 読み手は Application（`DESIGN.md` §2、layer-split decision）。
- Domain Error は `entities/errors/` に既存。
- 保存 Adapter は schema を import しない。検証をここへ寄せないと層分担が崩れる。

## 3. Canonical Sources

- `docs/decisions/2026-08-19T15-00-00-feature-generator-drive-adapter-layer-split.md` — generator の `contracts/` 読み手を Application にする判断
- `contracts/manuscript.schema.json` — 原稿 schema
- `contracts/drive-layout.md` — stem / 拡張子の対応
- `apps/generator/internal/application/port/episode_writer.go` — 呼び出し先 Port
- `apps/generator/internal/entities/errors/` — Domain Error
- `apps/generator/internal/infrastructure/drive/gdrive/writer.go` — 現在 infra にある検証 logic の移送元
- `architecture/backend/application` — UseCase / Port の責務
- `architecture/ports-adapters` — Application Port と fake / stub の境界
- test 方針 — testing-strategy skill

## 4. Scope

### In Scope

- Application 層の原稿検証（`contracts` + jsonschema）
- WriteEpisode UseCase（検証 → `EpisodeWriter.Write`）
- sociable unit（HTTP 0 回）
- depguard: `application` へ `contracts` / jsonschema allow
- 既存 infra validation の移送 / 削除

### Out of Scope

- Drive HTTP / OAuth
- cmd / GHA
- playback worker 読取検証

## 5. Contract

- schema 不適合 / stem 不一致 / 空 ID / 空 WAV → Domain Error、`EpisodeWriter` 未呼び出し
- 検証通過後のみ byte を Port へ渡す

## 6. Constraints

- Infrastructure（Drive 保存）が `contracts` を import しないこと
- JSON schema 適合・stem 一致・空値検証を 1 つの入口に集約し、呼び出し側が個別に再実装しない

## 7. Acceptance Criteria

- [ ] AC-1: 各 Domain Error case が unit で pass
- [ ] AC-2: 検証成功時のみ Fake `EpisodeWriter` が 1 回呼ばれる
- [ ] AC-3: generator unit gate / depguard が pass

## 8. Verification

```bash
cd apps/generator && go test ./internal/application/...
./scripts/test-unit.sh
./scripts/generator/check-static.sh
```

## 9. Dependencies

- 先行: Drive 保存 Adapter Issue（`EpisodeWriter` Port 実装）
- 後続: 原稿→TTS→書込 UseCase、cmd / GHA

## 10. Risks

- schema compile の置き場を誤ると `contracts` が validator engine を抱えるため、byte 公開と compile を分離したままにする


## 11. Notes

- 実装開始時の読む順: layer-split decision → `contracts/manuscript.schema.json` → `drive-layout.md` → `episode_writer.go` → `entities/errors/` → 現 `gdrive/writer.go`

