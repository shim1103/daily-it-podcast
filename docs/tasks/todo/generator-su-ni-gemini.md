## 1. Summary

このIssueでは、Gemini SpeechSynthesizerのSociable Unitと既存Narrow Integrationをtarget architecture（`*http.Client` + capability Config）へ揃え、境界I/OとAdapter内分岐の二重検証を解消する。完了後はUnitが境界I/Oを持たず、secretなしNarrowがAPI key付き外向きHTTP契約だけをself-validateする。

## 2. Context

1. `gemini_narrow_integration_test.go`は既にあるが、processenv / `secrettransport`前提のままである。
2. `synthesizer_sociable_unit_test.go` / edge側にDialTLS・processenv実到達観測が残り、Unit/Narrowが混在している。
3. 旧narrow-gate系にgemini専用taskは無かった。本fileがSU/NI latest化の正とする。

## 3. Canonical Sources

1. `docs/decisions/2026-08-28T12-49-00-docs-infra-unit-narrow-integration-latest.md` — M1後にSU/NIを揃える順序。
2. `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md` — Integration gateはsecretなしNarrow。
3. `docs/decisions/2026-08-27T13-56-14-docs-env-secret-management-reconsider.md` — HTTP Adapter依存形。
4. `apps/generator/internal/infrastructure/speech/gemini/` — AdapterとUnit test。
5. `apps/generator/test/gemini_narrow_integration_test.go` — 更新対象Narrow。
6. `testing-strategy` — levels / credential / 二重最小化。

## 4. Scope

### In Scope

1. 既存Narrowを`*http.Client` + capability Config形へ書き換える。
2. Narrow観測: upstream POST到達、API key header、成功時非空音声、failure messageへのdummy非露出。
3. Sociable Unit / edgeから境界I/O実到達観測を除き、retry・空text・status分岐等を残す。
4. 同一observableの二重検証をやめる。

### Out of Scope

1. Composition再結線と`secrettransport`削除（M1）。
2. Gemini実API、Broad / System・E2E。
3. PCM/WAV変換などHTTP境界と無関係なUnitの再設計。

## 5. Contract

1. Integration gateが更新後Narrowを含みexit 0。
2. 成功時API key headerが届き、非空WAV相当を返す。
3. failure messageにdummy API keyを出さない。
4. Unit = Adapter内分岐のみ。Narrow = 境界I/Oのみ。

## 6. Constraints

1. dummy値をfailure messageに出さない。
2. `secrettransport` / processenv Clientを再導入しない。
3. GitHub Issue化しない。本fileが契約の正。

## 7. Acceptance Criteria

1. [ ] Narrowが`secrettransport`をimportせず、Config / `*http.Client`形でgate実行できる。
2. [ ] upstream到達とAPI key headerを観測できる。
3. [ ] 成功時非空音声を返す。
4. [ ] Sociable Unit / edgeにDialTLS + processenv実到達の成功経路観測が残っていない。
5. [ ] UnitはAdapter内分岐のみ、Narrowは境界I/Oのみ。
6. [ ] error messageにdummy API keyがない。

## 8. Verification

```bash
bash scripts/generator/test-integration.sh
go test ./internal/infrastructure/speech/gemini/ -count=1
```

## 9. Dependencies

1. `docs/tasks/todo/generator-composition-http-adapters.md`（M1）完了後。

## 10. Risks

1. retry分岐をNarrowへ移すと二重になるため、retryはUnitへ残しNarrowは最小成功/失敗に限る。

## 11. Notes

1. M1の最小配線更新でNarrowが暫定greenでも、本Issueの責務分離ACを満たすまで未完了とする。
