# lessons

- 2026-09-07 [fix/playback-e2e-test] 障害調査で「最有力の原因」を宣言する前に、その候補を確定 or 棄却できる一次観測（server log・実 request の応答 body）を先に取る。候補を確度順に並べる作業と、決定的な log を 1 本引く作業では後者を優先する。外形（test の失敗 locator）だけで原因を推論して確度を付けると、実 log が別方向を指したとき（例: session 生存を示す header が付いていた）に手戻りする  # → layer:workflow
- 2026-09-07 [fix/playback-e2e-test] 外部 API 認証の失効を疑うとき、失効の HTTP status を仕様で確認してから犯人を名指しする。token 失効は 401、request 内容不正は 400 のように意味が分かれる。status を見ずに「token 切れ」と断定すると、同じ status を返す別原因（設定値の不正など）と混同する。加えて、その API client の共通 request 経路が複数種類の呼び出しを 1 つの error message へ畳んでいると、status だけでは呼び出し種別を切り分けられない。畳み込み経路には呼び出し種別の label を必須引数で持たせ、log から一次切り分けできる粒度にする  # → layer:terms
- 2026-09-07 [fix/playback-e2e-test] flow 指示（例: 「success したら deploy → commit → …」）の実行中に分岐点へ来ても、選択肢が git 履歴で復元可能な範囲なら user へ質問せず自律判断で進める。質問して turn を止めてよいのは、判断材料が本当に不足し、かつ不可逆な副作用に関わるときだけ。flow の完遂責任を「確認を挟む」で薄めない  # → layer:workflow
