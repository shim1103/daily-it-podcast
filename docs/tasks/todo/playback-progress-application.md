# feat(playback): 進捗Application（Write/pull mergeとlist embed）

## 1. Summary

このIssueでは、create／update／complete／pull の merge・404・冪等・完走ゾーンをPort doubleで実装し、`listEpisodes` のprogress embedを本実装する。D1 SQL／web／常設smokeは別Issue。

## 2. Context

1. 事実: use case／controller／routeのsignatureと足場はA済み。中身はzero／null embed。
2. 事実: embedはworker Applicationの責務であり、web sync Issueへ統合しない（runtimeが違う）。
3. 事実: list向け初回はembed・1往復、progress応答はWorkers Cacheに載せない（既存Decision）。

## 3. Canonical Sources

1. A: progress use case／`listEpisodes` signature・contracts schema
2. B: `docs/decisions/2026-09-22T19-13-35-feature-playback-progress.md`（merge・playing/stopped）
3. B: `docs/decisions/2026-09-22T18-58-38-feature-playback-progress.md`（冪等・404）
4. B: `docs/decisions/2026-09-19T16-26-00-feature-playback-progress.md`（embed）
5. B: `docs/decisions/2026-09-19T19-12-01-feature-playback-progress.md`（embed no-store）
6. B: `docs/decisions/2026-09-22T19-29-15-feature-playback-progress.md`（Cacheに載せない）
7. test方針: `testing-strategy`

## 4. Scope

### In Scope

1. Write／pull use caseの振る舞い（先勝ち／後勝ち・冪等・404・完走ゾーン）
2. `listEpisodes` のprogress embed（todo／null埋めの除去）
3. A足場test→behavior置換、申し送りcomment削除
4. Port doubleで閉じるVerification（実D1不要）

### Out of Scope

1. D1 adapter／local peer／Fake（`playback-d1-progress-infra`）
2. web VM同期・UI描画
3. 一過性dispatch・常設smoke
4. skew／push間隔／短尺の数値確定（lane）

## 5. Contract

既存HTTP／schema契約を満たす。未固定の公開型を新設しない。

## 6. Constraints

1. 数値（skew・間隔・短尺）はlaneの仮値でよい。正本化は別決定。
2. progress付きlist応答をWorkers Cacheに載せない。

## 7. Acceptance Criteria

1. [ ] create／update／complete／pull がBのmerge／冪等／完走意味を満たす（behavior test）
2. [ ] `listEpisodes` がprogressをembedする（行なしはnull）
3. [ ] 該当A足場がbehaviorに置換され、申し送りが消えている
4. [ ] VerificationがPort doubleのみで閉じる（実D1必須にしていない）

## 8. Verification

1. use case／list sociable（StubProgressRepository等）
2. playback unit／typecheck

## 9. Dependencies

1. 本番経路での読取確認を強めるなら `playback-d1-progress-infra` 後がよい。完了条件自体はPort doubleで独立可。

## 10. Risks

1. embedとWriteを同居させすぎてACが膨らむ → WriteとembedでVerificationを節分けする

## 11. Notes

web sync／UIへembedを混ぜない。
