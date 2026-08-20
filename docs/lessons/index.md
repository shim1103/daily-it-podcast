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
