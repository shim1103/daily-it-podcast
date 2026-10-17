---
name: "task-develop"
description: "developで進行中のtask。"
when_to_use: "developで作業を再開する時、PRを作る前に未完了を確かめる時。"
non-scope: "branch非依存の未完了とOpen questions（→wiki）、完了済みのtask、sessionの経緯（→daily）。"
---

1. `agent-standards`のskills変更をshimが確認する
    - 変更はworking treeのまま（`/commit`はshimの指示待ち）。対象は`coding-terms`（`visibility.md`新設・`module-design.md`）、`test-terms`、`go`、`swift`、Decision `file-split-and-visibility`
2. `2:platform/typescript`へ可視性の写像を足すか、shimに諮る
    - 今回の対象（`test-terms`・`coding-terms`・`swift`・`go`）に`typescript`は無く未着手。fileの分割と`export`が結合するのはTSなので、写像の欠けが最も効く言語
