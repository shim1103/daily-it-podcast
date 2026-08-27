## 1. Summary

このIssueでは、Cursor CLIをGitHub Actions runnerで非対話実行し、production command launch契約の未実測部分を観測する。完了後はCLIのinstall結果、実行binary、現在のargvの可否、最小child environment候補を、推測ではなくGitHub Actions上の結果から判断できる。

## 2. Context

1. Generatorのcredential付き実operationはGitHub Actions runnerだけで実行する。
2. 現在のCursor CLI Adapterはbinary名とargvを固定しているが、GitHub Actions上でのinstall path、model、sandbox、environment要件は未実測である。
3. 本Issueはproduction実装ではなくcapability probeであり、結果を後続A/Bへ固定することは別scopeである。

## 3. Canonical Sources

1. `docs/decisions/2026-08-27T12-17-00-docs-env-secret-management-reconsider.md` — credential付き実operationの実行環境をGitHub Actionsへ限定するDecision。
2. `apps/generator/internal/infrastructure/manuscript/cursorcli/constants.go` — probe対象となる現在のbinary名、model、mode、output、sandbox、trust引数。
3. `apps/generator/internal/infrastructure/manuscript/cursorcli/text_writer.go` — Cursor CLI Adapterが期待するstdin／stdout／Error境界。
4. `.github/workflows/test-unit.yml` — repositoryのGitHub Actions runnerとtoolchain setupの現行例。
5. `testing-strategy/credential.md` — test用credentialと観測データの取扱規則。

## 4. Scope

### In Scope

1. 明示実行だけで動く一時的なGitHub Actions probeを用意する。
2. 公式install手順の完了と、実際に解決されたCLI binaryの絶対pathを観測する。
3. 現在のCursor CLI argvについて、model利用可否、sandbox動作、repo root cwdでの非対話実行を観測する。
4. `HOME`、`PATH`、`TMPDIR`を未指定／限定指定したcaseを分離して実行し、必要なchild environment候補を観測する。
5. stderrは上限付きでprocess内部だけに保持し、workflow logやartifactへ本文を出さずにexit結果を観測する。
6. 各caseの入力条件、exit結果、binary path、最小child environment候補を後続判断へ渡せる形で残す。
7. probe完了後、一時workflowとprobe専用codeをcurrent treeから除く。

### Out of Scope

1. production Generator workflowの追加。
2. Cursor launcher、Composition、runtime Config契約の変更。
3. 実測結果のDecision化またはA契約への固定。
4. Cursor以外のvendor API検証。
5. 定期実行とCI必須gateへの追加。

## 5. Contract

1. probeは`workflow_dispatch`相当の明示操作でだけ実行し、通常のpush／pull request gateに入らない。
2. 各environment caseは独立して観測でき、一つの失敗で残りの結果を失わない。
3. `CURSOR_API_KEY`の値、prompt本文、stderr本文をworkflow log、artifact、Error表示へ出さない。
4. command成功を前提にせず、成功と失敗のどちらも観測結果として区別する。
5. 一時probe artifactは実測完了後のproduction treeへ残さない。

## 6. Constraints

1. 実行にはtest用`CURSOR_API_KEY`と、temporary workflowをremote branchで実行する明示許可が必要である。
2. probeのfailureをproduction contractへ自動昇格しない。service、entitlement、model、installer、environmentのどこで失敗したかを区別する。
3. API利用量を制限するため、各caseのpromptと実行回数を最小にする。
4. secret maskに依存してraw値を出力せず、最初から値をlog payloadへ含めない。

## 7. Acceptance Criteria

1. [ ] GitHub Actions runnerで公式install手順のexit結果と、解決されたCLI binaryの絶対pathを観測できる。
2. [ ] 現在のCursor CLI argvについて、model利用可否、sandbox動作、repo root cwdでの結果をcase別に観測できる。
3. [ ] `HOME`、`PATH`、`TMPDIR`の未指定／限定指定caseが互いに独立して実行され、必要性を比較できる。
4. [ ] 各caseについて実行条件、exit status、成功／失敗分類が残り、secret値、prompt本文、stderr本文は外部出力されない。
5. [ ] 観測結果から最小child environment候補と未確定事項を識別できる。
6. [ ] 一時workflowとprobe専用codeが完了差分に残っていない。

## 8. Verification

1. GitHub Actionsの各probe caseが実行され、run URLとcase別resultを確認できる。
2. workflow logとartifactを検索し、secret値、prompt本文、stderr本文が含まれない。
3. 最終差分で一時workflowとprobe専用codeが削除されている。
4. 通常のGitHub Actions workflowが変更されていない。

## 9. Dependencies

1. test用`CURSOR_API_KEY`がGitHub Actions Secretsに登録済みであること。
2. temporary workflowのcommit、push、明示実行が許可されること。
3. 本Issueは他のC Issueと並行実施できる。

## 10. Risks

1. Cursor側の一時障害やentitlement不足をenvironment要件と誤認するriskがあるため、失敗段階とexit分類を分けて記録する。
2. matrix case数に比例してAPI利用量が増えるため、同じ条件の再実行を避ける。
3. diagnostic出力へ機密情報が混入するriskがあるため、stderr本文をworkflow外へ出さず、結果metadataだけを記録する。

## 11. Notes

1. 本Issueの実測後、最小child environmentとCLI capabilityをA/B/Cへ再分類する。
