## 1. Summary

このIssueでは、未使用のTwitterAPI.io providerを現行artifactから除き、Generatorのproduction sourceをGetXAPIだけにする。完了後はcode、test、Composition、runtime名、latest docs、保留taskのいずれもTwitterAPI.ioをsupported sourceとして示さない。

## 2. Context

1. GetXAPIを唯一のproduction sourceとし、TwitterAPI.ioの現行artifactを削除するDecisionは確定している。
2. 現在はTwitterAPI.io Adapter、test、Composition constructor、secret bindingとenvironment名が残っている。
3. Runtime config inventoryからTwitterAPI.ioは除外済みだが、READMEは旧実装が残る移行中状態を明記している。

## 3. Canonical Sources

1. `docs/decisions/2026-08-27T13-56-15-docs-env-secret-management-reconsider.md` — TwitterAPI.ioの現行artifactを削除し、GetXAPIへ一本化するDecision。
2. `apps/generator/internal/config/` — Generatorが持つruntime config契約。Source capabilityの正。
3. `README.md` — latest runtime config inventoryとproduction source方針。
4. `apps/generator/internal/infrastructure/x/getxapi/` — 削除後も維持するproduction source Adapter。
5. `testing-strategy` — test scopeと削除時のregression確認に使う共通規則。

## 4. Scope

### In Scope

1. TwitterAPI.io Adapter、test、Composition constructorを削除する。
2. TwitterAPI.io用environment名、secret binding、current code commentを削除する。
3. TwitterAPI.io Narrow testの保留taskを削除する。
4. latest docsとlaneから移行中のTwitterAPI.io記述を除く。
5. GetXAPI production sourceとそのtestが維持されることを確認する。

### Out of Scope

1. GetXAPI Adapterの再設計。
2. `secrettransport`廃止とruntime Configへの接続。
3. vendor実APIへの通信。
4. 過去のDecision Recordとdaily recordの変更または削除。

## 5. Contract

1. Generatorが提供するproduction `ItemSource`はGetXAPIだけである。
2. TwitterAPI.io用credentialはcurrent runtime configとして要求されない。
3. 過去時点の検討・実装事実を持つ履歴artifactは保持する。

## 6. Constraints

1. TwitterAPI.ioとの比較を理由にGetXAPIのpublic contractを変更しない。
2. 過去のDecision Recordとdaily recordをlatest inventoryとして扱わない。
3. 削除対象に隣接する共有bindingを変更する時も、他capabilityのbindingを壊さない。

## 7. Acceptance Criteria

1. [ ] TwitterAPI.io Adapter、test、Composition constructorがcurrent treeに存在しない。
2. [ ] `TWITTER_IO_API_KEY`とそのbindingがcurrent Generator codeに存在しない。
3. [ ] TwitterAPI.io Narrow testの保留taskが存在しない。
4. [ ] READMEとGenerator laneがTwitterAPI.ioをcurrent実装または将来taskとして示さない。
5. [ ] GetXAPI Adapter、Composition constructor、Unit testが維持され、Generator testがpassする。
6. [ ] 過去のDecision Recordとdaily recordは変更されていない。

## 8. Verification

```bash
./scripts/generator/check-static.sh
./scripts/generator/test-unit.sh
./scripts/generator/test-race.sh
./scripts/generator/test-integration.sh
rg 'TwitterAPI|twitterapiio|TwitterIO|twitterIO|TWITTER_IO_API_KEY' apps/generator README.md DESIGN.md
test ! -e docs/tasks/todo/generator-narrow-gate-vendor-twitterapiio.md
git diff --check
```

`rg`はmatchなしでexit 1になることを確認する。履歴確認では`docs/decisions/`と`docs/daily/`を削除対象にしない。

## 9. Dependencies

1. `generator-remove-agentsecrets-local-real.md`にblockedされる。Compositionとsecret binding周辺の競合を避け、削除差分を独立reviewできるようにするため。

## 10. Risks

1. 共有bindingの削除範囲を広げるとGetXAPIや他capabilityを壊すため、TwitterAPI.io固有symbolだけを除く。
2. current code commentやtaskが残るとsupported providerに見えるため、source code以外のlatest artifactもscanする。

## 11. Notes

