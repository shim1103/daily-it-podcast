## 1. Summary

この Issue では、AgentSecrets HTTP proxy を `secrettransport/agentsecrets` の単一 runtime として再設計し、旧 `infrastructure/agentsecrets` の HTTP Client（秘密名文字列 API）と wrap 層を無くす。完了後、HTTP 出口の正本は1つだけになる。本 Issue の主眼は **file の mv/rename ではない**。`philosophy`（DRY / Orthogonality / DIP / KISS）に沿った所有境界の再設計である。

## 2. Context

1. local AgentSecrets HTTP transport Issue は、AC 達成のため旧 `infrastructure/agentsecrets` proxy Client を wrap する `secrettransport/agentsecrets.Client` を追加した。
2. Issue 本文は wrap を要求していなかった。wrap は結線復元の到達手段であり、HTTP 出口の最終形ではない。
3. Decision `2026-08-25T08-03-00` は `agentsecrets.Client` が `secrettransport.Client` を満たす前提で書かれている。正本の置き場は本 Issue の Decision を正とする。
4. shim 指摘: 完全移管していない。mv/rename だけで閉じず、設計原則に照らした最終形へ進める。
5. 仮定: wrap 実装と Composition の local factory（`agentsecretsSecretTransportClient`）は既に存在する。

## 3. Canonical Sources

1. `apps/generator/internal/infrastructure/secrettransport/contract.go` — HTTP transport 契約。
2. `apps/generator/internal/infrastructure/secrettransport/agentsecrets/` — 吸収先 runtime（現状は wrap）。
3. `apps/generator/internal/infrastructure/agentsecrets/proxy.go` — 吸収元 PROXY Client（削除対象）。
4. `apps/generator/internal/infrastructure/agentsecrets/env_wrapper.go` — command 側。本 Issue で消さない。
5. `docs/decisions/2026-08-25T19-36-11-feature-generator-agentsecrets-http-transport.md` — HTTP proxy 正本の吸収判断。
6. `docs/decisions/2026-08-25T08-03-00-feature-generator-processenv-http-transport.md` — proxy と processenv の I/O 差・`agentsecrets.Client` 前提。
7. `docs/decisions/2026-08-25T13-53-55-feature-generator-processenv-command-launcher.md` — 置き場×出口、runtime 配置。
8. `docs/decisions/2026-08-25T14-20-18-feature-generator-processenv-command-launcher.md` — 出口軸 Issue 分割。
9. test 方針 — `testing-strategy`。
10. 設計原則 — `philosophy` / `design-philosophy`（本 Issue の主眼。mv/rename 完了判定に使わない）。

## 4. Scope

### In Scope

1. **設計（必須・主眼）**: `philosophy` に従い、HTTP × AgentSecrets の所有を `secrettransport/agentsecrets` 1箇所へ閉じる。`secrettransport/processenv` と同型の runtime 配置にし、正本が2つある状態（旧名前 API Client + SecretRef wrap）を解消する。DIP（Adapter は契約だけ）、DRY（PROXY 知識の単一所有）、Orthogonality（command 側 EnvWrapper と混ぜない）、KISS（恒久 wrap を残さない）を満たす形にする。
2. PROXY プロトコルを `secrettransport.Client`（SecretRef / BindingResolver）を直接満たす単一実装として再配置する。単なる path 移動・rename・機械的コピーで終わらせない。
3. `infrastructure/agentsecrets` の HTTP 側公開型（proxy `Client` / `Request` / `Inject` 等）と、それにだけ属する test を削除する。
4. Composition の local HTTP factory が吸収後の Client だけを結線すること。
5. production processenv HTTP path の維持。

### Out of Scope

1. `EnvWrapper` / Cursor 専用 project / `commandlaunch` AgentSecrets 実装（`generator-agentsecrets-cursor-command-launcher.md`）。
2. production 既定を AgentSecrets HTTP に切り替えること。
3. processenv HTTP 実装の変更。
4. `cmd/generator`、GHA、実 proxy / keychain 常駐の運用自動化。
5. file を動かしただけ・import path を付け替えただけで設計判断を省略すること（禁止）。

## 5. Contract

1. HTTP Adapter は引き続き `secrettransport.Client` のみに依存する。
2. AgentSecrets HTTP runtime は SecretRef を BindingResolver で秘密名へ解決し、proxy へは秘密名だけを載せる。秘密値を Go process の request へ入れない。
3. 未解決または無効な有効 SecretRef は外部 I/O 前に失敗する。
4. error / log に秘密値・request body・proxy response body を出さない。
5. 吸収後、Generator production / test code が旧 proxy 公開 API（秘密名文字列の `Request` / `Inject`）を import しない。

## 6. Constraints

1. Generator code は GitHub Actions 固有 API を import しない。
2. `EnvWrapper` と Cursor project 規約を削除・移動しない。
3. processenv HTTP 実装を削除・置換しない。
4. test は dummy / test double を使い、本番 keychain と実 proxy を必須にしない。
5. mv/rename や機械的コピーだけで完了扱いにしない。設計主眼（正本の単一化・wrap 解消・processenv と同型配置）を満たすこと。

## 7. Acceptance Criteria

1. [ ] HTTP × AgentSecrets の正本が `secrettransport/agentsecrets` のみであり、旧 proxy Client への wrap 依存が無い（設計: DRY / DIP）。
2. [ ] `infrastructure/agentsecrets` に HTTP proxy 公開 API と専用 test が残らない。
3. [ ] runtime 配置が `secrettransport/processenv` と同型として読める（設計: Orthogonality / 一貫性）。単なる rename 痕跡だけの配置になっていない。
4. [ ] Composition local factory が吸収後 Client を返し、HTTP Adapter が具象旧 Client へ戻らない。
5. [ ] production processenv HTTP path が退行しない。
6. [ ] Generator の static、Unit、Integration gate が pass する。

## 8. Verification

```bash
./scripts/generator/check-static.sh
./scripts/generator/test-unit.sh
./scripts/generator/test-integration.sh
```

## 9. Dependencies

1. blocked by: local AgentSecrets HTTP transport（結線復元。完了済み / todo file 削除済み）。
2. related: `generator-agentsecrets-cursor-command-launcher.md`（EnvWrapper 所有。互いに HTTP 吸収を block しないが、EnvWrapper 削除は禁止）。

## 10. Risks

1. wrap 削除時に PROXY header 組み立てを取りこぼす → 吸収先 sociable / Narrow で `X-AS-*` を固定する。
2. EnvWrapper を誤って削除する → Out of Scope と Dependencies で command Issue を明示する。
3. processenv path を巻き込む → production 結線と既存 processenv gate の退行を Verification で見る。

## 11. Notes

1. 決定本文は Canonical Sources の Decision を参照し、ここへ転記しない。
2. follow-up の command 出口は `generator-agentsecrets-cursor-command-launcher.md`。
