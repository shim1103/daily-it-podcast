---
name: playback list の interactive 色は「選択＝紫、hover＝黄」の 2 色制。左 edge 線は非選択でも常時（灰）出す
date: 2026-09-04T16:57:15
branch: refactor/playback-web-design-fix
---

## 1. Decision

1. **左 edge 線は常時出す**: `.episode-item` の左 edge（`box-shadow: inset 3px`）は、非選択・非 hover でも灰の線（`--el-edge-idle`）を出す。「item 間の gray hairline で区切る」に加え、各 item の左端にも弱い線を置いて行の始まりを示す。
2. **interactive の色は 2 色に分ける**: hover / keyboard focus は**黄**（`--el-hover`、Apple system yellow 系 `#ffd60a`）、選択中は**紫**（`--el-accent`）。左 edge 線の色を「非選択非 hover＝灰／hover＝黄／選択中＝紫」で切り替える。選択中は hover より優先（`[data-selected="true"]` と `[data-selected="true"]:hover` が同じ紫）。
3. **hover は item 全体を持ち上げる**: hover 中は左 edge 線を黄にするだけでなく、item 全体の面を黄でごく薄く（`color-mix(... --el-hover 10%)`）持ち上げる。行内（`.episode-row`）は hover 背景を持たず、hover の面強調は親 `.episode-item` に一本化する。keyboard focus（`:focus-within`）だけ行内で薄く示す。
4. **再生ボタンの丸は黄を持たない**: 右端の再生 / 停止ボタンの丸い背景は「非再生＝薄灰／再生中（`data-active`）＝白」。hover は拡大（`transform: scale`）のみで背景色を変えない。item の hover 黄とは独立させ、ボタン自体には interactive 色を乗せない。
5. `--el-edge-idle` / `--el-hover` の値の正は CSS（`episode-list.css` の token 定義）。本 Decision は色の役割（どの状態にどの色を割り当てるか）だけを固定し、hex 値・px を写さない。

置き換え範囲: 先行 Decision `2026-08-28T19-20-01-docs-playback-list-page-design.md` を次のとおり部分的に置き換える。§1-2「紫は interactive（hover・keyboard focus の左 edge 等）に限り、常時の塗り全面には使わない」のうち、**hover を紫が担う**という含意を本 Decision §1-2 で置き換える。hover は黄、選択中だけが紫。§Rejected-2「hover / focus で行全体を紫 fill する」は「紫で」行全体を塗ることの拒否として維持し、本 Decision §1-3 の「黄でごく薄く面を持ち上げる」はこれに当たらない（accent の意味＝選択可能を濁さないため、面の持ち上げは選択色の紫とは別系統の黄で、かつ薄く）。§1-4「item 間は gray hairline」は維持し、本 Decision §1-1 で各 item の左 edge にも灰線を足す（hairline は item 間の区切り、左 edge 線は行頭の印で役割が別）。維持範囲: 先行 §1-1（夜地・紫 accent の軸）、§1-3（specular・艶で質感を出す）、§1-4 の「囲い card なし・title は meta より大きく・左に余白」、`2026-08-28T19-20-00` の物語前提（concept / setting / motif）は維持する。

## 2. Reason

1. shim が list 改修の見た目指示で明示したのは「非選択の左 edge は紫の代わりに灰の線」「hover で黄」「item 全体を hover で黄」。`design-philosophy.md` §5 末尾「規約 doc と現在の task 指示が食い違う時は現在の指示を優先し、規約 doc は後続で整合させる」に従い、指示を採って先行 Decision の色軸を追補する。session メモに留めると次の agent が「hover も紫」に戻す。
2. 左 edge 線を非選択でも出すのは、item 間の hairline だけだと「どこからが 1 行か」の視覚的な始まりが弱いため。灰で弱く常時出し、選択・hover でその線に色が乗る形にすると、同じ 1 本の線が状態を伝える 1 経路になる（`Least Astonishment` §4-5、状態表示の一貫）。
3. hover と選択を同じ紫にすると、マウスを乗せた行と実際に選択中の行が同色になり「今どれが選ばれているか」が hover 追従で濁る。先行 Decision §Rejected-5 が二重紫を拒否したのと同じ理由で、hover を別色（黄）に分けると「黄＝今カーソルがある」「紫＝選択済み」が一目で分かれる。
4. hover の面の持ち上げを紫でやると先行 §Rejected-2 の「行全体を紫 fill」に触れ、accent（＝選択可能）の意味が薄まる。黄で、かつ 10% の薄さに留めれば、夜地の黒面を残しつつ「乗っている」感触だけ足せる。行内に hover 背景を持たせず親に寄せるのは、hover 対象（item 全体）と描画責務（item）を一致させ、Row を「props で受けた状態を描くだけ」に保つため。
5. 再生ボタンの丸に hover 黄を乗せると、item hover の黄面 + ボタンの黄丸が二重の黄になり、「押せるのはどっちか」が曖昧になる。ボタンは拡大だけで「押せる」を示し、色は再生状態（薄灰／白）専用にすると、item の interactive 色（灰・黄・紫）とボタンの状態色（薄灰・白）が別レイヤーで読める。
6. 形（hex・px・token 一覧）を Decision へ写すと CSS と二重 SSOT になる（`design-philosophy.md` §2-2、`decisions.md` §4-4）。色の役割割り当てだけを残す。

## 3. Rejected

1. 非選択の左 edge を透明のままにし、選択・hover 時だけ線を出す（先行実装の形） — shim の「灰の線を入れて」に反する。行頭の印が状態変化でしか出ないと、静止状態の一覧で行の境界が hairline 頼みになる。
2. hover も紫のまま（先行 Decision §1-2 の素直な読み）にし、選択とは濃さ・線幅で差をつける — 同系色の濃淡は環境（輝度・視野角）で潰れやすく、`2026-08-28T19-20-01` §2-2 が screenshot で確認した「hairline・title/meta 差」のような明快さが出ない。色相を分ける方が確実。
3. hover の面を紫の薄い持ち上げにする — 先行 §Rejected-2 に真っ向から当たる。accent を面に使うと semantic 階層が濁る。黄は accent ではない別系統なので面へ薄く使える。
4. 再生ボタンの丸も hover で黄にして item の hover 黄と揃える — 「押せる要素」が二重に黄で光り、主対象（行 select か、ボタン再生か）が分からなくなる。ボタンは形（拡大）で affordance を示し、色は状態専用にする。
5. `--el-hover`（黄）を accent token（`--el-accent`）の別名にする — 黄と紫は役割が違う（hover と選択）。別 token にしないと「hover 色を変えたら選択色も変わる」結合が生まれる（Orthogonality §2-1）。
