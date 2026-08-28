---
name: playback list の視覚言語は夜の紫黒・Apple semantic 階層。行の見え方の参照は Apple Podcasts list
date: 2026-08-28T19:20:01
branch: docs/playback-list-page-design
---

## 1. Decision

1. **色の軸**: 面はほぼ黒の夜地。accent は紫（Apple system purple 慣用 `#bf5af2`）。global 地は純黒へわずかに紫を混ぜる
2. **semantic 色**: Apple 風に label / secondary / tertiary の階層を使う。紫は interactive（hover・keyboard focus の左 edge 等）に限り、常時の塗り全面には使わない
3. **質感の出し方**: specular・艶・細い反射・暗い面として出す。物語上の motif 名・物体禁止は `2026-08-28T19-20-00-docs-playback-list-page-design` を正とし、本 Decision は再掲しない
4. **参照の見え方**: list 行は Apple Podcasts の show episode list を手本にする。囲い card なし、item 間は gray hairline、title は meta より大きく、左に余白
5. **本 Decision の範囲**: 色・semantic 階層・参照からの見た目合意。layout 構成の詳細（grid 列・component 所有）と CSS 変数の数値百科は正本にしない。数値の正は CSS。concept / setting / motif の物語前提は `2026-08-28T19-20-00` を正とする

## 2. Reason

1. shim の指示は「紫・黒・Apple が推奨する色使い」であり、list 改修のたびに再発する。session 実装メモだと次の agent がカード囲い・全面紫 fill へ戻る
2. Apple Podcasts を参照にしたのは「行一覧の慣習」を借りるため。screenshot で確認した見え方（hairline・囲いなし・title/meta 差）を視覚合意に残す。公式 HIG color page は取得できないことがあったため、公開 system palette 慣用を採用した
3. Pico classless が article / button / main に card 余白・max-width を付けると、夜地の上に「謎の黒 gap」が出る。視覚言語として「囲いなし・full-bleed 地」を決めておかないと、framework 既定に引きずられる
4. motif の物体禁止や謎放送の物語を本 file に写すと `2026-08-28T19-20-00` と二重になり DRY を破る。空気の前提は参照し、ここは色と見え方だけを答える
5. 左紫 edge は interactive の印。常時紫塗りや sticky mouse focus の二重紫は、accent の意味（選択可能）を濁すので採らない

## 3. Rejected

1. elevated 黒 card で各 item を囲い、item 間に大きな黒 gap を空ける — Apple list 参照と「囲いは不要」に反する。Pico 既定の article card もこれに属する
2. hover / focus で行全体を紫 fill する — accent を全面に使うと semantic 階層が壊れ、夜の黒面が消える。紫は左 edge 等の線に留める
3. 白／クリーム地＋青 accent のパレットを list 既定にする — 夜地・紫 accent の軸と逆（物語側の Rejected は `2026-08-28T19-20-00`）
4. Decision 本文に padding px・font-size・全 CSS 変数を正本として列挙する — 実装が動けば即腐敗する。見た目の合意だけ残し、数値の正は CSS
5. motif 物体禁止や concept 本文を本 file に再掲する — 物語 Decision との DRY 違反。参照だけにする
