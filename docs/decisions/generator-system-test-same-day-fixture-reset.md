---
name: "system-test 再実行で同日 already_produced をどう避け、毎回 full path を通すか"
description: "test bucket の同日 date 一致 json（+ sibling mp3）を Run 前に削除し、不完全残骸も含める。episodeID 非空を必須。本番 DELETE 防護は caller / token-scope / workflow 結線に置き、Adapter の本番名 blacklist は採らない。本番の同日 skip・暦日・contracts は変えない。"
---

## 1. Decision

1. 本番の同日完成ペア skip（`ProduceEpisode.Run` の Fetch 前成功 return。Decision `2026-08-30T23-30-00`）は維持する。
2. system-test は Run 前に **test 用 R2 bucket** 上の、表示暦日が一致する json を削除し、sibling `{stem}.mp3` も無条件削除する（404 は成功）。完成ペア要件は無い。不完全残骸（例: json-only）も消す。別日の object は触らない。mp3-only（json 無し）は date が分からないため残す。
3. 合否は `runErr == nil` だけでは足りない。成功 produce は `episodeID` 非空を必須とし、空 ID の成功（`already_produced`）は Fail にする。
4. 成功 produce の直後に同じ `now` で 2 回目 Run し、空 ID・error 無し（実 R2 上の同日冪等）を確認する。
5. 暦日を偽らない。本番 code に system-test 専用分岐を入れない。`contracts/` に system-test flag を足さない。Compose に Lookup 差し替え口を足さない。
6. 本番 bucket への誤 DELETE 防護は、Adapter 内の本番名 blacklist ではなく、呼び出し側・token scope・workflow 結線が担う。`SameDayObjectCleaner` は渡された bucket を知り禁止しない。

## 2. Reason

1. 同日 skip は業務規則（quota 保護・1 暦日 1 本）であり、coverage のために弱めない。商用の分け方は「規則は本番のまま、test が自 tenant の前提状態を reset する」。
2. Drive 時代の Decision `2026-09-02T16-57-00` §1-4 は「run 前掃除、または skip を明示 fail」だった。R2 移行後の full system-test は cleanup も fail も無く、`already_produced` を緑にしていた（false green）。意図を R2 test bucket で復元する。
3. date を偽ると Fetch 窓と表示暦日の意味が壊れ、本番と同条件で測れない。contracts / Compose に test flag を足すと test・prod instance の契約が混ざる。
4. Lookup を常に偽の double にすると、実 `HasPair` 両分岐と本番結線の一致が検証できない。fixture reset なら本番結線のまま両分岐（Run1=偽→full path、Run2=真→skip）を観測できる。
5. 完成ペアに限ると、json-only 等の不完全残骸が残り、次回 Run が `already_produced` や中途半端な状態で false green / 不安定になりうる。同日 date 一致の json を起点に sibling mp3 まで消すと、不完全残骸も含めて full path 前提を復元できる。他日成果物の履歴は消さない。
6. Adapter に本番 bucket 名を埋め込むと、名前変更・別名・誤検知で契約が壊れ、Infrastructure が運用上の「本番」定義を持つことになる。防護の正は credential / token の到達範囲と、system-test を test bucket へ結線する workflow 側にある。

## 3. Rejected

1. **date に非暦日や一意 key を入れる案** — 表示 date の意味を汚す。Fetch 窓もずれる。
2. **本番 `ProduceEpisode` / Compose に system-test 専用分岐や Lookup override 口を足す案** — test 都合の本番分岐。Decision `2026-09-02T16-57-00` Rejected と同型。
3. **`contracts/` や Config に system-test flag / prefix を足す案** — 公開契約と prod/test instance 分離を汚す。今の製品に multi-tenant prefix 要求は無い（YAGNI）。
4. **skip を明示 Fail するだけの案** — false green は止まるが、再 run が赤になるだけで「毎回 full path」を満たさない。
5. **run ごと ephemeral bucket を作る案** — account 級 token と作成失敗 mode が増える。現行の test bucket + 同日削除で足りる。
6. **成果物を一切残さず全件 DELETE する案** — 別日・調査用残骸まで消す。要求は同日 skip 回避だけ。
7. **同日完成ペア（json+mp3）だけを削除し、不完全残骸は触らない案** — 以前の答え。user が狭い scope を却下。不完全残骸が残ると fixture reset が足りない。
8. **Adapter の `ProductionBucketName` panic（本番名 blacklist）案** — 以前の答えの一部。本番名を Adapter が知る・禁止するのは運用定義の誤帰属。防護は caller / token-scope / workflow 結線へ移す。
9. **repo `.vscode` の gopls settings で防護する案** — IDE 設定は実行時契約でも CI 結線でもない。別 spawn の責務でも、本問いの防護手段としては採らない。
