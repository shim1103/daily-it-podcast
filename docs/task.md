---
name: "task-feature/playback-progress-versioned-persistence"
description: "feature/playback-progress-versioned-persistenceで進行中のtask。"
when_to_use: "feature/playback-progress-versioned-persistenceで作業を再開する時、PRを作る前に未完了を確かめる時。"
non-scope: "branch非依存の未完了とOpen questions（→wiki）、完了済みのtask、sessionの経緯（→daily）。"
---

1. playback/workerの同形の型・写しを削る範囲をshimと決める
    - 調査済み（editなし）。候補は`ProgressMergeIncoming`の再構築3箇所の削除、progress→contract写像の1関数化、`ProgressWriteUseCaseOutput`を`Pick<EpisodeProgress, ...>`へ寄せること
    - `EpisodeManuscript`と`EpisodeItem`の統合は`application.md`§7-2との関係で慎重に扱う。shimの許可待ち
2. 時刻のUTC固定幅前提を型へ載せるか決める
    - `contracts/`が`entities/`をimportしてよいかが先に要る。`2026-08-25T18-42-00`のDecisionは未読
    - `route-handler.md`の標準形（Controllerがparse）と実装（route zod）の逸脱をDecisionに残すか
3. `progress-repository.ts`の`ProgressCursor`のdocを直す
    - 「変換と形式の検証はController」と書いてあるが、Decision `2026-10-01T23-32-40`は「routeのzod。Controllerは数値→文字列の戻しだけ」と言う
4. epic branch名の確認
    - `origin/feature/playback-progress`と推定。`gh`がTLS errorで取れず未確認
