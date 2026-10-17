---
name: "task-feature/playback-progress-client-at-skew"
description: "feature/playback-progress-client-at-skewで進行中のtask。"
when_to_use: "feature/playback-progress-client-at-skewで作業を再開する時、PRを作る前に未完了を確かめる時。"
non-scope: "branch非依存の未完了とOpen questions（→wiki）、完了済みのtask、sessionの経緯（→daily）。"
---

1. PRを作る（shim指示待ち）。merge後に#217を閉じる
   - skew検査は`911804e`でpush済み。baseは親epic（`feature/playback-progress`）想定
2. worktree直下の`_review_*.txt`／`_review_*.json`を消す（agentは`rm` deny。shimが実行）
3. 未追跡の`apps/playback/pnpm-lock.yaml`の採否を決める
   - 正本は`package-lock.json`（Decision）。`pnpm install`副作用の可能性。削除かignoreかcommitしないまま放置かを決める
4. `assertClientAtWithinSkew`の`clientAt: ClientAt`と`serverAt: Date`の型非対称を揃えるか決める（新規・未実装）
   - 選択肢は調査済み（共通`UtcInstant`化／別名`ServerAt`／現状維持）。実装はshimが案を選んだ後
5. 事前失敗の`contracts/http.sociable_unit.test.ts`（`pads_to_fixed_width_when_clientAt_omits_seconds`）を別途直すか決める
