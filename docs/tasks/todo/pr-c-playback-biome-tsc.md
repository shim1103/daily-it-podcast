## 1. Summary
`apps/playback` に **formatter + linter + typecheck** を導入し、root の静的入口から検知できる状態にする。

完了後は `./scripts/check-static.sh` と `./scripts/test-unit.sh` で static 追加が観測できる。

## 2. Context
playback は Vitest の Unit/Integration を持つが、型や書式/未使用の静的検査は入口が未整備。
この PR では層検知や coverage とは切り分け、static entry を先に整備する。

## 3. Canonical Sources
`apps/playback/package.json`
`apps/playback/vitest.config.mjs`
`scripts/check-static.sh`
`scripts/test-unit.sh`
`DESIGN.md`（playback の技術選定方針）

## 4. Scope
### In Scope
1. `scripts/playback/check-static.sh` を新規追加し、Biome と `tsc --noEmit` を実行する
2. root の `scripts/check-static.sh` が generator + playback を呼ぶ
3. 設定ファイル（Biome / tsconfig）は `apps/playback/` 配下で完結する

### Out of Scope
1. depcruise 等での層検知（別 PR）
2. coverage gate（別 PR）
3. playback UI の機能追加

## 5. Contract
1. 追加する formatter/linter は playback の技術選定（vanilla + Pico.css）に合わせ、React/Next 専用規則を入れない
2. typecheck は `tsc --noEmit` を `apps/playback` の `tsconfig` で実行する
3. root の静的入口は `scripts/check-static.sh` を正とし、toolchain 実装を YAML/hookへ複製しない

## 6. Constraints
1. playback の static 検査導入のみ（層検知・coverage の責務を混ぜない）
2. 失敗時の復元は整形/unused/既存型整合のみで行う

## 7. Acceptance Criteria
- [ ] `./scripts/check-static.sh` を実行して exit 0 で完了する
- [ ] `./scripts/test-unit.sh` を実行して exit 0 で完了する（static 追加で Unit が落ちない）

## 8. Verification
```sh
./scripts/check-static.sh
./scripts/test-unit.sh
```

## 9. Dependencies
root の static entry（`scripts/check-static.sh`）が存在し、generator と playback の両方を呼べること。

## 10. Risks
Biome / tsc の導入により既存コードの “未整備” が顕在化し、PR が大きくなる risk。
Mitigation: React/Next 規則を入れず、復元範囲は整形・unused・既存型整合に限定する。

## 11. Notes
層検知（depcruise）と coverage gate は別 PR とし、本 PR では変更しない。

