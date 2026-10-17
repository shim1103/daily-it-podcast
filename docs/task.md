---
name: "task-develop"
description: "developで進行中のtask。"
when_to_use: "developで作業を再開する時、PRを作る前に未完了を確かめる時。"
non-scope: "branch非依存の未完了とOpen questions（→wiki）、完了済みのtask、sessionの経緯（→daily）。"
---

1. file分割・可視性・test単位のskills修正をshimへ諮る
    - 調査済み。module/packageの言語差、可視性3段階、test単位の振る舞い基準、private helperのC→A→B判定を提示した
    - 修正案は3つ（coding-termsへ可視性追記、test-terms/naming-and-layoutの単位修正、Decision `file-split-and-visibility`）。shimの許可待ち
    - 語彙を言語中立（file間/package内/外部）にするかGo寄りにするか、shimの回答待ち
    - 追加で回答済み：export/公開の2段、責務での分割とSRPの粒度、index/helperのtest分担とcoverage、table-driven
2. `/std`の実行はshimの再指示待ち
    - 調査・提示のみのactionに切り替わったため未実行
3. `/wiki --update backlog`で「import境界lint導入（index.ts経由のみ）」を載せる
    - Issue更新の副作用があり、調査・提示のみのactionでは未実行。実行指示待ち
    - 設計上の注意：helper_test.tsからhelper.tsへの直接importを違反にしない（test除外、またはdir内importの許可）
