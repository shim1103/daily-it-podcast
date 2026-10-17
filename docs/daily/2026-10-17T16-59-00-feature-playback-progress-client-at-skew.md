---
name: Issue #217 の skew 検査を実装し、serverAtInstant 改名・秒省略契約の修復まで進めて PR 前まで閉じた
date: 2026-10-17T16:59:00
session_id: playback-progress-client-at-skew-2026-10-17
branch: feature/playback-progress-client-at-skew
prev: なし
---

## 1. Summary

#217 の達成契約を最新 codebase に合わせて更新し、create／update／complete の I/O 前 skew 検査と Clock 注入を実装・push した。続けて引数名を `serverAtInstant` に改め、秒省略 `clientAt` を契約が再び受理するよう直した。`pnpm-lock.yaml` は既に作業木から消えていた。PR 作成はこの記録の直後。

## 2. Changes

1. Issue #217 の Context／Scope を、complete の双 repository 引数・版つき Port 未切替（#219）・`createClock` の `todo:` 残存に合わせて更新した。
2. UseCase sociable と composition で境界値・I/O 非呼び出し・ValidationError 写像を確認した。scoped unit は緑。全 unit の事前失敗（秒省略）は後続 commit で解消した。
3. 過去側上限は 24h・未来側は 300s（5分）。「5分過去＝retry」は現定数では拒否しない。Decision の仮値を変えていない。
4. MEMORY は空だった（削除対象なし）。

### Commits

- `911804e`
- `47cf459`
- `b0c9c97`
