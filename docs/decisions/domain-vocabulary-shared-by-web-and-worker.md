---
name: "web と worker が共有する http 非依存の語彙は、どこに置き、不変を何で守るか"
description: "共有語彙は apps/playback/domain/ に置き、時刻は brand 型 UtcInstant にする。不変は domain の関数を通った値だけが型を得ることで守り、各層は再検証せず、未知の入力が既知になる境界だけが保証する。"
---

## 1. Decision

1. web・worker・contracts が共有する、http 非依存の語彙と不変は、`apps/playback/domain/` に置く。外部依存ゼロの純粋 TS（zod も import しない）で、domain は他のどの dir も import しない。worker の `entities/` は domain の語彙を型に使ってよい（entities → domain の一方向。domain は何も import しない）。
2. 時刻は brand 型 `UtcInstant`（UTC 固定幅の ISO 表記）にする。固定幅への変換と instant の範囲の正本は、domain の `toUtcInstant`（入力は `Date`）とする。この関数を通った値だけが `UtcInstant` になる。入力文字列の文法（offset 付き ISO、実在する暦日）は domain が持たず、contracts の zod が持つ。
3. `contracts/` は HTTP の wire 形だけを持つ。時刻は zod の `z.iso.datetime({ offset: true })` で文法を検証したあと、domain の `toUtcInstant` を呼び、範囲外を HTTP の issue に写す。Request／Query は、client が送る wire 形（入力型）と、route の zod が parse した後の形（`Parsed…`）を別の型にする。
4. 不変の信頼は型が運ぶ。UseCase・Port・Controller は時刻を再検証しない。保証するのは、未知の入力が既知になる境界だけとする。route の zod、原稿 JSON の ajv、保存先から読む D1 adapter（`isUtcInstant` で正規形を確認し、満たさなければ Infrastructure Error）である。
5. domain に置く語彙は、時刻（brand 型 `UtcInstant`）、`EpisodeId`・`PositionSec`・`ClientAt`（別名）、`ProgressReport`（端末からの**進捗報告**。`positionSec` と `clientAt`）、`EpisodeProgress` とする。基本型で表す domain の語は、型の名前を domain に置く。brand にするのは、取り違えると静かに壊れる値に限り、今は時刻だけである。別名の三つは名前の SSoT で、不変（非空・非負）は入口の zod と JSON Schema が保証する。cursor は範囲外とする。
6. UseCase の入出力・Port の引数と戻り値（束）は application が持ち、domain の語彙と layer 固有の都合から合成する（交差型・`Pick`）。field を再宣言しない。worker だけが使う語彙（`EpisodeManuscript`・`EpisodeAudio`）は worker の `entities/` に置く。

## 2. Reason

1. 時刻は `string` で、`mergeFields` が文字列の `<` で大小を比べる。比較が正しいのは固定幅の UTC 表記を前提にするからだが、前提は型に無く、contracts の transform・entity の doc・`@require` の複数箇所に文章で散っていた。入口の関数を通った値だけが型を得る形にすれば、前提の正本が1箇所になり、通らない値は compile で弾かれる。
2. 語彙を worker の `entities/` に置くと、web が worker の内部 dir を runtime で import することになる。web は応答 schema を値で import して parse しており（`web/src/api/playback-api-client.ts`）、schema が domain の関数を呼べば、その関数が web の bundle に入る。web・worker が並んで読める最内側の dir が要る。
3. 依存の向きは `contracts → domain` で、内側へ向かう。HTTP の変更が domain を巻き込むのは、共有する語彙自体が変わる時だけになる。`2026-08-25T18-42-00-chore-playback-worker-web-layer.md` が起こした実害（wire の概念 `audioRef` が application の型へ入った）は、wire 形を domain に置かない限り再発しない。
4. 各層で再検証しない理由は、検証の責務を境界に1つずつ置いて SSoT を保つためである。再検証すると、同じ判定が複数箇所に複製され、片方だけの修正で食い違う。型が信頼の根拠を運ぶので、再検証は要らない。D1 の読出し確認は、自分が書いた値でも保存先は外部であり、未知の値が既知の型になる境界だから、再検証ではなく保証である。
5. 文法を contracts の zod に任せ、domain の入力を `Date` にするのは、再発明を避けるためである（`selection-terms/criteria.md` 第1基準）。`z.iso.datetime({ offset: true })` は、実在しない暦日・24 時・60 秒・`+24:00`・offset なし・日付のみを全て拒否することを実測した。その後なら `new Date(v)` が正しく換算する。domain が `Date` だけを受ければ、offset の無い文字列を実行環境の時刻帯で解釈する危険（`Date.parse` の挙動）が domain に入らない。`isUtcInstant` は `toUtcInstant(new Date(v)) === v` の往復で足りる（2 月 30 日は 3 月へ繰り上がり、一致しない）。
6. 小数が 4 桁以上の表記は、Date Time String Format（ECMA-262）が規定せず、実装依存の恐れがある。contracts が 3 桁へ切ってから `Date` に渡す（Node では切り捨てを実測。workerd は未確認）。契約が現に受理する形（秒の省略、6 桁の小数秒の切り捨て）は、既存の契約 test を正として保つ。
7. 語彙を domain に置く基準は、web・worker・contracts のうち2つ以上が使うことである。束を domain に置くと、最内側（web も読む）が Port と UseCase の形を知る逆向きの依存になる。束の field は domain の語彙そのものなので、application が合成すれば、`episodeId` の型のような知識が1箇所に定まる。同じ理由で、worker の `EpisodeManuscript.episodeId` も `EpisodeId` を使う。そのため entities → domain の import を許す（domain は外部依存ゼロで何も import しないので、向きは崩れない）。
8. `EpisodeId` などを別名にする理由は、brand にする根拠（`entities.md` §10-1 の「同じ不変が繰り返し破られた」証拠）が無いことと、非空を domain が持つと入口の zod と二重になることである。別名なら test の literal を変えずに、名前の SSoT を得られる。後で brand に上げる時の差分は、literal だけになる。
9. `EpisodeAudio` を domain に置かないのは、web が音声 byte を型として持たない（`web/src` に `Uint8Array`・`ArrayBuffer`・`Blob` の出現が無いことを実測）ため、「2つ以上が使う」を満たさないからである。
10. Request を入力型と検証後の型に分けるのは、client が送る値は文字列で、検証後だけ `UtcInstant` になるからである。分けないと、web は送る前に brand を作る必要が出る。

## 3. Rejected

1. **語彙を worker の `entities/` に置く案** — web が worker の内部 dir へ runtime で依存する（上の Reason 2）。
2. **契約全体を Entities へ移す案** — `2026-08-25T18-42-00-chore-playback-worker-web-layer.md` が却下済みで、本 Decision は語彙だけを分ける別の案である。wire 形（`audioRef`・path・Request にその path param を足した形）は contracts に残る。
3. **brand にせず、`@require` の文章で前提を述べ続ける案**（現状） — 前提が型に無く、複数箇所に散り、文字列比較の誤動作が compile で検出できない。
4. **時刻を `Date` 型や epoch ms の `number` にする案** — wire の形（ISO 文字列）と D1 の列（TEXT）を変える移行になる。本 Decision の目的（前提を型で運ぶ）に対して費用が大きい。
5. **UseCase や adapter が時刻を再検証する案** — 判定が複数箇所に複製される（上の Reason 4）。
6. **domain が入力文字列の文法・暦・offset 算術まで手書きで持つ案**（旧答え） — zod の検査と二重で、約70行の再発明になる。当時は「domain を単体で呼んでも時刻帯に依らず安全にする」ことを理由にしたが、入力を `Date` にすれば解消するので、根拠を失った。
7. **domain が zod に依存し、schema と `.brand()` で型と検査を1箇所に持つ案** — 手書きを消す点は `Date` を受ける入口でも達せられる。最内側が zod 4 に結合し、zod 3→4 で brand・iso・error の API が変わった前例がある。`entities.md` §7 の「外部依存ゼロ」の目的が test double を要しないことだけなら zod は許されるが、字面の意図は確定できない。依存の小ささ（`criteria.md` 第3基準）を優先する。
8. **`EpisodeId` を brand にし、非空の入口を domain に置く案** — 非空を domain が手書きで持つと、入口の zod の `min(1)` と二重になる。zod で守るには domain が zod に依存する必要があり、7 で却下した。brand にする根拠（繰り返し破られた証拠）も無い。test の literal は166箇所、入口は4つで、費用が便益に見合わない。
9. **UseCase の入出力・Port の引数と戻り値（束）を domain に置く案** — 最内側が Port と UseCase の形を知る逆向きの依存になる（Reason 7）。
10. **`Uint8Array` の別名（音声）を domain に置く案** — web が使わず、platform の型の別名にすぎない（Reason 9）。worker の `entities/` に `EpisodeAudio` を置く。
