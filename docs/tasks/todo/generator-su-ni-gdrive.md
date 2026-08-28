## 1. Summary

このIssueでは、Google Drive EpisodeWriterのSociable Unitと既存Narrow Integrationをtarget architecture（`*http.Client` + capability Config）へ揃え、境界I/O契約を最新配線でself-validateする。完了後はUnitが`secrettransport` Stubに依存せず、secretなしNarrowがlist→create→upload境界だけを検証する。

## 2. Context

1. `gdrive_narrow_integration_test.go`は既にあるが、processenv / `secrettransport`前提のままである。
2. Sociable Unitは境界I/OなしのClient Stubだが、Stubが`secrettransport.Client`形である。
3. 旧narrow-gate系にgdrive専用taskは無かった。本fileがSU/NI latest化の正とする。

## 3. Canonical Sources

1. `docs/decisions/2026-08-28T12-49-00-docs-infra-unit-narrow-integration-latest.md` — M1後にSU/NIを揃える順序。
2. `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md` — Integration gateはsecretなしNarrow。
3. `docs/decisions/2026-08-27T13-56-14-docs-env-secret-management-reconsider.md` — HTTP Adapter依存形。
4. `apps/generator/internal/infrastructure/drive/gdrive/` — AdapterとUnit test。
5. `apps/generator/test/gdrive_narrow_integration_test.go` — 更新対象Narrow。
6. `testing-strategy` — levels / credential / 二重最小化。

## 4. Scope

### In Scope

1. 既存Narrowを`*http.Client` + capability Config形へ書き換える。
2. Narrow観測: list→create→uploadの到達、Authorization Bearer、json/wav stem一致、dummy Folder IDの非露出。
3. Sociable UnitのStubを移行後の依存形（標準Client差し替えまたは同等の境界なしdouble）へ合わせる。
4. Unitに残すのはAdapter内分岐。境界I/O成功経路の網羅はNarrowだけ。

### Out of Scope

1. Composition再結線と`secrettransport`削除（M1）。
2. OAuth TokenSourceのSU/NI（`generator-su-ni-oauth.md`）。
3. Google実API、Broad / System・E2E。

## 5. Contract

1. Integration gateが更新後Narrowを含みexit 0。
2. 成功時にlist→create→uploadが観測され、Bearerが付く。
3. failure messageにdummy Folder IDを出さない。
4. Unit = Adapter内分岐のみ。Narrow = 境界I/Oのみ。

## 6. Constraints

1. dummy値をfailure messageに出さない。
2. `secrettransport` / processenv Clientを再導入しない。
3. GitHub Issue化しない。本fileが契約の正。

## 7. Acceptance Criteria

1. [ ] Narrowが`secrettransport`をimportせず、Config / `*http.Client`形でgate実行できる。
2. [ ] list→create→upload到達とAuthorization Bearerを観測できる。
3. [ ] json/wavのstemが一致する成功経路がある。
4. [ ] Sociable Unitが`secrettransport.Client` Stubへ依存していない。
5. [ ] UnitはAdapter内分岐のみ、Narrowは境界I/Oのみ。
6. [ ] error messageにdummy Folder IDがない。

## 8. Verification

```bash
bash scripts/generator/test-integration.sh
go test ./internal/infrastructure/drive/gdrive/ -count=1
```

## 9. Dependencies

1. `docs/tasks/todo/generator-composition-http-adapters.md`（M1）完了後。
2. OAuth TokenSourceのNarrowは`generator-su-ni-oauth.md`と並行できる。

## 10. Risks

1. TokenSource StubとHTTP境界を混同するとoauth taskと二重になるため、gdrive NarrowはDrive API到達に限定する。

## 11. Notes

1. M1の最小配線更新でNarrowが暫定greenでも、本Issueの責務分離ACを満たすまで未完了とする。
