---
name: "task-feature/playback-progress-client-at-skew"
description: "feature/playback-progress-client-at-skewで進行中のtask。"
when_to_use: "feature/playback-progress-client-at-skewで作業を再開する時、PRを作る前に未完了を確かめる時。"
non-scope: "branch非依存の未完了とOpen questions（→wiki）、完了済みのtask、sessionの経緯（→daily）。"
---

1. shim の merge を待つ。`merged` を受けたら merge 済みを確認し、#217 を `gh issue close` する
   - PR 作成済み（base `feature/playback-progress`）。skew・`serverAtInstant`・秒省略契約修復まで commit 済み
