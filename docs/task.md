---
name: "task-develop"
description: "developで進行中のtask。"
when_to_use: "developで作業を再開する時、PRを作る前に未完了を確かめる時。"
non-scope: "branch非依存の未完了とOpen questions（→wiki）、完了済みのtask、sessionの経緯（→daily）。"
---

1. `agent-standards`のpushをshimが判断する
    - mainにcommit済み（ahead 24）。`--push`は指示されていないため未push
2. wiki #192のB33（importを入口fileだけに絞る規則）の扱いをshimが決める
    - Backlogに載せた。`2:platform/typescript/visibility.md`が写像を持つので、課す場合は既存の`dependency-cruiser`で決める
