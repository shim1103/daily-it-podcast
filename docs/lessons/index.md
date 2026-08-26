# lessons

177：2026-08-26 [docs/architecture-reconsider-react-hono] 質問tool・plan mode切替toolがpermission規約でdenyされた時、機能停止と扱わず、同じ目的（判断の提示・承認確認）をその場のtext出力とuserの次inputで代替する。tool不通は判断や進行を止める理由にしない  # → layer:workflow
178：2026-08-26 [docs/architecture-reconsider-react-hono] workflow skillの説明文にある動詞（show等）を、file書き込みを伴う動詞（create・write・fix）と同一視しない。skill本文の分類が「提示」を指す時は、承認前にartifactを実体化しない  # → layer:workflow
179：2026-08-26 [docs/architecture-reconsider-react-hono] userの短い指示語（remember等）を字面通りの機能呼び出しと解釈する前に、直前の文脈（何を保存すべき情報として提示したか）と照らして意図を確認する。誤った解釈で副作用のある処理を実行すると、後から全体を取り消す作業が発生する  # → layer:workflow
180：2026-08-26 [docs/architecture-reconsider-react-hono] 委譲先が「成功」と報告した内容でも、実行に使ったtoolchain・package manager・生成物が対象repoの既存の正本（lockfile形式、config file）と一致しているかを個別に確認する。報告の成功可否と、環境への副作用の正しさは別の検証項目である  # → layer:workflow
181：2026-08-26 [docs/architecture-reconsider-react-hono] 言語・toolchainの設定directiveが期待通り機能するかは、設定を追加した直後に実際のtoolを実行して確認する。ドキュメントの記憶や類推だけでは、同義とみなした値の組み合わせが実行系によって自動的に無効化される仕様を見落とす  # → layer:terms
182：2026-08-26 [docs/architecture-reconsider-react-hono] 特定の値をハードコードで検証するtestがある状態でその値の重複除去（DRY化）を行うと、実装は正しくなってもtestの断定的assertionが失敗する。値の同期を求めるtestを見つけたら、実装のDRY化とtestの検証方法（直書き値の比較→参照先の存在確認）を同じ単位で直す  # → layer:terms
183：2026-08-26 [docs/architecture-reconsider-react-hono] git操作（stage解除・reset等）がpolicyでblockされた時、blockされた操作を別コマンドで回避しようとせず、現在stageされている内容をそのまま活かせる単位でcommit粒度を再設計する  # → layer:platform
177：2026-08-26 [feature/generator-agentsecrets-http-proxy-absorb] 同じ置き場名でも出口契約が違うなら runtime package を出口ごとに分ける。見た目の重複を DRY 違反とみなし1袋へ畳むと Orthogonality が壊れる  # → layer:terms
178：2026-08-26 [feature/generator-agentsecrets-http-proxy-absorb] path の移動は契約吸収ではない。所有境界の固定と、出口契約を満たす単一実装への再設計を同一の完了条件にしない  # → layer:terms
179：2026-08-26 [feature/generator-agentsecrets-http-proxy-absorb] default 接続先の test は dial 失敗を証拠にしない。差し込んだ transport が受け取った URL を assert し、port 占有や到達可否に依存しない  # → layer:terms
180：2026-08-26 [feature/generator-agentsecrets-http-proxy-absorb] 削除 tool の承認 UI の表示対象と実際の削除 path を混同しない。承認前に対象 path を確認し、無関係な lockfile と同一視しない  # → layer:workflow
181：2026-08-26 [docs/migrate-lessons-2] goalに完了条件（例：特定fileが空であること）が明記されている時、同時に書かれたrole宣言（non-edit等）はreview・planへの権限制限であり、goal自体の完遂を差し止める指示ではない。role宣言を理由にgoalの完了条件を満たさないまま止めない  # → layer:workflow
182：2026-08-26 [docs/migrate-lessons-2] 複数runtimeへ同一知識をsymlink展開する構成で、SSOT側に実体が存在しないことだけを見て新設と判断しない。展開先の各runtimeを横断的に確認し、どこか1つにのみ実体が存在する非対称配置（linkは追加されたが実体の一部runtimeへの反映が漏れていた等）を見落とすと、既存知識を上書きして消失させる  # → layer:workflow
183：2026-08-26 [feature/generator-cmd-entrypoint] 公開境界の契約 documentation は entrypoint と置換可能な公開 symbol に限る。unexported helper や test からだけ呼ばれる symbol へ契約 tag を置くと公開境界との二重定義になり DRY を破る  # → layer:terms
184：2026-08-26 [feature/generator-cmd-entrypoint] 完了済み仕様を未完了 index に残さない。checklist を完了にしたら待ち一覧からも同時に外し、完了記録と未完了 index の責務を混ぜない  # → layer:workflow
185：2026-08-26 [feature/generator-cmd-entrypoint] 設計・実装の議論軌跡（なぜそうした／しなかったか）で code comment に残さなかった判断は Decision の Reason / Rejected へ残す。局所の Why not だけ code に置き、再発する問いへの答えは decisions が持つ  # → layer:meta
186：2026-08-26 [feature/generator-cmd-entrypoint] CLI の成否観測は runtime 内部状態ではなく OS process の exit code と stderr を正とする。同一 binary の写像を local と CI で分けない  # → layer:platform
187：2026-08-26 [feature/generator-cmd-entrypoint] 査読必須指摘は委譲先報告をそのまま採用せず、規約原文と現物を委譲元が照合してから修正対象へ格上げする。borderline 見送りも明文違反なら格上げする  # → layer:workflow
188：2026-08-26 [feature/generator-agentsecrets-cursor-command-launcher] Composition 内で「表」を名乗る file に runtime factory を同居させると、変更理由が2つ同居して Least Astonishment を破る。対応表と契約実装の組み立ては file を分ける  # → layer:terms
189：2026-08-26 [feature/generator-agentsecrets-cursor-command-launcher] 結線層の test で具象型 assert により「実装が間違っていない」ことを見るのは振る舞い検証ではない。分岐の無い結線に構造 guard を増やさない  # → layer:terms
190：2026-08-26 [feature/generator-agentsecrets-cursor-command-launcher] external test package と unexported helper の white-box test は同一 file に結合できない。結合するなら公開経路の観測へ吸収する  # → layer:platform
191：2026-08-26 [feature/generator-agentsecrets-cursor-command-launcher] Composition の vendor 結線 file と bindings/runtime の責務を片方だけ直すと、次の問いがすぐ再発する。表・runtime・each の3役割を同時に揃える  # → layer:terms
192：2026-08-26 [feature/generator-agentsecrets-cursor-command-launcher] 自動 review が必須なしと見送っても、coding-style に明文で違反していれば確認待ちせず修正対象へ格上げする  # → layer:workflow
193：2026-08-26 [feature/playback-web-primitive-component-jsx] 下位層だけ先に宣言的 UI へ移し上位が命令的 DOM 組み立てのままだと、削除契約と Verification 緑が衝突する。上位の本格移行 Issue を侵さず機械的追従だけ許すなら、寿命付きの橋を上位側に置き恒久 abstraction にしない  # → layer:terms
194：2026-08-26 [feature/playback-web-primitive-component-jsx] createRoot で描画した子だけを別 DOM へ移して unmount しないと orphan root が残る。unmount すると管理下の子が壊れる。root を持たない静的 markup 経路を選ぶ  # → layer:platform
195：2026-08-26 [feature/playback-web-primitive-component-jsx] 静的 markup では ref が走らない。動的 dataset key を commit 後 mutation に頼らず、browser dataset と同値の data-* へ写像して declarative に渡す  # → layer:platform
196：2026-08-26 [feature/playback-web-primitive-component-jsx] precondition 違反の test は、その検査自体が公開 postcondition のときだけ足す。throw しない pass-through 部品に異常系を足す根拠は契約に無い。空文字や複数 hump などの境界は別物として最小化して足せる  # → layer:terms
197：2026-08-26 [feature/playback-web-primitive-component-jsx] 査読の must-fix は報告を転送する前に、委譲元が現物と再現条件を自分で確認してから差し戻す。見送り指摘も設定一貫性など明文の欠落なら格上げする  # → layer:workflow
193：2026-08-26 [feature/playback-worker-hono-route-definition] skillのroleが「non-edit」と定義されている時、flowの各stepに実行主体（誰がAgent toolで委譲するか）が明記されていないと、managerが自らfileをedit・test実行してしまう。role宣言だけでなくflowの各行に主体を明記しないと、非edit原則は実行時に守られない  # → layer:workflow
194：2026-08-26 [feature/playback-worker-hono-route-definition] 不可逆性を判断する時、git管理下でcommit前のfile変更は`git diff`/`git checkout`で復元できる可逆操作であり、質問toolで実行を止める理由にならない。既存hookが「git履歴で復元できる範囲は自律判断で進めてよい」と明示している時はそれに従う  # → layer:workflow
195：2026-08-26 [feature/playback-worker-hono-entry-cutover] 削除対象を参照するfile一覧は、Issue本文の記載を正本にせず、削除実行前に自分でgrepして確定する。Issueの依存記述は作成時点のsnapshotであり、後続sessionで追加されたfile（test含む）を捕捉できていないことがある  # → layer:workflow
196：2026-08-26 [feature/playback-worker-hono-entry-cutover] managerが事前調査で見つけたIssue非記載の追加依存は、委譲先への指示に「発見済みの事実」として明記して渡す。委譲先が独自に発見し直す前提に置くと、同じ見落としが再発するか二重調査が発生する  # → layer:workflow
