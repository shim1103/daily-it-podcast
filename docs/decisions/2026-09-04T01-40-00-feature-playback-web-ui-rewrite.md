---
name: EpisodeEntry を廃し EpisodeManuscript に選択展開の描画を吸収する。Entry は domain 役割名としては残すが専用 component を立てない
date: 2026-09-04T01:40:00
branch: feature/playback-web-ui-rewrite
---

## 1. Decision

1. `episode-entry.tsx`（`EpisodeEntry`）を廃止する。中身は `<section className="episode-entry"><EpisodeManuscript body onSeek /></section>` で、`body` / `onSeek` を透過するだけの 1 段ラッパだった。
2. 選択中 episode の原稿描画は page が `EpisodeManuscript` を直接配置する。「選択された行の直後に原稿を出す」という配置判断は page（統合点）が持つ（`selectedEpisode?.episodeId === row.episodeId` の分岐）。
3. `EpisodeManuscript` の root class は `episode-manuscript` のまま（既存 CSS がこの class に付く）。廃止する `episode-entry` class は未スタイルだったので CSS 影響なし。page / component test の `.episode-entry` セレクタは `.episode-manuscript` へ寄せる。
4. 先行 Decision `2026-09-02T15-00-00` §1-7 の「**Entry** は選択中 episode の manuscript のみ。`title` / `date` は Row が既に示すため重ねない」という **domain 上の役割記述**は維持する。役割としての Entry（＝選択展開部は原稿だけを持ち、Row と情報を重複させない）は `EpisodeManuscript` + page の配置で満たす。専用 component を立てないだけで、役割の制約は生きている。
5. 契約値（`EpisodeManuscript` の props・page の配置）の正本は A artifact。本 Decision は方針だけを固定する。

置き換え範囲: 先行 Decision `2026-09-02T15-00-00` §1-7 の「Row / Entry / AudioControls の domain 配置」のうち、**Entry を独立した Feature Component として実装する**という含意を本 Decision で置き換える。Entry は component 名ではなく「選択展開部は原稿のみ」という配置制約の名前になり、`EpisodeManuscript` が担う。維持範囲: 同 §1-7 の「Entry では `title` / `date` を重ねない」「AudioControls は再生対象のみ、selection と独立」、同 Decision の selection と playback の直交、`2026-09-04T01-10-00` の「page は分岐と配置のみ」は維持する。`EpisodeManuscript` → `EpisodeTopic` の構造（原稿を opening / topics / closing に組む）は変えない。

## 2. Reason

1. `EpisodeEntry` は `<section>` タグと `episode-entry` class を足す以外の処理を持たない。`body` も `onSeek` も加工せず `EpisodeManuscript` へ素通しする。`design-philosophy.md` §2-3（KISS）と §3-1（SRP）に照らすと、1 component は 1 つの実責務を持つべきで、「別 component をラップして class を 1 つ足す」だけの層は indirection のコスト（ファイル・test・import・JSDoc）に見合う価値がない。`feature-component.md` §2 の「複数の Primitive を組み合わせた機能画面」にも当たらない（組み合わせているのは `EpisodeManuscript` 1 つだけ）。

2. `episode-entry` class は CSS で一度も参照されていない（`rg 'episode-entry' web/src --glob '*.css'` が 0 件）。見た目の分離点として機能していない。将来 Entry に固有スタイル（枠線・余白）が要るなら、その時に `EpisodeManuscript` の外側 class か wrapper を足せばよく、今から空の class を持つ component を維持する理由がない（YAGNI、`design-philosophy.md` §2-4）。

3. 「選択された行の直後に原稿を出す」は配置の判断で、`2026-09-04T01-10-00` §1 で page の責務と確定済み（「`rows` を map して Feature Component を配置する」）。page が `{selectedEpisode?.episodeId === row.episodeId && <EpisodeManuscript … />}` と書けば、Entry component を経由せずに同じ配置になる。Entry を挟むと配置判断が page と Entry に半分ずつ乗る。

4. 先行 Decision §1-7 が Entry を立てたのは「選択展開部が Row と情報を重複させない（`title` / `date` を出さない）」という制約を名前で固定するため。その制約は `EpisodeManuscript` が `body`（opening / topics / closing）しか受け取らない時点で型で保証されている。`EpisodeManuscript` に `title` / `date` を渡す口がないので、役割名を component 名にしなくても制約は破れない。

5. 形を Decision 本文へ写すと A artifact と二重 SSOT になる（`decisions.md` §4-4）。本 Decision は「1 段ラッパ component を廃し、配置は page、原稿描画は既存 component」という方針だけを固定する。

## 3. Rejected

1. `EpisodeEntry` を残し、将来 Entry 固有の UI（閉じるボタン等）が付く余地として維持する — 現時点でその UI は要件になく（`docs/decisions` に記述なし）、`2026-09-02T15-00-00` §1-7 も「manuscript のみ」と明記している。閉じる操作は `2026-09-03T13-40-00` §1-1 が「blocking error に閉じる操作を提供しない」、selection の解除は Row の toggle が担うと決まっており、Entry に閉じるボタンが付く筋がない。必要になった Issue で component を立て直せばよく、空ラッパを保持する理由にならない（YAGNI）。

2. `EpisodeManuscript` を `EpisodeEntry` にリネームして 1 component に統合する（名前は Entry を採る）— `EpisodeManuscript` は原稿の組版（opening / topics / closing を `EpisodeTopic` で組む）が実責務で、その名前が内容を正しく表す。「Entry」は配置上の役割名で、原稿の組版とは別の抽象レベル。リネームすると「原稿を組む」という実責務が名前から消える。`EpisodeManuscript` の名前を保ち、Entry は Decision 上の役割記述に留める。

3. page ではなく `EpisodeRow` が選択時に `EpisodeManuscript` を内包する — Row の責務は「1 episode の meta と select / play affordance」（`2026-09-02T15-00-00` §1-7）。原稿の展開を Row に入れると Row が選択状態に応じて子ツリーを変える stateful な見た目になり、`feature-component.md` §3（状態に応じた分岐を Row 内に持たない、props で受けた状態を描画するだけ）から外れる。配置は page。

4. `episode-entry` class を `EpisodeManuscript` の root に付け替えて残す — 未スタイルの class を移動するだけで、CSS の参照は増えない。`.episode-manuscript` が既にスタイルの掛かる class で、原稿ブロックの見た目はそこで完結している。class を 2 つ持たせても片方は死んだまま。test のセレクタも `.episode-manuscript` に統一する。
