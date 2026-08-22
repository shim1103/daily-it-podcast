# lessons

108：2026-08-20 [generator-google-oauth-adapter] `non-edit` の manager は実装を直接変更せず、executor へ委譲して reviewer の指摘を再実行で反映する。権限境界を守ることは完了処理を止めることではない  # → layer:workflow
109：2026-08-20 [generator-google-oauth-adapter] workflow が起点branchを指定している時、既存branchが同じcommitを指す状況証拠で代用せず、指定されたbranch作成手順をそのまま実行する  # → layer:workflow
110：2026-08-20 [playback-runtime-config-http-contract] workflowの仮branch名はtaskのpropertyから機械的に導出し、作成直後に現在branch名を検査する。placeholderを正式branchとして残さない  # → layer:workflow
111：2026-08-20 [playback-runtime-config-http-contract] ApiErrorはUI表示文ではなく操作判断用の中間語彙として保つ。server内部障害と外部service一時不能は、UI表示が同じでもretry意味が違うためcodeを統合しない  # → layer:terms
112：2026-08-20 [issue/playback-web-api-client] Sociable Unit Testは協調先の判定結果を再assertせず、必要な協調先をdoubleにして対象unitの入力伝播・戻り値伝播だけを検証する  # → layer:terms
113：2026-08-20 [issue/playback-web-api-client] 同じ公開failureへ収束する内部経路は重複testにせず、異なるobservableなfailureや後続処理の失敗だけを別caseとして所有する  # → layer:terms
108：2026-08-20 [generator-episode-validation] coverage対象外のComposition Rootでも、実装選択を固定する結線testは省略しない。coverage除外は計測範囲の設定であり、責務境界の検証免除ではない  # → layer:terms
109：2026-08-20 [generator-episode-validation] Driven Portを実装するAdapterと、そのPortを呼ぶUseCaseを同じ型へ兼務させない。Composition RootはUseCaseへAdapterを注入し、呼び出し側がvalidationを迂回できない戻り値を公開する  # → layer:terms
110：2026-08-20 [generator-google-oauth-adapter] `non-edit` の manager は実装を直接変更せず、executor へ委譲して reviewer の指摘を再実行で反映する。権限境界を守ることは完了処理を止めることではない  # → layer:workflow
111：2026-08-20 [generator-google-oauth-adapter] workflow が起点branchを指定している時、既存branchが同じcommitを指す状況証拠で代用せず、指定されたbranch作成手順をそのまま実行する  # → layer:workflow
112：2026-08-20 [playback-runtime-config-http-contract] workflowの仮branch名はtaskのpropertyから機械的に導出し、作成直後に現在branch名を検査する。placeholderを正式branchとして残さない  # → layer:workflow
113：2026-08-20 [playback-runtime-config-http-contract] ApiErrorはUI表示文ではなく操作判断用の中間語彙として保つ。server内部障害と外部service一時不能は、UI表示が同じでもretry意味が違うためcodeを統合しない  # → layer:terms
114：2026-08-20 [docs-status-audit] goal に workflow 名（例: finished /scope-split）が含まれても、直前 user input が説明要求なら workflow 実行ではなく分類・説明のみ。workflow 名を goal と同一視しない  # → layer:workflow
115：2026-08-20 [docs-status-audit] `non-edit` が二重付きで goal が分類のみの時、A 候補の列挙は許可されても実行はしない。後段の「やっていい」は分類基準の説明であり実行許可と読み替えない  # → layer:workflow
116：2026-08-20 [docs-status-audit] scope に `non-scope: C` がある時、C に属する Issue draft・worktree・branch を先取りしない。誤って作った場合は revert してから A/B だけを履歴へ残す  # → layer:workflow
117：2026-08-22 [playback-ui-detail-design-boundary] 質問tool（AskUserQuestion相当）がpermission規約でdenyされた時、代替として自分でstdoutへ質問文を出力し、次のuser inputを待つ。tool不通=判断放棄の理由にしない  # → layer:workflow
118：2026-08-22 [playback-ui-detail-design-boundary] user inputが質問形式（例:「Plan」を含む）の時、質問への回答提示だけが許可範囲であり、editやsubagentへの委譲実行までは許可されない。「git履歴で復元できるから自律進行してよい」という一般規約は、明示的な非edit宣言や質問形式のinputには適用しない。両者が衝突したら、より個別具体的なsession内指示を優先する  # → layer:workflow
119：2026-08-22 [playback-ui-detail-design-boundary] 越権実行に気づいたら、正当化や取り繕いをせず、何を逸脱したか・なぜ逸脱したかを最初の一文で明示してから状況を提示し、次の指示を待つ  # → layer:workflow
120：2026-08-22 [playback-ui-detail-design-boundary] test runnerのdefault excludeが対象fileの命名patternと衝突すると、testは0件収集のまま無言でsuccess終了し、実行結果のtest数だけでは検出できない。新規test fileを追加したら、runnerのlist/verboseで対象fileが実際に収集されたことを個別に確認する  # → layer:platform
121：2026-08-22 [playback-ui-detail-design-boundary] 本番route handlerと同型のrouting/switchロジックをdev-only tooling側に複製すると、命名・配置は正しくてもDRY違反が残る。複製に気づいたら、production側のexport可能な関数へ委譲できないか先に検討し、不可能な場合のみ複製を許容してdocsへ理由を明記する  # → layer:terms
