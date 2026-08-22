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
117：2026-08-22 [feature-generator-cursor-text-writer] mutationでtestが落ちたことを有効性の十分条件にしない。注入した差分が仕様変更（定数SSoTの値変更）ならそのtestは将来の正当な更新を止める側であり、bugを止める側ではない  # → layer:terms
118：2026-08-22 [feature-generator-cursor-text-writer] 定数がSSoTである層の値を、testが同じ定数参照で書き写しても値の誤りは検出できない。期待値が実装と同時に変わるため常に一致する。値の正しさは消費側の振る舞いで検証する  # → layer:terms
119：2026-08-22 [feature-generator-cursor-text-writer] mutationが生存したという判定も、compileが通った上での生存かを確認してから下す。未使用importなどでbuildが落ちた場合はmutationとして無効であり、検出力の証明にも反証にもならない  # → layer:terms
120：2026-08-22 [feature-generator-cursor-text-writer] 依存の位置情報は提供側が既定値まで解決し、呼び出し側へ組み立てを要求しない。組み立てを呼び出し側へ置くと解決規則がcallerの数だけ散り、同種の依存を増やすたびに規約が写経される  # → layer:terms
121：2026-08-22 [feature-generator-cursor-text-writer] 組み立て層が環境変数を読み始めたら、既存の同種factoryが環境を読んでいるか数える。1つだけ環境依存になっているなら、その知識は提供側が持つべき既定値である  # → layer:terms
122：2026-08-22 [feature-generator-cursor-text-writer] 秘密を扱う機構の「値を見てはいけない主体」に、agentだけでなく自分のprocessも含める。取得して渡す実装は、取得後の値がlog・error message・trace へ漏れる経路を作る  # → layer:terms
123：2026-08-22 [feature-generator-cursor-text-writer] 権限を絞る判断軸は「相手が漏らすか」ではなく「相手の正当な目的に必要か」。信頼できる相手にも不要な資格情報は渡さない。信頼は現時点の評価であり恒久保証ではない  # → layer:terms
124：2026-08-22 [feature-generator-cursor-text-writer] 外部CLIの機能有無は、library APIの範囲だけを見て断定しない。同じvendorがCLI subcommandとして別経路を持つことがあり、実binaryのhelpを読むまで「その機構は使えない」と結論しない  # → layer:platform
125：2026-08-22 [feature-generator-cursor-text-writer] 子processへ渡す環境変数は明示構築する。環境変数の集合を未設定のまま起動すると親の全環境を継承する言語runtimeがあり、空の集合を渡すこととは意味が異なる  # → layer:platform
126：2026-08-22 [feature-generator-cursor-text-writer] 査読者が複数いて所見が食い違う時、報告の説得力ではなく自分で再現した実測で採否を決める。検出力の主張は特に、対応するmutationを自分で打ってから転送する  # → layer:workflow
127：2026-08-22 [feature-generator-cursor-text-writer] 既存実装に倣えという指示は、対象を1件見て倣うのではなく同種を全数数えてから多数派を採る。1件だけを参照すると、その1件が例外だった場合に誤った前例を複製する  # → layer:workflow
128：2026-08-22 [feature-generator-cursor-text-writer] 完了指示の「finish」が成果物の作成完了か、その成果物が定める作業の実行完了かを、着手前に対象の状態を実測して切り分ける  # → layer:workflow
129：2026-08-22 [playback-ui-detail-design-boundary] 質問tool（AskUserQuestion相当）がpermission規約でdenyされた時、代替として自分でstdoutへ質問文を出力し、次のuser inputを待つ。tool不通=判断放棄の理由にしない  # → layer:workflow
130：2026-08-22 [playback-ui-detail-design-boundary] user inputが質問形式（例:「Plan」を含む）の時、質問への回答提示だけが許可範囲であり、editやsubagentへの委譲実行までは許可されない。「git履歴で復元できるから自律進行してよい」という一般規約は、明示的な非edit宣言や質問形式のinputには適用しない。両者が衝突したら、より個別具体的なsession内指示を優先する  # → layer:workflow
131：2026-08-22 [playback-ui-detail-design-boundary] 越権実行に気づいたら、正当化や取り繕いをせず、何を逸脱したか・なぜ逸脱したかを最初の一文で明示してから状況を提示し、次の指示を待つ  # → layer:workflow
132：2026-08-22 [playback-ui-detail-design-boundary] test runnerのdefault excludeが対象fileの命名patternと衝突すると、testは0件収集のまま無言でsuccess終了し、実行結果のtest数だけでは検出できない。新規test fileを追加したら、runnerのlist/verboseで対象fileが実際に収集されたことを個別に確認する  # → layer:platform
133：2026-08-22 [playback-ui-detail-design-boundary] 本番route handlerと同型のrouting/switchロジックをdev-only tooling側に複製すると、命名・配置は正しくてもDRY違反が残る。複製に気づいたら、production側のexport可能な関数へ委譲できないか先に検討し、不可能な場合のみ複製を許容してdocsへ理由を明記する  # → layer:terms
134：2026-08-22 [playback-ui-detail-design-boundary] workflow skillのflowが「/skill-name」を短く言及するだけの時、skill fileを実際にreadせず文字列から実行方法を推測しない。読まずに実行すると、Investigation・template・完了条件をすべて飛ばして表面だけ真似ることになる  # → layer:workflow
129：2026-08-22 [chore/generator-ci-test-configuration-hardening] scope 分割の B は実装後の doc 同期だけで完了しない。同じ問いへ再利用する判断は logging の Decision Record として保存する  # → layer:workflow
130：2026-08-22 [chore/generator-ci-test-configuration-hardening] coverage、mutation、fuzzing は検出する品質特性と実行 cost が異なる。1つの gate へ統合せず、各 metric の能力境界と実行場所を個別に決める  # → layer:terms
131：2026-08-22 [test/generator-time-determinism] Unit Test の入力が現在時刻に依存しない仕様では、固定した時刻を明示して渡す。実行時の外部状態を入力にすると同じ契約を毎回同じ条件で検証できない  # → layer:terms
