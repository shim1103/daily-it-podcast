## 1. Summary

この Issue では、`apps/generator/test/secrettransport_agentsecrets_narrow_integration_local_test.go` を `local_real` build tag 付きで追加し、実 AgentSecrets proxy 経由で SecretRef → 秘密名注入（`X-AS-*`）が有効に機能し、ターゲット upstream へ到達することを self-validate する。

## 2. Context

1. `local_real` build tag により、local 実物 suite は CI の Integration gate 収集から除外する契約が `apps/generator/test/local_real_build_tag.go` と `scripts/generator/test-integration-local.sh` で固定済みである。
2. Fake 版の Narrow Integration(既存 1 file) は proxy 実装をテスト内で差し替え、注入ヘッダ（`X-AS-Inject-Bearer` 等）と “未解決 secret は proxy 前に失敗” を検証している。
3. 本 Issue の目的は、Fake ではなく実 AgentSecrets proxy 経路（`secrettransport/agentsecrets`）を、local の OS keychain test 値前提で self-validate することである。

## 3. Canonical Sources

1. `docs/decisions/2026-08-26T17-43-00-docs-infra-test-discussion.md` — local_real 収集除外は build tag で行う。
2. `docs/decisions/2026-08-26T17-44-00-docs-infra-test-discussion.md` — local 実物は OS keychain の test 専用値を使い、本番値は test に流用しない。
3. `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md` — Integration gate は secret なし Narrow のみ。
4. `testing-strategy/levels.md` — Narrow Integration の定義（外部境界1つの実 I/O 契約検証）。
5. `apps/generator/internal/infrastructure/secrettransport/agentsecrets/client.go` — AgentSecrets proxy 経由で送る HTTP リクエスト contract（`X-AS-Target-URL` / `X-AS-Method` / `X-AS-Inject-*`、error message に秘密値・request body・proxy response body を含めない）。
6. `apps/generator/test/secrettransport_agentsecrets_narrow_integration_test.go` — Fake 版が検証している observable の境界。
7. `.agent/workflows/agentsecrets.md` — local 実 proxy の運用手順（`agentsecrets proxy start` 等）。

## 4. Scope

### In Scope

1. `apps/generator/test/secrettransport_agentsecrets_narrow_integration_local_test.go` の追加（`//go:build local_real`）。
2. 実 AgentSecrets proxy 経路で、resolved secret の request が upstream（自前の probe server）へ到達することを検証する。
3. unresolved secret は proxy に到達する前に error となり、upstream が呼ばれないことを検証する。
4. upstream 側の検証観測は secret 値を含めず “存在/形式” だけを見る（失敗メッセージに secret 値を載せない）。

### Out of Scope

1. HTTP 出口の vendor 実 API（GetX / Twitter / Gemini / Drive / OAuth）まで実到達させること。
2. System / E2E / Broad Integration（複数境界の合成を CI で回すこと）。
3. `commandlaunch/agentsecrets` の実 wrapper 経路（別 Issue）。

## 5. Contract

1. local_real suite は `-tags local_real` のときのみ実行される（CI の default `go test ./test/...` へ混入しない）。
2. `secrettransport/agentsecrets.Client` の成功時、ターゲット upstream（テスト内の probe server）へ到達する。
3. Bearer 注入の成功時、upstream で観測される `Authorization` 等の認証ヘッダは非空で、想定フォーマット（例: `Bearer ` prefix）に一致する（値そのものは出力/ログ化しない）。
4. unresolved secret の request は proxy へ到達する前に失敗し、upstream へのアクセスは観測されない。
5. error message は秘密値 / request body / proxy response body を含まない。

## 6. Constraints

1. 実 proxy を使うため、local 実行前に proxy が到達可能であること（デフォルト port と endpoint）が必要。
2. upstream 側で受信したヘッダの生値を失敗メッセージへ含めない。
3. secret 値をテスト側で取り出すための追加 agentsecrets CLI（value を表示するコマンド等）を使わない。

## 7. Acceptance Criteria

- [ ] `apps/generator/test/secrettransport_agentsecrets_narrow_integration_local_test.go` は `local_real` build tag を持ち、`go test -tags local_real ./test/...` でコンパイル/実行できる。
- [ ] upstream probe server が呼ばれ、認証ヘッダが “非空かつ想定prefixを持つ” ことを検証する（secret 値を出さない）。
- [ ] unresolved secret のケースで `Do()` がエラーになり、upstream が呼ばれない。
- [ ] unresolved/resolved いずれの失敗でも error message に秘密値・request body が含まれない。
- [ ] `bash scripts/generator/test-integration-local.sh` 実行時に本 Issue の local_real suite が pass する（前提: local 環境が正しくセットアップ済み）。

## 8. Verification

1. 事前に local 環境で AgentSecrets が初期化され、実 proxy が起動していることを確認する（例: `agentsecrets status` と `agentsecrets proxy start`）。
2. リポジトリ root から以下を実行する。

```bash
bash scripts/generator/test-integration-local.sh
```

3. 期待: 本 Issue の local_real suite が pass する。

## 9. Dependencies

- `scripts/generator/test-integration-local.sh`（local_real suite の入口）。
- `apps/generator/test/local_real_build_tag.go`（収集境界の code 契約）。
- local 環境の AgentSecrets proxy（デフォルト proxy endpoint）と Cursor 用 keychain/test values。

## 10. Risks

1. local 環境で proxy が起動していない/到達できない risk。
   1. 代替: テスト失敗メッセージを “proxy not running” に限定し、secret 値を出さない。
2. proxy 設定により認証ヘッダの注入先が期待と違う risk。
   1. 代替: フォーマット検証は “prefix/非空” に留め、値一致は要求しない。
3. upstream のログ出力に secret 値が混入する risk。
   1. 代替: failure 時にヘッダ生値を出さない。

## 11. Notes

この Issue は local 専用（実 AgentSecrets / 実 OS keychain / 実 proxy 前提）であり、CI の gate とは切り分ける。

