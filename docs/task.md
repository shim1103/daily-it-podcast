---
name: "task-feature/playback-progress-client-at-skew"
description: "feature/playback-progress-client-at-skewで進行中のtask。"
when_to_use: "feature/playback-progress-client-at-skewで作業を再開する時、PRを作る前に未完了を確かめる時。"
non-scope: "branch非依存の未完了とOpen questions（→wiki）、完了済みのtask、sessionの経緯（→daily）。"
---

1. PRを作る（shim指示待ち）。merge後に#217を閉じる
   - skew検査は`911804e`でpush済み。baseは親epic（`feature/playback-progress`）想定
2. worktree直下の`_review_*.txt`／`_review_*.json`を消す（agentは`rm` deny。shimが実行）
3. 未追跡の`apps/playback/pnpm-lock.yaml`の採否を決める（本Issue外の`pnpm`副作用）
4. 事前失敗の`contracts/http.sociable_unit.test.ts`（`pads_to_fixed_width_when_clientAt_omits_seconds`）を別途直すか決める
5. 小文字の`todo:`（worker内の残り）を`TODO:`に揃えるか決める（follow-up。CAS／cursor置換の申し送り含む）
