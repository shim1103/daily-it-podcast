## 1. Summary

`application/build.ComposeBrief` を A/B どおり実装し、Fetch 結果から TextWriter へ渡す brief 平文 1 本を組み立てる。固定 Prompt 文案は触らない。

## 2. Context

1. brief 組立の所有と Prompt 配置は Decision で確定している（`docs/decisions/2026-08-29T16-30-00-*`、`17-00-00-*`）。
2. 数値 placeholder 注入 helper（`embedManuscriptDraftLimits`）は既に存在する。
3. `ComposeBrief` 本体と `{{SOURCES}}` / `{{JSON_EXAMPLE}}` 組立は stub のまま。

## 3. Canonical Sources

1. `docs/decisions/2026-08-29T16-30-00-docs-produce-episode-run-spec-brief-compose-build.md`
2. `docs/decisions/2026-08-29T17-00-00-docs-produce-episode-run-spec-brief-prompt-field-limits-merge.md`
3. `apps/generator/internal/application/build/brief.go`
4. `apps/generator/internal/application/build/brief_limits_embed.go`
5. `apps/generator/internal/entities/constants/text_writer_brief_prompt.go`
6. `apps/generator/internal/entities/models/writer_output.go`

## 4. Scope

### In Scope

1. `embedManuscriptDraftLimits` → `{{SOURCES}}` 平文列挙 → `{{JSON_EXAMPLE}}` の順で `TextWriterBriefPrompt` を置換する。
2. `{{SOURCES}}` は各 `SourceItem` の `SourceID`・`OccurredAt`・`Context` を平文列挙する（窓幅説明なし）。
3. trim 後非空の brief 1 本を返す。
4. sociable unit / table test を追加する。

### Out of Scope

1. `TextWriterBriefPrompt` 文案・`manuscript_draft_limits` 数値のチューニング（D）。
2. `TextWriter.Write` 呼び出し、`ManuscriptDraftFromWriterOutput`、TTS、`ProduceEpisode.Run`（別 Issue / D）。
3. OpeningGreeting / ClosingFarewell の brief への混入。

## 5. Contract

1. Prompt 散文を `application/build` に hardcode しない。
2. 数値 SSOT は `manuscript_draft_limits.go` のみ。文案組立で limits prose を新設しない。
3. `Context` を structured parse しない。

## 6. Constraints

1. A/B で固定されていない Port・型・定数を新設しない。
2. 過去 Decision Record を変更・削除しない。

## 7. Acceptance Criteria

1. [ ] `ComposeBrief` が panic stub でない。
2. [ ] 数値 placeholder がすべて定数で置換され、`{{SOURCES}}` / `{{JSON_EXAMPLE}}` が埋まる。
3. [ ] 0 件 slice は呼び出し側責務（本 Issue では precondition 違反を panic または contract test で明示）。
4. [ ] `go test ./internal/application/build/...` が pass する。

## 8. Verification

```bash
cd apps/generator
go test ./internal/application/build/... -count=1
go build ./...
```

## 9. Dependencies

1. A/B artifact（`TextWriterBriefPrompt`、`WriterOutput` wire、`brief.go` stub）が branch 上にあること。
2. `ProduceEpisode.Run` 本体は不要（D）。

## 10. Risks

1. `JSON_EXAMPLE` を手書きすると wire drift する。`WriterOutput` から生成する。

## 11. Notes

1. GitHub Issue 化は別判断。本 file が達成契約の正。
