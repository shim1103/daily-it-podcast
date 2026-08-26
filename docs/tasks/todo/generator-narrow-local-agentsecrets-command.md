## 1. Summary

この Issue では、`apps/generator/test/commandlaunch_agentsecrets_narrow_integration_local_test.go` を `local_real` build tag 付きで追加し、実 AgentSecrets（`agentsecrets env --`）経由で Cursor CLI 用 child process が正しい cwd（Cursor 専用 project dir）と allowlist env のみを受け取ることを自動検証する。

## 2. Context

1. `local_real` build tag により、local 実物 suite は CI の Integration gate 収集から除外する契約が `apps/generator/test/local_real_build_tag.go` と `scripts/generator/test-integration-local.sh` で固定済みである。
2. Fake 版の Narrow Integration(既存 1 file) は child cwd / allowlist env / 観測面の漏れ（stdin・stderr 本文や parent-only env）がないことを検証している。
3. 本 Issue の目的は、Fake ではなく実 AgentSecrets wrapper 経路（`commandlaunch/agentsecrets`）を、local の OS keychain test 値前提で self-validate することである。

## 3. Canonical Sources

1. `docs/decisions/2026-08-26T17-43-00-docs-infra-test-discussion.md` — local_real 収集除外は build tag で行う。
2. `docs/decisions/2026-08-26T17-44-00-docs-infra-test-discussion.md` — local 実物は OS keychain の test 専用値を使い、本番値は test に流用しない。
3. `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md` — Integration gate は secret なし Narrow のみ。
4. `testing-strategy/levels.md` — Narrow Integration の定義（外部境界1つの実 I/O 契約検証）。
5. `apps/generator/internal/infrastructure/commandlaunch/agentsecrets/env_wrapper.go` — `agentsecrets env --` wrapper の contract（ProjectDir 絶対 path、SecretKeys を argv/child env に載せない等）。
6. `apps/generator/internal/infrastructure/commandlaunch/agentsecrets/launcher.go` — Launcher contract（program empty の事前失敗、stdin/stderr 本文を error に載せない、child env の allowlist 制限）。
7. `apps/generator/test/commandlaunch_agentsecrets_narrow_integration_test.go` — Fake 版が検証している observable の一覧。

## 4. Scope

### In Scope

1. `apps/generator/test/commandlaunch_agentsecrets_narrow_integration_local_test.go` の追加（`//go:build local_real`）。
2. 実 AgentSecrets wrapper 経路で、既存 Fake 版と同等の observable を検証する（同一の失敗/観測面境界を self-validate）。
3. child process 側の出力は secret 値を含めない形で redaction し、テスト失敗時でも値がログに出ないようにする。

### Out of Scope

1. CI / push の Integration gate への混入（本 Issue の収集契約は A の code 契約に従う）。
2. HTTP 出口（`secrettransport/agentsecrets`）の実 proxy 経路検証（別 Issue）。
3. vendor 実 API（GetX / Twitter / Gemini / Drive / OAuth 等）や System / E2E。

## 5. Contract

1. local_real suite は `-tags local_real` のときのみ実行される（CI の default `go test ./test/...` へ混入しない）。
2. child cwd（観測可能な `pwd` / `readlink` 相当）は Cursor 専用 project dir（`composition.CursorCommandProjectDir()` の解決結果）に一致する。
3. child env は `composition.CursorCommandInheritedEnvNameAllow()` の許可名と、実 wrapper が注入する Cursor API key の「存在」に限定される。parent-only env（テストで設定した識別キー名）は含まれない。
4. child 起動失敗（例: empty program 相当）は child を開始せず、親側は nil/stdout だけを返さない形で失敗を表す。
5. child が stdin / stderr に書いた識別トークンは、返る error message に含まれない。

## 6. Constraints

1. 実 AgentSecrets の secret 値そのものをテストコード内に取り出し / 表示しない（観測・検証は “set/present” 等の値非依存にする）。
2. `agentsecrets` CLI が未初期化または Cursor 専用 project のセットアップが無い環境では、テストは失敗または明示的に中断する（secret 値を chat に貼る指示は禁止）。
3. CI / hooks / GHA から実行されないことを前提とし、実行コマンドは local のみ（`scripts/generator/test-integration-local.sh`）に帰属させる。

## 7. Acceptance Criteria

- [ ] `apps/generator/test/commandlaunch_agentsecrets_narrow_integration_local_test.go` は `local_real` build tag を持ち、`go test -tags local_real ./test/...` でコンパイル/実行できる。
- [ ] 実 wrapper 経路で child cwd が Cursor 専用 project dir と一致する。
- [ ] 実 wrapper 経路で child env は allowlist env 名だけが存在し、parent-only env 名は存在しない（secret 値を含めずに観測できる）。
- [ ] child が非0 exit の場合に返る error message が、テストで用いた stdin / stderr 識別トークン文字列を含まない。
- [ ] 本 Issue が生成するテストは secret 値そのものを出力/ログ化しない (fail message に値を含めない)。

## 8. Verification

1. 事前に local 環境で AgentSecrets が初期化されていることを確認する（例: `agentsecrets status`）。
2. リポジトリ root から以下を実行する。

```bash
bash scripts/generator/test-integration-local.sh
```

3. 期待: `commandlaunch_agentsecrets_narrow_integration_local_test.go` を含む `local_real` suite が pass する。

## 9. Dependencies

- `scripts/generator/test-integration-local.sh`（local_real suite の入口）。
- `apps/generator/test/local_real_build_tag.go`（収集境界の code 契約）。
- local 環境の AgentSecrets セットアップ（Cursor 専用 project と必要な key の存在）。

## 10. Risks

1. local 環境の AgentSecrets セットアップ不足で失敗する risk がある。
   1. 代替: テスト中断時の失敗メッセージを “set up missing” に限定し、secret 値を出さない。
2. child 側の観測コマンドが secret 値をログに出す risk がある。
   1. 代替: `${VAR:+set}` のように “値非表示の set/present” だけを出す観測に寄せる。

## 11. Notes

この Issue は local 専用（実 AgentSecrets / 実 OS keychain 前提）であり、CI の gate とは切り分ける。

