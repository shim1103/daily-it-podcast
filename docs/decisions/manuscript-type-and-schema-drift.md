---
name: "playback の EpisodeManuscript 型と manuscript.schema.json の食い違いを、静的に検出するか"
description: "検出しない。ajv が実行時に schema で検証し、手書きの EpisodeManuscript（domain の語の型）は残す。型の生成、fixture の contract test、zod による接続は採らない。"
---

## 1. Decision

1. `EpisodeManuscript` 型と `manuscript.schema.json` の食い違いを、静的には検出しない。型は手書きのまま（domain の語 `EpisodeId`・`DurationSec`・`PositionSec` で書く）で、ajv が実行時に schema で検証する（`verifyManuscript`）。
2. 型の生成、fixture の contract test、`z.fromJSONSchema` による接続、zod からの JSON Schema の生成は、採らない。
3. 食い違いが実害として観測された時に、この Decision を更新して見直す。

## 2. Reason

1. 食い違いの実績が無い。実行時には ajv が、schema に適合しない原稿を弾く（`EpisodeContentError`）。静的な食い違いが起こすのは、型が現実とずれることで、データの破損ではない。
2. 型の生成は、素の基本型（`string`・`number`）を作る。手書きの型は domain の語で書いており、名前の SSoT（`domain-vocabulary-shared-by-web-and-worker.md`）と衝突する。`ajv.compile<T>` の `T` は呼び出し側が主張する型で、schema との一致は検査されない。型の一致を検査するだけに使う案でも、tool の導入、版と hash の固定、生成物の運用が要る。
3. 費用が、観測されていない問題に見合わない。shim が fixture の contract test も、費用の大きさから退けた。

## 3. Rejected

1. **JSON Schema から TS の型を生成し、手書きの型を置き換える案** — 上の Reason 2。名前の SSoT が崩れる。
2. **生成した型を、手書きの型との一致の検査にだけ使う案** — domain の語の型は守れるが、tool の導入と運用の費用が実績の無い問題に見合わない。
3. **schema に適合する fixture を型付きで書き、ajv で検証する contract test** — 費用が大きいと判断した。required な field の食い違いしか検出しない。
4. **`z.fromJSONSchema` で TS の検証 schema を実行時に作る案** — 検証は動くが、出力型が静的に推論されないので、型の接続にならない（zod 4.4.3 で実測）。
5. **zod で schema を書き、JSON Schema を生成して Go と共有する案** — SSoT が TS 側に移り、Go の generator が TS の build に依存する。言語横断の SSoT は JSON Schema のままにする。
