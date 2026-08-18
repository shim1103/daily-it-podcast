## 1. Summary
`apps/generator` に **lint / formatter（Go の整形）/ 層違反検知**を静的入口へ導入する。

完了後は `./scripts/generator/check-static.sh` を実行すると exit 0/非0 で静的検査結果を観測できる。

## 2. Context
generator は既に Unit gate で `depguard` と statement coverage を扱っている。
この PR では Unit gate とは別軸の static entry を用意し、hook / GHA の入口として使える状態にする。

## 3. Canonical Sources
`docs/decisions/2026-08-17T14-45-00-chore-test-and-ci-coverage-layer.md`
`apps/generator/.golangci.yml`
`scripts/generator/check-static.sh`
`scripts/generator/test-unit.sh`

## 4. Scope
### In Scope
1. `scripts/generator/check-static.sh` が **lint / formatter**を実行し、exit code で結果を返す
2. `apps/generator/.golangci.yml` が **curated subset** のみを enable する（既存 `depguard` を維持）
3. `internal/entities` / `internal/application` / `internal/infrastructure` の import 制約（`depguard`）を維持する

### Out of Scope
1. playback（Biome / depcruise / coverage）の導入
2. worker/web の層検知（別 PR）
3. coverage gate の閾値や除外方針の変更（Unit gate を正とする）

## 5. Contract
1. `./scripts/generator/check-static.sh` は `golangci-lint` へ委譲し、root の YAML/hook に toolchain 実装を複製しない
2. formatter/整形は **実行可能な check** として扱い、自動修正 `--fix` はこの PR の責務に含めない
3. depguard の allow / deny は `apps/generator/.golangci.yml` の既存規則を維持する

## 6. Constraints
1. playback 側の静的検査（Biome / tsc）は変更しない
2. 新規 linter を大量 enable しない（curated subset のみ）

## 7. Acceptance Criteria
- [ ] `./scripts/generator/check-static.sh` を実行して exit 0 で完了する
- [ ] `./scripts/generator/test-unit.sh` が exit 0 で完了する（static 追加で Unit gate が壊れていない）

## 8. Verification
```sh
./scripts/generator/check-static.sh
./scripts/generator/test-unit.sh
```

失敗時は `golangci-lint` の指摘に従って該当 Go file の整形 / unused / errcheck などの **ローカル修正のみ**で復元する。

## 9. Dependencies
`./scripts/check-static.sh` が generator の static entry を呼べること（static entry の入口として利用される）

## 10. Risks
静的検査追加により “未整備の死角” が顕在化し、レビューが増える risk。
Mitigation: curated subset のみに限定し、失敗はローカル修正で閉じる。

## 11. Notes
coverage の閾値 / 除外方針は Unit gate の既存 decision を正とし、この PR では変更しない。

