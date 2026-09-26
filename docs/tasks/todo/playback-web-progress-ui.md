# feat(playback): 進捗list表示（配置意味・描画）

## 1. Summary

このIssueでは、完走印・位置・`lastPlayedAt`の配置意味に沿ってlist行の描画（slot埋め・必要なCSS）を完了する。HTTP／retry／derive所有は持たず、渡された表示propsだけを描く。

## 2. Context

1. 事実: UI信号の意味と配置方針はDecision済み。見た目polishはAにしない。
2. 事実: FeatureはHTTP／deriveを持たない（同期Issue／今日Decision）。

## 3. Canonical Sources

1. B: `docs/decisions/2026-09-19T16-25-18-docs-playback-audio-history.md`（UI信号意味）
2. B: `docs/decisions/2026-09-25T05-53-19-docs-playback-audio-history.md`（配置の意味・描画はAにしない）
3. frontend層: `1:terms/architecture/frontend`
4. test方針: `testing-strategy`

## 4. Scope

### In Scope

1. 日付右＝完走印、下段＝位置＋`lastPlayedAt`に沿う描画
2. slot埋め・必要なCSS／bar（design未確定の過剰polishは避ける）
3. 渡された表示propsのみ描くこと

### Out of Scope

1. progress HTTP・retry・pull間隔
2. D1／worker
3. 常設smoke／e2e拡張（release Issue）
4. page-hook公開面の契約凍結

## 5. Contract

未固定の公開API型を新設しない。既存Featureの表示境界を守る。

## 6. Constraints

1. FeatureにHTTP／deriveを載せない。
2. 色・寸法の百科をDecisionに増やさない。

## 7. Acceptance Criteria

1. [ ] 未play／途中／完走後の表示が配置意味どおりに分かれて見える
2. [ ] Featureがprogress HTTP／deriveを持たない
3. [ ] 同期Issueで更新される共有状態を描画できる

## 8. Verification

1. component／page系のsociableまたは既存test慣習に沿った確認
2. typecheck

## 9. Dependencies

1. blocked by: `playback-web-progress-sync`（共有状態の更新契機）

## 10. Risks



## 11. Notes

releaseの常設smoke／e2eは別Issue。
