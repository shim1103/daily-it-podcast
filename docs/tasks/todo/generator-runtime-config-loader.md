## 1. Summary

このIssueでは、`apps/generator/internal/config`にfreeze済みのGenerator runtime config契約を実装し、process environment由来の入力を検証済みConfigへ変換できる状態にする。完了後はconfiguration boundaryのlogicとUnit testが独立してpassし、production Compositionへの接続を待てる。

## 2. Context

1. Config、environment名、Secret、Loader、validation ErrorのA契約は`apps/generator/internal/config`に存在する。
2. configuration boundaryをstartup時に一度だけ通すDecisionは確定している。
3. 現在のpackageは契約宣言だけで、Secretの具体実装、load／validation、Errorの具体実装を持たない。

## 3. Canonical Sources

1. `apps/generator/internal/config/names.go` — environment keyのA契約。
2. `apps/generator/internal/config/config.go` — capability別ConfigのA契約。
3. `apps/generator/internal/config/secret.go` — opaque SecretのA契約。
4. `apps/generator/internal/config/loader.go` — LookupEnvとLoaderのA契約。
5. `apps/generator/internal/config/error.go` — validation violationとErrorのA契約。
6. `docs/decisions/2026-08-27T13-56-13-docs-env-secret-management-reconsider.md` — configuration boundaryの所有、load時点、group化を定めるDecision。
7. `testing-strategy` — Unit test、credential、Error contract検証の共通規則。

## 4. Scope

### In Scope

1. freeze済み`Secret`契約を満たす具体実装を追加する。
2. `LookupEnv`からConfigを構築するloader logicを追加する。
3. A契約に定義済みのvalidation、違反集約、Error表示、defensive copyを実装する。
4. 正常系、違反分類、集約順、redaction、defensive copyをUnit testで検証する。
5. `Loader`型との適合をcompile時に検証する。

### Out of Scope

1. Composition Root、cmd entrypoint、`ProduceEpisodeFactory`への接続。
2. HTTP Adapter、Cursor launcher、Google OAuth、Drive writerのconstructor変更。
3. `secrettransport`、`commandlaunch/processenv`、secret bindingの削除。
4. `.env` loaderと複数config sourceの追加。
5. GitHub Actions Variables／Secretsの設定。

## 5. Contract

1. 公開境界、field、environment名、validation、Error、Secretの契約は`apps/generator/internal/config`のA artifactを変更せず実装する。
2. loaderは外部I/OやCompositionを知らず、注入された`LookupEnv`だけを入力源にする。
3. package単体でlogicと契約testを完了し、production graphへの未接続をErrorやfallbackで隠さない。

## 6. Constraints

1. Aでfreezeされていない新しいconfig source、capability、environment keyを追加しない。
2. raw runtime値をError、test failure message、fmt出力へ含めない。
3. validationのために入力値を自動trimして受理しない。
4. package外の既存production codeを変更しない。

## 7. Acceptance Criteria

1. [ ] 全入力が有効な時、loaderがA契約どおりのConfigを返す。
2. [ ] Aで定義された各違反分類と複数違反の集約順がtable-driven Unit testで検証される。
3. [ ] Secretの表示契約と`Reveal`契約がUnit testで検証される。
4. [ ] Error文字列、Violation、test failure messageにraw値が含まれない。
5. [ ] `Error.Violations()`の戻りをcallerが変更しても内部状態が変化しない。
6. [ ] loader実装が`Loader`型に適合し、`apps/generator/internal/config`単体testがpassする。
7. [ ] Composition、Adapter、既存transportには変更がない。

## 8. Verification

```bash
cd apps/generator
go test ./internal/config -count=1
go test ./internal/config -race -count=1
cd ../..
./scripts/generator/check-static.sh
./scripts/generator/test-unit.sh
git diff --check
```

## 9. Dependencies

1. `apps/generator/internal/config/`のA artifactに依存する。
2. `docs/decisions/2026-08-27T13-56-13-docs-env-secret-management-reconsider.md`に依存する。
3. 本IssueはAgentSecrets削除、TwitterAPI.io削除、Cursor CLI probeと並行実施できる。
4. production graphへの接続を行う後続Issueは本Issueにblockedされる。

## 10. Risks

1. Error testが期待値へraw secretを埋め込むriskがあるため、漏洩確認は値そのものをfailure出力へ展開しないassertionで行う。
2. package外まで同時変更するとconfiguration boundary実装とmigrationのfailureを分離できないため、変更範囲を`internal/config`へ限定する。

## 11. Notes

1. Production Composition接続と`secrettransport`廃止は後続Issueで扱う。
