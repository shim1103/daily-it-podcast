## 1. Summary

`application/build.ManuscriptDraftFromWriterOutput` を実装し、TextWriter 戻り string（JSON wire）を Domain validation 経由で `ManuscriptDraft` へ解釈する。

## 2. Context

1. wire 形の正本は `entities/models.WriterOutput`、parse 规则の正本は `manuscript_draft_limits.go` である（Decision `14-11`、`15-00`）。
2. 失敗は Domain Error（`invalid_manuscript_draft`）。Infra Error にしない。
3. 本体は stub のまま。

## 3. Canonical Sources

1. `docs/decisions/2026-08-29T14-11-00-docs-produce-episode-run-spec-manuscript-draft-parse-domain-rules.md`
2. `docs/decisions/2026-08-29T15-00-00-docs-produce-episode-run-spec-writer-output-json-wire.md`
3. `apps/generator/internal/application/build/draft_from_writer.go`
4. `apps/generator/internal/entities/constants/manuscript_draft_limits.go`
5. `apps/generator/internal/entities/models/writer_output.go`
6. `apps/generator/internal/entities/models/manuscript_draft.go`

## 4. Scope

### In Scope

1. JSON unmarshal → limits 定数どおりの field / topic 数 / total 文字数検証。
2. 各朗読 field の日本語含有・trim 後非空・末尾 `。` 検証。
3. 成功時 `ManuscriptDraft`（`Title` 含む）を返す。
4. table test（valid と主要 reject 経路）。

### Out of Scope

1. limits 数値・Prompt 文案のチューニング（D）。
2. `WriteEpisode` の schema validation（Gate）。
3. markdown code fence 除去等の vendor 固有前処理（必要なら最小限のみ、新 B なしで A wire に従う）。
4. `ProduceEpisode.Run`（D）。

## 5. Contract

1. Infrastructure・vendor envelope を知らない。
2. parse 规则の SSOT は constants + models。Decision 本文へ数値を再掲しない。
3. 失敗 Op は `invalid_manuscript_draft`。

## 6. Constraints

1. A/B で固定されていない error 型・field を新設しない。
2. Gate 相当の schema 検査を二重実装しない。

## 7. Acceptance Criteria

1. [ ] `ManuscriptDraftFromWriterOutput` が panic stub でない。
2. [ ] valid wire が `ManuscriptDraft` へ変換される。
3. [ ] 句点・日本語・range・topic 数・total の reject 経路が test で cover される。
4. [ ] `go test ./internal/application/build/...` が pass する。

## 8. Verification

```bash
cd apps/generator
go test ./internal/application/build/... -count=1 -run ManuscriptDraft
go build ./...
```

## 9. Dependencies

1. A/B artifact（`WriterOutput`、`manuscript_draft_limits`、`draft_from_writer.go` stub）。
2. `ComposeBrief` 実装は不要（並行可）。

## 10. Risks

1. code fence 対応を広げすぎると vendor 依存が Application/build に漏れる。trim + json.Unmarshal の最小で足りるか test で確認する。

## 11. Notes

1. GitHub Issue 化は別判断。本 file が達成契約の正。
