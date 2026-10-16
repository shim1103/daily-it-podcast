---
name: "task-feature/playback-progress-versioned-persistence"
description: "feature/playback-progress-versioned-persistenceで進行中のtask。"
when_to_use: "feature/playback-progress-versioned-persistenceで作業を再開する時、PRを作る前に未完了を確かめる時。"
non-scope: "branch非依存の未完了とOpen questions（→wiki）、完了済みのtask、sessionの経緯（→daily）。"
---

1. PRのmerge（shim）を待つ。`merged`を受けたら、merge済みを確認し、#218を`gh issue close`する
    - PRは1つにまとめた（範囲外の変更はDeviationsに記載）。baseは`feature/playback-progress`
2. 2026-10-08のdailyの「#218・#219・#220の本文は更新済み」は誤り（実際は未更新で、今回更新した）。過去の記録は書き換えない方針で、PRのDeviationsに残した
3. 小文字の`todo:`（ts系13箇所）を`TODO:`に揃えるか決める（follow-up）
