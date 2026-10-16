---
name: "web と worker が共有する http 非依存の語彙は、どこに置き、不変を何で守るか"
description: "共有語彙は apps/playback/domain/ に置き、時刻は brand 型 UtcInstant にする。不変は domain の関数を通った値だけが型を得ることで守り、各層は再検証せず、未知の入力が既知になる境界だけが保証する。"
---

## 1. Decision

1. web・worker・contracts が共有する、http 非依存の語彙と不変は、`apps/playback/domain/` に置く。外部依存ゼロの純粋 TS（zod も import しない）で、domain は他のどの dir も import しない。worker の `entities/` とは並立の最内側で、互いに import しない。
2. 時刻は brand 型 `UtcInstant`（UTC 固定幅の ISO 表記）にする。文法・暦の存在・instant の範囲の判定と正規化の正本は、domain の `normalizeUtcInstant` とする。この関数を通った値だけが `UtcInstant` になる。
3. `contracts/` は HTTP の wire 形だけを持つ。時刻の判定は domain の関数を呼び、検証失敗を HTTP の issue に写す。Request／Query は、client が送る wire 形（入力型）と、route の zod が parse した後の形（`Parsed…`）を別の型にする。
4. 不変の信頼は型が運ぶ。UseCase・Port・Controller は時刻を再検証しない。保証するのは、未知の入力が既知になる境界だけとする。route の zod、原稿 JSON の ajv、保存先から読む D1 adapter（`isUtcInstant` で正規形を確認し、満たさなければ Infrastructure Error）である。
5. brand にするのは、取り違えると静かに壊れる値に限る。時刻はその一つで、`episodeId` や cursor は本 Decision の範囲外（別の問い）とする。

## 2. Reason

1. 時刻は `string` で、`mergeFields` が文字列の `<` で大小を比べる。比較が正しいのは固定幅の UTC 表記を前提にするからだが、前提は型に無く、contracts の transform・entity の doc・`@require` の複数箇所に文章で散っていた。入口の関数を通った値だけが型を得る形にすれば、前提の正本が1箇所になり、通らない値は compile で弾かれる。
2. 語彙を worker の `entities/` に置くと、web が worker の内部 dir を runtime で import することになる。web は応答 schema を値で import して parse しており（`web/src/api/playback-api-client.ts`）、schema が domain の関数を呼べば、その関数が web の bundle に入る。web・worker が並んで読める最内側の dir が要る。
3. 依存の向きは `contracts → domain` で、内側へ向かう。HTTP の変更が domain を巻き込むのは、共有する語彙自体が変わる時だけになる。`2026-08-25T18-42-00-chore-playback-worker-web-layer.md` が起こした実害（wire の概念 `audioRef` が application の型へ入った）は、wire 形を domain に置かない限り再発しない。
4. 各層で再検証しない理由は、検証の責務を境界に1つずつ置いて SSoT を保つためである。再検証すると、同じ判定が複数箇所に複製され、片方だけの修正で食い違う。型が信頼の根拠を運ぶので、再検証は要らない。D1 の読出し確認は、自分が書いた値でも保存先は外部であり、未知の値が既知の型になる境界だから、再検証ではなく保証である。
5. 入力の文法まで domain が持つのは、`Date.parse` が offset の無い文字列を実行環境の時刻帯で解釈し、結果が環境で変わるからである。`Date.UTC` は存在しない暦日（2 月 30 日）や 24 時を黙って繰り上げるので、往復比較で弾く。契約が現に受理する形（秒の省略、6 桁の小数秒の切り捨て）は、既存の契約 test を正として保つ。
6. Request を入力型と検証後の型に分けるのは、client が送る値は文字列で、検証後だけ `UtcInstant` になるからである。分けないと、web は送る前に brand を作る必要が出る。

## 3. Rejected

1. **語彙を worker の `entities/` に置く案** — web が worker の内部 dir へ runtime で依存する（上の Reason 2）。
2. **契約全体を Entities へ移す案** — `2026-08-25T18-42-00-chore-playback-worker-web-layer.md` が却下済みで、本 Decision は語彙だけを分ける別の案である。wire 形（`audioRef`・path・Request にその path param を足した形）は contracts に残る。
3. **brand にせず、`@require` の文章で前提を述べ続ける案**（現状） — 前提が型に無く、複数箇所に散り、文字列比較の誤動作が compile で検出できない。
4. **時刻を `Date` 型や epoch ms の `number` にする案** — wire の形（ISO 文字列）と D1 の列（TEXT）を変える移行になる。本 Decision の目的（前提を型で運ぶ）に対して費用が大きい。
5. **UseCase や adapter が時刻を再検証する案** — 判定が複数箇所に複製される（上の Reason 4）。
6. **`z.iso.datetime` を入口に残し、domain は範囲と正規化だけを持つ案** — 「offset が必須で文法が正しい」という知識が contracts と domain に分かれる。domain は単体で呼ばれても環境に依らず安全であるべきなので、文法ごと domain が持つ。
