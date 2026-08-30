---
name: playback web/worker への React・Hono・Hono RPC 導入と CI toolchain version 一本化
date: 2026-08-26T14:50:59
session_id: none
branch: docs/architecture-reconsider-react-hono
prev: なし
---

## 1. Summary

playback web/worker のarchitecture再検討から、React/Hono/Hono RPC導入の decision確定、A区分（契約固定）の実装、CI/Node/Go toolchain versionのSSOT一本化、残作業のIssue 6本への分割までを完了した。scope-splitのA/B/C/Dに沿って、B（decision）とA（契約固定code）を今session内で完了し、C（振る舞い実装）はIssue化のみで実装は次sessionへ送った。

## 2. Changes

1. 過去decision（`2026-08-18T11-12-00-feature-playback-web.md`、React/Next.js/shadcn不使用）を、学習コスト・慣れを評価軸から除外しYAGNIと学習目的の裁量採用のみで判断する軸に立て直して再評価し、Hono導入による一貫性前提の変化を根拠にReact/Hono/Hono RPC採用へ更新した。
2. haiku executorへA区分（dependency追加・jsx設定・Hono instance・RPC client型配線）を委譲し、完了報告を検証した結果、pnpm誤用によるlockfile汚染、React 18系への古いversion固定、branch coverage gateの偽陽性、biome `useHookAtTopLevel`の偽陽性（worker側の`useCase`命名を誤検知）の4件を発見し、いずれも修正した。
3. Go/Node toolchain versionが`go.mod`とCI workflow YAMLの複数箇所に直書きされていたため、`go-version-file`/`node-version-file`参照へ一本化し、`.nvmrc`を新設、`engines`+`engine-strict`でlocalの version不一致を`npm ci`時点で検知するようにした。
4. `go.mod`へ`toolchain go1.26.6`を追加したが、`go` directiveと同一versionのため`go mod tidy`が自動的に除去することを実機確認し、decision本文をこの実際の仕様に合わせて訂正した。
5. CI gate test（`test-gate-composer-sociable-unit.shell`）が`go-version: "1.26.6"`という直書き文字列をgrepで検証しており、DRY化した実装と衝突したため、`go-version-file`/`node-version-file`参照の存在を検証する形へ書き換えた。
6. worker側のHono route定義（C1）・入口切替（C2）、web側のViewModel hook化（C3）・Primitive/Feature ComponentのJSX化（C4/C5）・PageのJSX化とmount切替（C6）を、依存順を明記した6本のIssue fileへ分割した。
7. commit 3本（feat: React/Hono/RPC契約固定、chore: CI toolchain一本化、docs: Issue分割）へ分割してpushした。pre-commit/pre-push hookの`mktemp`権限エラーはsandbox制約によるものと確認しsandbox解除で回避した。git stage解除（`restore --staged`/`reset HEAD`）がpolicyでblockされたため、stage済みの内容をそのまま活かす単位でcommit粒度を組み直した。

### Commits

1. `73449ed`
2. `1cbf994`
3. `8f0d063`
