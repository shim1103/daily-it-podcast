## 1. Summary

このIssueでは、secret なし Broad Integration で `ProduceEpisode` 合成経路の配線・状態伝播・error 伝播を self-validate する。完了後、Integration gate（pre-push / GHA）が当該 suite を収集し、Narrow では検出できない合成失敗を落とせる。

## 2. Context

1. Integration gate は secret なし Narrow と Broad を載せる（Decision `2026-08-30T11-56-00`）。
2. Narrow は vendor 1 境界の I/O を所有する。Broad は複数 production Adapter 結線越しの合成だけを追加で見る。
3. `ProduceEpisode.Run` は現時点 stub。本 Issue の Verification は Run 実装後に初めて満たせる。
4. System（credential 付き・GHA）は本 Issue の対象外（lane D）。

## 3. Canonical Sources

1. `docs/decisions/2026-08-30T11-56-00-docs-generator-broad-system-e2e-plan.md` — gate に Broad を載せる判断
2. `apps/generator/internal/application/produce_episode.go` — Run の `@ensure`（0 件 / 成功手順 / 途中失敗で書込なし）
3. `docs/decisions/2026-08-29T14-14-00-docs-produce-episode-run-spec-fetch-zero-source-items.md` — 0 件時 Domain Error
4. `docs/decisions/2026-08-29T14-10-00-docs-produce-episode-run-spec-tts-segment-split.md` — TTS 分割順
5. 既存 Narrow（`apps/generator/test/*_narrow_integration_test.go`）— httptest / DialTLS redirect・fake child の形
6. `DESIGN.md` §5 — 置き場・命名・gate
7. test 方針 — `testing-strategy`（levels / minimization / naming-and-layout / contracts）

## 4. Scope

### In Scope

1. `apps/generator/test/` に `*_broad_integration_test.go` を追加する（build tag なし。gate 収集）。
2. production Adapter 結線（Composition 経路）を通し、真外部は httptest TLS redirect / fake Cursor binary 等の double にする。
3. 合成で初めて見える postcondition だけを assert する（成功時の書込到達、0 件・途中失敗時の書込なし、代表的な call 回数）。
4. Narrow file 固有 helper への相乗りをやめ、共有が必要なら `apps/generator/test/` 内の中立 support へ出す。

### Out of Scope

1. System Test（`test/system`・`test-system.sh`・GHA credential 付き）。
2. Sociable Unit が所有する Adapter 内分岐・envelope / HTTP 枝の網羅。
3. Narrow が所有する 1 境界 I/O 契約の再 assert。
4. `ProduceEpisode.Run` 本体の新規設計（契約は既存 `@ensure` を正とする）。
5. test 専用 GHA Secret / Drive folder inventory。

## 5. Contract

1. gate（`scripts/generator/test-integration.sh`）が Broad file を収集・実行し exit 0 になる。
2. Fetch 0 件なら Domain Error（`no_source_items`）。Drive 相当 upload は 0。TextWriter / Synthesize は呼ばない。
3. 成功時は TextWriter 1 回、Synthesize 回数は Greeting+Intro + 各 topic の Preface+Detail + Summary+Farewell、Write（Drive 相当）は 1 組。
4. TextWriter 失敗または Synthesize 失敗の後は Write しない。
5. credential 実値を error / log に出さない（dummy のみ）。

## 6. Constraints

1. secret / 実 vendor API / AgentSecrets / local keychain を使わない。
2. System build tag package（`apps/generator/test/system`）と `test-system.sh` を触らない。
3. 上位 Scope が下位 Scope の内部詳細を再 assert しない（minimization）。
4. 過去 Decision 本文を書き換えない。

## 7. Acceptance Criteria

1. [ ] AC-1: Broad Integration file が Integration gate で収集・実行される。
2. [ ] AC-2: 全 upstream 成功時、Drive 相当へ episode 成果物 1 組が届く。
3. [ ] AC-3: Source 0 件時、`no_source_items` で終了し書込・Writer・TTS 呼び出しが 0。
4. [ ] AC-4: TextWriter 失敗時、書込が 0。
5. [ ] AC-5: Synthesize 失敗時、書込が 0。
6. [ ] AC-6: Narrow / Unit が所有する境界 I/O・分岐網羅を Broad が再 assert しない。
7. [ ] AC-7: `./scripts/generator/test-integration.sh` が pass する。

## 8. Verification

```bash
./scripts/generator/check-static.sh
./scripts/generator/test-unit.sh
./scripts/generator/test-integration.sh
```

test 方針は `testing-strategy` を参照する。

## 9. Dependencies

1. blocked by: `ProduceEpisode.Run` 実装（現状 stub。lane D）。
2. blocked by: composition ProduceEpisode 結線（`docs/tasks/todo/generator-composition-produce-episode-wiring.md`）が Run を呼べる production graph を返すこと。
3. related: 既存 Narrow suite（helper 形の参照先。本 Issue は境界 I/O を奪わない）。

## 10. Risks

1. Narrow helper を Broad が import すると層逆向き依存になる — 共有が必要なら中立 support へ出す。
2. Run 未完のまま Broad だけ先に書くと Verification が再判断になる — Run 完了後に実装する。

## 11. Notes

1. GitHub Issue 化は別判断。本 file が達成契約の正。
2. System は lane D（`generator-system-e2e-produce-episode`）。本 Issue と混ぜない。
