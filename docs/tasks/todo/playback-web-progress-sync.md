# feat(playback): 進捗web同期（playing push／stopped pull／retry）

## 1. Summary

このIssueでは、既存ApiClient（1試行）を使い、ViewModel側でplaying中push・stopped中pull・有限retry・成功後の共有UI更新を実装する。FeatureにHTTP／deriveを載せない。見た目polishは別Issue。

## 2. Context

1. 事実: ApiClientのcreate／update／complete／pullはA済み（1試行。retryはVM）。
2. 事実: page／compose公開面・hook分割は凍らせない（Decision `2026-09-25T05-53-19`）。
3. 事実: Write失敗はuser通知なし・`console.error`・sessionで捨てる（Decision `2026-09-19T19-12-30`）。

## 3. Canonical Sources

1. A: `apps/playback/web/src/api/playback-api-client.ts`（1試行）
2. B: `docs/decisions/2026-09-22T19-13-35-docs-playback-audio-history.md`
3. B: `docs/decisions/2026-09-19T19-12-30-docs-playback-audio-history.md`
4. B: `docs/decisions/2026-09-25T05-53-19-docs-playback-audio-history.md`
5. frontend層: `1:terms/architecture/frontend`
6. test方針: `testing-strategy`

## 4. Scope

### In Scope

1. playing: pushのみ（応答positionでseek引き戻ししない）
2. stopped: stop時push＋定期pull（間隔は仮値可）
3. 有限retry・`console.error`・成功後に共有進捗UI更新
4. Feature（Row／Item）にHTTP／progress deriveを載せない

### Out of Scope

1. 配置意味どおりの見た目・CSS／bar polish（UI Issue）
2. worker merge／D1／embed
3. page-hook公開型の契約凍結、hooks file分割の確定
4. 常設smoke／e2e拡張

## 5. Contract

既存ApiClient signatureを変えない（retryをclientに戻さない）。

## 6. Constraints

1. 公開compose面をAとして凍らせない。B制約下の最小配線。
2. push／pull間隔の正本数値はlane。仮値でACを閉じる。

## 7. Acceptance Criteria

1. [ ] playing中はpushのみで、負けたpushでローカルseekを引き戻さない
2. [ ] stopped中にstop pushとpullが動く
3. [ ] Write失敗時にuser向け失敗UIが無く、有限retryと`console.error`がある
4. [ ] 共有進捗UIはWrite成功後に更新される
5. [ ] Featureがprogress HTTP／deriveを持たない

## 8. Verification

1. web view-model sociable（fetch double）
2. typecheck／関連unit

## 9. Dependencies

1. worker緑はVerification強化用。ApiClient A済みのため着手は独立可。
2. UI Issueは本Issueの後がよい（共有状態の更新契機が先）。

## 10. Risks

1. hook所有権未決で配線が仮になる → 公開面を契約化せず最小変更に留める

## 11. Notes

follow-up: `playback-web-progress-ui`
