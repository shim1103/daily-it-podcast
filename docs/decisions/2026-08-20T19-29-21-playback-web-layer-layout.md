---
name: playback web は frontend 7 層を dir へ 1 対 1 対応させ、Pico.css を npm 依存として main.ts で 1 回 import する
date: 2026-08-20T19:29:21
branch: docs-status-audit
---

## 1. Decision

1. `apps/playback/web/src/` は frontend skill の role と dir を 1 対 1 対応させる。`view-models/` と `lib/` を新設する

   | role | dir |
   |---|---|
   | Page / Route | `main.ts` + `pages/` |
   | Feature / Primitive Component | `components/` |
   | ViewModel層 | `view-models/` |
   | API Client | `api/` |
   | 純粋関数層 | `utils/` |
   | External Dependencies層 | `lib/` |

2. Component は `feature/` と `primitive/` に dir 分割せず、`components/` 直下へ置く。役割は file 名と import 制約で表す
3. ViewModel は React hook ではなく、state を閉じた module と購読関数で表す。role は Use Cases ring のまま変えない
4. Component は `HTMLElement` を返す関数として書く。JSX・仮想 DOM を導入しない
5. Pico.css は `@picocss/pico` を npm dependency として入れ、classless 版を `main.ts` で 1 回だけ import する
6. Vite の entry は `apps/playback/web/index.html` に置く。mount 先は `<main id="app">` とし、`class` 属性を書かない

## 2. Reason

1. 層と dir が一致していれば、どこへ置くかの判断が毎回同じ結論になる（Least Astonishment）。後続の web 層違反検知（PR-G）も dir を境界としてそのまま書ける
2. component の 2 dir 分割は、現在 file が 0 個の段階では境界より indirection を先に増やす。Feature か Primitive かは import 制約で判定でき、dir に頼らなくても検知できる（KISS / YAGNI）
3. React を使わない決定は `2026-08-18T11-12-00-feature-playback-web.md` で確定済み。skill の `hooks` は platform 実装の呼称であり、role 自体は runtime に依らない。dir 名を `hooks/` にすると存在しない React を示唆する
4. 一覧・再生・原稿表示に必要な DOM 更新は差分描画を必要としない。JSX は transform と runtime を要求する（Least Power）
5. CDN の `<link>` は実行時に外部 host へ依存し、Access 内配信という前提と噛み合わない。npm 依存なら Vite が bundle へ取り込み、version が lockfile に固定される
6. classless 版は semantic tag を選択子にする。`div` で組むと styling が効かず、class を足すと classless を選んだ理由が消える

## 3. Rejected

1. `hooks/` という dir 名 — React を使わないため、実装手段を誤って示唆する
2. `components/feature/` と `components/primitive/` の分割 — file 0 個の段階では早すぎる。数が増えて判別が困難になった時点で再検討する
3. Pico.css を CDN の `<link>` で読む — 外部 host への実行時依存。version 固定も lockfile の外へ出る
4. Pico.css の class 付き版 — classless を選んだ理由（UI design に時間をかけない）と衝突する
5. `<div id="root">` へ mount — classless Pico が semantic tag を見るため styling が効かない
6. component ごとに css を import — 同じ知識が複数 file に散る（DRY）
7. `view-models/` と `lib/` を必要になるまで作らない — role と dir の対応が虫食いになり、置き場の判断が毎回発生する
