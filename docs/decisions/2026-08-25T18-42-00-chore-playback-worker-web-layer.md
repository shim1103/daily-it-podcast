---
name: playback の層境界は code（dir・depcruise・Drive schema）を正とし static gate する
date: 2026-08-25T19:20:00
branch: chore/playback-worker-web-layer
---

## 1. Decision

1. Feature / Primitive の dir・層 import・Drive 原稿検証の正は code にある（`web/src/components/{feature,primitive}/`・`apps/playback/.dependency-cruiser.mjs`・`worker/.../manuscript-schema.ts`）。本 Decision は値を写さない
2. 層違反は `dependency-cruiser` を `scripts/playback/check-static.sh` から実行して検知する。generator の depguard と同型（allow-list・Composition/Page 無制限）。ただし **Application → `apps/playback/contracts` は禁止**（Entities・Infrastructure は元から禁止。test は層 lint の対象外で、層ごとに差を付けない）。UseCase の入出力は Application が持つ型（`XxxUseCaseInput`・`XxxUseCaseOutput`）で受け、契約の型への写しは Controller が field ごとに行う。repo 根の `contracts/`（Drive の原稿 JSON Schema）は別物で、Application が読んでよい
3. 本 decision は次を上書きする: `2026-08-20T19-29-21-playback-web-layer-layout`（Feature/Primitive を dir 分割しない）、`2026-08-17T14-45-00` / `2026-08-18T14-35-00` / 旧 `DESIGN.md` §5.9（playback に層 lint を載せない）
4. 入力値の検証と値の変換（文字列から数値への変換・省略時の既定）は、route の zod schema（Hono の `zValidator`）が担う。Controller の Request → UseCaseInput は field の対応を写すだけにする。`architecture/backend/controller` の 3 段（Controller が Request → UseCaseInput を変換する）からの**意図的な逸脱**で、Hono の型推論を活かすためである

## 2. Reason

1. Feature と Primitive は import 規則が違う（Primitive は ViewModel 不可）。同一 dir のまま file 名例外で分けると、新 file 追加のたびに例外追記が必要になり、generator の「allow に無いものは止まる」と逆になる
2. 以前「分けない」とした理由は component file が 0 個の時点の YAGNI だった。いま Feature 複数と Primitive が共存するため、その却下理由は消えている
3. Biome は層 import の allow-list を持たない。ESLint を足すと静的検査が二重になる。`dependency-cruiser` は TS 側で depguard 相当を果たす
4. Application が契約（HTTP の schema の推論型）を import すると、契約の変更（field の追加・形式の変更）が use case と Port の型へ波及し、内側が外側を知る逆向きの依存になる。実例として、Port の `ProgressUpdatedEntry` が pull 契約の応答型から導かれ、`audioRef`（HTTP の path という wire の概念）が application の型に入って、application が `episodeAudioPath` を呼んでいた。skill（`architecture/backend/application` §7）は既に禁止していたが、application の 11 file が import していて、人手の確認では守られなかった。機械的に禁止する。以前の「generator 寄せ」は根拠にならない。generator の Application が読む `contracts` は repo 根の Drive 契約（原稿 JSON Schema）で、playback の HTTP 契約とは別物である
5. Drive 原稿に HTTP 専用 field は無い。HTTP schema から Drive 検証を導くと、HTTP 変更が Drive 読取を壊す。Drive の正は repo 根 `contracts/manuscript.schema.json`（`DESIGN.md`）
6. 型を application と contracts で別に持っても、SSOT は分裂しない。契約は wire の形、application の型は domain の形で、答える問いが違う。両者の食い違いのうち、Response の field の欠落は、Controller の写し（契約の型への代入）が型検査で検出する。入れ子の `map` 内の余分な field と Request 側の新 field は型検査が検出しないので、Controller の test が写しを field ごとに固定する。旧 Rejected 4 の「二重 SSOT」は、この写しが型検査される点で解ける
7. route の `zValidator` は、検証と同じ schema の `transform`・`default` で値を変換でき、`c.req.valid` が変換後の型を運ぶ。値の変換を Controller の Request → UseCaseInput に置くと、検証済みの値を再び解釈する段が増え、検証と変換が離れる。標準形から外れるが、Hono の仕組みに沿うほうが、変換の置き場が 1 箇所に定まる
8. test を層 lint の対象外にするのは、test と実装の整合が TDD の途中で一時的に崩れるのが普通で、全 test に import の禁止を課すのは現実的でないからである。契約の import 禁止は Application だけの規則ではない（Entities・Infrastructure も禁止）ので、Application の test だけを厳しくする理由も無い。test が契約や infrastructure を使うのは、test の責務の範囲として許す

## 3. Rejected

1. `components/` 直下のまま Primitive だけ path 例外 — 例外表が肥大し、新 file で穴が開く
2. Feature/Primitive を `web/src/feature/`・`web/src/primitive/` にして `components` を廃止 — 既存の Component 傘と page からの相対 path を広く壊す
3. ESLint + `eslint-plugin-boundaries` — Biome と linter が分裂する
4. Application → `apps/playback/contracts` を許可する案（旧答え。generator と読み手を揃える） — 契約の変更が use case と Port へ波及し、wire の概念（`audioRef`）が application に入った。generator が読むのは別物の Drive 契約で、揃える根拠にならなかった
5. Infrastructure が HTTP schema の omit を維持 — Drive と HTTP の混線が残る
6. 境界契約を markdown に置いて code と二重化する案 — 契約 SSOT が分裂する
7. 契約を Entities へ移して application と共有する案 — web も import する契約を、worker の内側の層に置けない
8. Controller の Request → UseCaseInput で値を変換する標準形 — route で検証済みの値を再び解釈する段が増え、検証と変換が離れる。Hono の型推論も活きない
