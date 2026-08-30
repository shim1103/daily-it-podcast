---
name: Cursor CLI TextWriter の Narrow Integration 追加と coverage 計測分母の見直し
date: 2026-08-30T12:05:00
session_id: none
branch: feature/generator-cursorcli-narrow-integration
prev: なし
---

## 1. Summary

`cursorcli.TextWriter` → `processenv.Launcher` → PATH 上の fake `agent` の端から端を、secret なし Narrow Integration として Integration gate へ載せた。成功時 fragment 返却・失敗時 Infrastructure Error と断片空・dummy secret 非漏洩・brief が exec stdin を上って子へ届くことを実プロセス境界で self-validate する。あわせて `processenv` の Sociable Unit から実 `cat` / `env` を起動して環境を観測する test 4 本を Narrow へ寄せて除去し、SU は nil ガードの Adapter 内分岐だけを持つ形へ戻した。`processenv` Narrow が借りていた `cursorcli.CursorAPIKeyEnvName` の逆依存を dummy 環境変数名へ置き換えて import を切った。`SandboxValue` は PR80 の GHA capability probe 実測に合わせ `enabled` → `disabled` とした。

SU から Narrow へ観測を移した結果、実プロセス起動が必須の production code が Unit coverage から外れて gate が 87.8% < 90% で落ちた。coverage 計測分母に secret なし Narrow を算入する方針へ変更し、`test-unit.sh` を `-coverpkg` で production package へ instrument を固定・実行対象へ `./test/...` を追加した。判断と先行 Decision の部分 supersede 範囲は Decision Record に残した。達成契約 file を削除し lane を完了へ更新した。

## 2. Changes

- 検証: `bash scripts/generator/test-unit.sh` 87.8% → 91.9%（`processenv` package は 32% → 97%、`Launch` 100% / `buildChildEnv` 87.5%）で exit 0。`go build ./...` OK、`bash scripts/generator/test-integration.sh` OK、`bash scripts/test-gate-composer-sociable-unit.shell` OK。
- `test-unit.sh` の `go test` 行を変えたため、その行を固定文字列で検査する `scripts/test-gate-composer-sociable-unit.shell` の needle も追随更新（この行は複数行分割不可）。
- `-coverpkg` を付けると各 test binary の per-package 表示 % が「全 production に対する自 test の寄与率」になり低く見える。判定は統合 profile の total で行う。
- 新規 Decision `docs/decisions/2026-08-30T11-52-00-feature-generator-cursorcli-narrow-integration.md` は先行 Decision `2026-08-17T14-45-00-chore-test-and-ci-coverage-layer.md` の「除外は Composition Root と cmd のみ／Integration に coverage を載せない」を部分 supersede（先行 file 本文は無変更）。
- branch を property name へ改名（`feature/generator-narrow-gate-vendor-cursorcli` → `feature/generator-cursorcli-narrow-integration`）。worktree dir 名は旧名のまま。sandbox が `.git/config` の write を拒むため config 更新は失敗したが branch ref の改名は成立。
- この session は途中で non-edit Ask / non-edit Question / issue-manager flow / `/pr-completion` と指示形態が切り替わった。詳細な進め方の誤りは lessons へ分割。
- GitHub Issue は無し（local task file が契約の正、削除済み）。
- PR 作成前に `origin/develop`（PR #94 マージ済み）を merge（ort strategy、conflict なし）。merge 後も generator gate 全緑を再確認。
- PR: https://github.com/shim1103/daily-it-podcast/pull/95 （base `develop`）

### Commits

- `2c71287`
- `303ca43`
- `a63255c`
