## 1. Summary

このIssueでは、GeneratorのCompositionをruntime config loaderへ接続し、HTTP Adapter（getxapi / oauth / gemini / gdrive）を標準の`*http.Client`とcapability Configだけへ依存する形へ移行する。完了後は`secrettransport`が現行artifactから消え、既存Unit / Integration gateが移行後配線でpassする。

## 2. Context

1. configuration boundaryのA契約と、HTTP Adapterを`*http.Client`+Configへ依存させるDecisionは確定している。
2. AgentSecrets / `local_real`除去とTwitterAPI.io除去は完了している。
3. 現行HTTP Adapterと関連testはまだ`secrettransport` / processenv前提である。
4. vendorごとのSociable Unit / Narrow責務分離の仕上げは後続taskが所有する（Decision `2026-08-28T12-49-02`）。

## 3. Canonical Sources

1. `docs/decisions/2026-08-27T13-56-13-docs-env-secret-management-reconsider.md` — configuration boundaryの所有。
2. `docs/decisions/2026-08-27T13-56-14-docs-env-secret-management-reconsider.md` — HTTP Adapter依存形。
3. `docs/decisions/2026-08-28T12-49-00-docs-infra-unit-narrow-integration-latest.md` — M1後にSU/NIを揃える順序。
4. `docs/decisions/2026-08-28T12-49-02-docs-infra-unit-narrow-integration-latest.md` — M1のgreen境界と後続分割。
5. `apps/generator/internal/config/` — runtime config契約とloader実装（C-04完了後）。
6. `DESIGN.md` — Generator target architectureとIntegration gate方針。
7. `testing-strategy` — Scope、credential、二重最小化。

## 4. Scope

### In Scope

1. Compositionをconfig loaderの検証済みConfigへ接続する。
2. getxapi / oauth / gemini / gdrive のHTTP Adapterを`*http.Client`と必要なcapability Config / credentialだけへ依存させる。
3. `secrettransport` package、SecretRef binding、関連processenv HTTP実装と専用testを現行artifactから除く。
4. 既存Unit / Integrationが移行後配線でpassする最小のtest更新を行う。

### Out of Scope

1. getxapi / oauth / gemini / gdrive のSU/NI責務分離の仕上げ（後続`generator-su-ni-*`）。
2. Cursor CLI Adapter、`commandlaunch`、child environment再設計。
3. ProduceEpisode本体、Broad / System・E2E、vendor実API。
4. GitHub Actions production workflowの新設。
5. 過去のDecision Recordとdaily recordの変更または削除。

## 5. Contract

1. HTTP Adapterはenvironment keyやcredential保存元を知らない。
2. Compositionは検証済みcapability Configを必要な依存へ渡す。
3. 現行treeに`secrettransport`実装とそれを要するproduction結線が存在しない。
4. default Generator Unit / Integration gateがlocal secretなしでpassする。

## 6. Constraints

1. Cursor / commandlaunch経路をこのIssueで変更しない。
2. local secret storeや`.env` loaderを追加しない。
3. credential値をtest、log、Error、docsへ記録しない。
4. 過去のDecision / dailyをlatest inventoryとして書き換えない。

## 7. Acceptance Criteria

1. [ ] getxapi / oauth / gemini / gdrive のproduction constructorが`*http.Client`とcapability Config / credentialだけを受け取る。
2. [ ] Compositionがconfig loader結果からHTTP経路を結線し、`secrettransport`を参照しない。
3. [ ] `secrettransport` packageとその専用Unit / Narrowがcurrent treeに存在しない。
4. [ ] `./scripts/generator/check-static.sh`、`test-unit.sh`、`test-race.sh`、`test-integration.sh`がpassする。
5. [ ] Cursor / `commandlaunch`のproduction契約がこのIssueの差分で変わっていない。
6. [ ] 過去のDecision Recordとdaily recordは変更されていない。

## 8. Verification

```bash
./scripts/generator/check-static.sh
./scripts/generator/test-unit.sh
./scripts/generator/test-race.sh
./scripts/generator/test-integration.sh
rg 'secrettransport|SecretRef|BindingResolver' apps/generator
git diff --check
```

`rg`はmatchなしでexit 1になることを確認する。履歴確認では`docs/decisions/`と`docs/daily/`を削除対象にしない。

## 9. Dependencies

1. C-04（runtime config loader、`apps/generator/internal/config`のKISS化）は完了済み（進捗は`docs/tasks/todo/generator-lane.md`のC-04行）。本Issueはそのloader結果をproduction graphへ接続する責務を持つ。
2. 本IssueはCursor CLI GHA probe（C-03）と並行できる。
3. 本Issue完了後に`generator-su-ni-getxapi.md` / `oauth` / `gemini` / `gdrive`が着手可能になる。

## 10. Risks

1. 最小test更新を省略するとgateが赤のまま次taskへ漏れるため、M1完了条件にgate passを含める。
2. SU/NI仕上げまでM1に取り込むと失敗原因が混ざるため、責務分離ACは後続taskへ残す。

## 11. Notes

1. GitHub Issue化は別判断。本fileが達成契約の正。
2. Cursor NarrowはDecision `2026-08-28T12-49-01`どおり後回し。
