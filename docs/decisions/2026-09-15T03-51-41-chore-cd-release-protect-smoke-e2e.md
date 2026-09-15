---
name: masterへのCD gate配線をPR/push triggerへ揃え、gate外weekly cronを廃止する
date: 2026-09-15T03:51:41
branch: chore/cd-release-protect-smoke-e2e
---

## 1. Decision

1. `master` branchに`branch protection`を新設する。`required_status_checks.contexts`は既存必須gate（`static-and-unit`・`integration`、いずれも`test-unit.yml`・`test-integration.yml`のjob名）のみを常時requiredにする。`enforce_admins: false`、`required_pull_request_reviews`は設定しない（review 0人）。
2. `generator-system.yml`は`schedule`を外し、`pull_request: {branches: [master]}` + `workflow_dispatch`にする。
3. `playback-e2e.yml`は`schedule`を外し、`push: {branches: [master]}`（deploy直後）+ `workflow_dispatch`にする。
4. 新設する`playback-smoke`（別Decision `2026-09-15T03-52-*` 参照）は`pull_request: {branches: [master]}` + `workflow_dispatch`にする。
5. `system`・`playback-smoke`をrequired statusへ追加するのは、それぞれのworkflow新設PRがmergeされた**次のPRから**とする。新設PR自身では追加しない（GitHub仕様上、まだmasterに存在しないcontextはそのPRでrequiredにできない）。

## 2. Reason

1. `master`には現在`branch protection`が存在しない（`gh api .../branches/master/protection` → 404）。push制限も必須check強制も無く、gateが機能する保証が無かった。
2. `generator-system.yml`・`playback-e2e.yml`は共に`on: schedule + workflow_dispatch`のみで、そもそもPRに一度も連動していなかった。「masterへのPRでe2e/system-testが走る」という前提自体が実装と食い違っていた（issue記載の前提誤り）。
3. weekly cronは「時間経過」に対する回帰検知であり、目的が「codeの変更」に対する回帰検知（本Decisionの主眼）とは異なる。変更が無い週にも同じtestを繰り返す価値は無く、release（masterへのPR + merge）に紐付けるほうがPrinciple of Least Astonishment（一貫性）に沿う。外部API仕様の経時変化検知は別軸の課題であり、本Decisionの scope外。
4. `playback-e2e`をpushトリガー（deploy後）にするのは、E2Eの検証対象（Access込みの本番URL疎通）がdeploy後にしか物理的に存在しないため。deploy前に検証できる部分は`playback-smoke`が担う（別Decision）。
5. yml新設PRでrequired statusが即時強制されない制約（Reason 1本目に付随する既知のGitHub仕様）を運用手順として明示しておかないと、「新設workflow追加PRだけなぜかgateが効かない」という再度の混乱を招く。2段階運用として明文化する。

## 3. Rejected

1. **既存workflowをPR triggerに変えず、branch protectionだけ新設する案** — 新設したcontextsが実際にはPRで発火しないため、requiredに指定してもPRがstuckするか、そもそも意味を持たない。protection単体では解決しない。
2. **`system`・`playback-e2e`のscheduleを残したままPR triggerを追加する案（両方持たせる）** — weekly cronの目的（時間経過での検知）と release gateの目的（変更検知）が同じworkflowに同居すると、どちらの目的で落ちたか読み手が判別しづらくなる（Fault Isolation低下）。cron自体の価値も薄い（Reason 3本目）ため、削除する。
3. **新設workflowのcontextを新設PR自身でrequiredにする案** — GitHub側の仕様上、そのPRの時点でcontextがmasterに存在しないため強制されない。「効いていないrequired設定」を残すと後続の判断者を誤らせる。2段階運用に切り出す。
