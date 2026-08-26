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
135：2026-08-22 [chore/generator-condition-coverage-report] 動的な文字列を正規表現や置換式へ渡す時は、shell quote だけで安全と判断しない。対象文字列を literal として扱う API を選び、特殊文字を含む値で変換結果を検証する  # → layer:platform
136：2026-08-22 [feature-generator-infras-all-narrow-integration] 後続Issueが守るinterface・定数・dir・責務境界は、実装Issueへ残さず先にcontract artifactとして固定する。実装Issueはそのartifactを参照し、新しい境界判断を持ち込まない  # → layer:workflow
137：2026-08-22 [feature-generator-infras-all-narrow-integration] 未決定・未実測の内容をDecisionや実装Issueへ書くと、仮定が契約として固定される。決定済みの事実と不足している決定を分け、不足側は再分類条件とともにlaneへだけ残す  # → layer:workflow
138：2026-08-22 [feature-generator-infras-all-narrow-integration] runtime名が共通でも、外部境界・contract・検証方法が異なる実装は別Issueに分ける。共有するruntime名ではなく、独立して完了・検証できる境界をIssue単位にする  # → layer:workflow
139：2026-08-22 [test/generator-mutation-testing] Decision Record が tool名だけを固定し module path を書いていない時、名前の類似から推測した path で `go install` しない。web検索で実配布元（作者・repository）を確認してから固定 version を install する  # → layer:workflow
140：2026-08-22 [test/generator-mutation-testing] reviewer の懸念が「ツールの内部挙動」に関するものである時、挙動を再現するテストケースを作れなくても、tool の source（module cache 配下）を直接読んで反証・確証できる。憶測で採否を決めない  # → layer:workflow
141：2026-08-25 [feature/playback-ui-structure] goalの「finished /skill-name」は最終到達点の宣言であり、そのturnで実行すべき作業内容ではない。同時に書かれたflowが「調査してまとめよ」の性質なら、goalの語だけでskillの実行フェーズ（decision確定・file作成）まで進めない。goalとflowが揃っている時はflowの動詞で着手範囲を決める  # → layer:workflow
142：2026-08-25 [feature/playback-ui-structure] `non-edit`が明示されたturnでも、後続に別skill（scope-split等）への参照があると「そのskillを今回実行してよい」と読み違えやすい。non-edit宣言はskill参照の有無に関わらず最優先の制約として維持し、file書き込みを伴う工程はskillの名前だけで判定せず、宣言と矛盾しないかを個別に確認してから着手する  # → layer:workflow
143：2026-08-25 [feature/playback-ui-structure] forkへ複数論点（一般調査＋UI具体案のように独立した2つ）を1つのpromptへ混在させると、forkが片方だけを実行して完了報告することがある。異なる調査対象を1つのfork呼び出しに詰め込まず、依頼した論点が全て報告に含まれているかを受信時に照合する。欠けていれば同じagentへ再送し、新規に別forkを立てない  # → layer:workflow
144：2026-08-25 [feature/playback-ui-structure] 「executorに委譲するんだけどね」という念押しは、直前の作業配分が単純作業を自分で抱え込んでいたことへの訂正。TDDのRed-Green-Refactorのうち、設計判断を要するRedまでを自分で書き、分量のあるGreen実装・配線をexecutorへ渡す境界を、指摘を受ける前に自分で引く  # → layer:workflow
145：2026-08-25 [feature/playback-ui-structure] 複数fileへ同型の追記（層境界の定義など）をする時、file間の相互参照（他fileの節番号を名指しする文言）を貼った直後に、参照先の実際の節番号と一致するか照合する。追記した節番号を先に決め打ちすると、他fileの追記位置とずれて誤参照が残る  # → layer:terms
146：2026-08-25 [feature/playback-ui-structure] 委譲先の完了報告で追加された実装判断（指示に無かった分岐の実装等）は、報告文の主張をそのまま信じず、対応するsource（型定義・契約関数の挙動）を自分で読んで、既存契約の範囲内かを検証してから承認する  # → layer:workflow
147：2026-08-25 [feature/playback-ui-structure] DRY指摘を受けた時、「同じコード片の重複」と「同じ値の画面上の重複表示」は別種のDRY違反であり、対処法も異なる（前者は共通関数への抽出、後者はcomponentの契約自体から重複箇所を除去）。指摘の対象がどちらかを、既存実装のどの層（構文 or 表示結果）を指しているか本人に確認せず先走って直さない場合は、両方の可能性を実装を見てから判定する  # → layer:terms
148：2026-08-25 [feature/playback-ui-structure] pre-commit hookがrepo全体の静的検査・testを実行する構成では、system標準の一時dir書き込みがsandbox制限で失敗することがある。commitがtmpfile権限エラーで止まったら、変更内容の欠陥ではなくsandbox実行環境の制約を先に疑い、sandbox解除で原因を切り分けてから対応する  # → layer:platform
149：2026-08-25 [feature-generator-processenv-command-launcher] 起動や注入の入出力を定義する契約 package は結線ではない。結線オンリーなのは Composition Root であり、契約と組み立てを同名にすると次の読み手が契約を移動・削除しやすい  # → layer:terms
150：2026-08-25 [feature-generator-processenv-command-launcher] 同じ外側 ring 内で、変わり方の軸が違う境界を抽象へ依存させるのは DIP であり Dependency Rule 違反ではない。Application Port には UseCase の語彙だけを上げ、手段語彙の契約を Application へ持ち上げない  # → layer:terms
151：2026-08-25 [feature-generator-processenv-command-launcher] ある runtime path の YAGNI で Composition 所有の設定を削る時、別 runtime が同じ設定を必要とするなら follow-up task を残す。削ることと思想の撤回を同一視しない  # → layer:workflow
152：2026-08-25 [feature-generator-processenv-command-launcher] 秘密の置き場と外部出口は独立軸である。同じ置き場名でも出口の契約・失敗モード・検証境界が違うなら Issue を分け、runtime 名だけで束ねない  # → layer:workflow
153：2026-08-25 [feature-generator-processenv-command-launcher] 同じ出口でも local と remote で Least Privilege の手段が異なりうる。片方の path の実装を外しても、他方 path の決定を黙って無効化しない  # → layer:terms
154：2026-08-25 [feature-generator-processenv-command-launcher] hook が変更範囲外の系統を動かし、その系統の依存が未導入で起動に失敗した時は実行環境起因である。product code へ回避を書き込まず、依存導入側で通す  # → layer:platform
155：2026-08-25 [feature/playback-worker-deploy] Decision Record の厚みは値の再掲ではなく Reason と Rejected の対である。正本がある契約値を本文へ写して厚く見せると、更新が二重化し「薄い」指摘の本質を外す  # → layer:workflow
156：2026-08-25 [feature/playback-worker-deploy] Decision は 1 判断 1 ファイルにする。文書分業とセキュリティ方針のように問いが独立なら分け、束ねた Decision 箇条は Reason が追いつかない  # → layer:workflow
157：2026-08-25 [feature/playback-worker-deploy] 境界契約 artifact を正とする分類では、決定記録と運用文書は契約字段を参照するだけで写さない。写した瞬間に契約の正本が複数になる  # → layer:workflow
158：2026-08-25 [feature/playback-worker-deploy] datetime 付きの記録 dir は latest 運用の入口にしない。同じ問いに日付なしで答える運用方針は、地図・層規則と並べる恒久文書へ置く  # → layer:workflow
159：2026-08-25 [feature/playback-worker-deploy] logging の decision は「どう書くか」、scope-split は「何をどの SSOT に固定するか」の 4 分けである。片方を直しても他方が曖昧なままだと、薄い Decision と契約複製が再発する  # → layer:workflow
160：2026-08-25 [feature/playback-worker-deploy] user が lane file 名を誤って指定しても、作業の所属 domain を優先して正しい lane を更新する。誤った系統の lane を汚さない  # → layer:workflow
161：2026-08-25 [feature-generator-processenv-http-transport] 小さい重複や不整合を見つけた時、修正コストがKISSの範囲に収まるなら follow-up 候補として先送りせず、その場で直す。先送りは「今すぐ直せない規模の変更」のためにあり、数行の共通化にまで適用しない  # → layer:terms
162：2026-08-25 [feature-generator-processenv-http-transport] DIPの充足はconstructorで抽象をDIしていることだけでは判定できない。構造体がinterfaceを持っていても、内部methodが具体的なruntime関数を直書きしていればその箇所はDIされていない。DI済みかは「その値がconstructor経由で差し替え可能か」を実装内部まで辿って確認する  # → layer:terms
163：2026-08-25 [feature-generator-processenv-http-transport] 同じ役割を持つ複数の実装のうちどちらが規範か判断する時、「後から書かれた方」「直近で自分が触った方」を規範だと推測せず、両方の実装を実際に読んで契約充足度を比較してから決める。時系列や記憶の新しさは正しさの根拠にならない  # → layer:workflow
164：2026-08-25 [feature-generator-processenv-http-transport] 自動review（複数観点の並列agent）が"borderline"として見送った指摘は、レビュー対象の規約docと直接照合し、規約に明文で違反していれば人間の確認を待たず修正対象へ格上げする。自動reviewの見送り判定をそのまま最終判断として扱わない  # → layer:workflow
165：2026-08-25 [feature-generator-cmd-usecase-boundary] Port の戻り型は外部能力の形に留め、業務上の構造化・解釈は Application が行い Domain Error で失敗を表す。能力の外側に業務型を載せるのは Port ではない  # → layer:terms
166：2026-08-25 [feature-generator-cmd-usecase-boundary] 完成成果物の永続では、構築方針を持つ Builder と検査してから書く Gate を分ける。名前が「書く」でも構築・呼び出し順・集約を Gate へ移さない  # → layer:terms
167：2026-08-25 [feature-generator-cmd-usecase-boundary] 対称性や層の見た目のために、方針のない空洞 UseCase を増やさない。方針の所有者を空洞化して下位へ移す分割もしない  # → layer:terms
168：2026-08-25 [feature-generator-cmd-usecase-boundary] 機械的な分解・結合 helper は完成契約の公開 Entities API に押し上げず、方針 UseCase の非公開 helper に置いてよい  # → layer:terms
169：2026-08-25 [feature-generator-cmd-usecase-boundary] Composition Root の factory は Port でも UseCase でも返してよい。結線単位は製品入口が必要とする境界に合わせ、Port だけに揃えない  # → layer:terms
170：2026-08-25 [feature-generator-cmd-usecase-boundary] coverage gate で結線専用の入口 package が未実行のまま閾値を割る時は、製品ロジックを歪めず除外対象へ寄せてよい。除外した事実は coverage SSOT に残す  # → layer:platform
171：2026-08-25 [feature-generator-cmd-usecase-boundary] panic だけの stub でも statement を実行する観測 test（recover）を置けば、未実装のまま coverage と契約存在を両立できる  # → layer:platform
172：2026-08-25 [feature-generator-cmd-usecase-boundary] scope-split の C は local Issue file で足りる。remote Issue 作成は別指示がない限り自動でやらない  # → layer:workflow
173：2026-08-25 [feature-generator-cmd-usecase-boundary] Decision の戻り型や層置きが後から誤りと分かったら、本文を黙って書き換えず superseded を明示して新 Decision を正にする  # → layer:workflow
174：2026-08-25 [feature-generator-cmd-usecase-boundary] lane の checkbox 順は実装メモであり最終境界の正ではない。正は philosophy / architecture と確定した Decision である  # → layer:workflow
165：2026-08-25 [feature-generator-agentsecrets-http-transport] Acceptance Criteria 通過のための一時 wrap を最終形とみなさない。契約実装の正本がどこか、という再発判断は Decision に残し、到達手段の完了と最終配置の完了を混同しない  # → layer:workflow
166：2026-08-25 [feature-generator-agentsecrets-http-transport] Issue が既存実装を Canonical Source に挙げても wrap 義務ではない。正本の単一化が必要な時は設計原則（DRY / DIP）で最終形を決め、手段を契約へ昇格させない  # → layer:workflow
167：2026-08-25 [feature-generator-agentsecrets-http-transport] 「移管」を path の移動・改名と同一視しない。所有境界・契約充足・正本の単一化を設計し直すのが本体であり、path 変更はその結果にすぎない  # → layer:terms
168：2026-08-25 [feature-generator-agentsecrets-http-transport] manager が Issue 本文に無い実装選択を freeze した時、その選択が恒久形かどうかを同じ session で判定し、恒久でないなら Rejected と対になる最終形判断を残す。AC 通過だけで閉じない  # → layer:workflow
169：2026-08-25 [chore/playback-worker-web-layer] scope-split の A は境界契約を code（dir・設定・型）として固定する工程であり、契約を説明した markdown を別途作ることではない。B は A の code を正として Decision / 運用方針文書を更新する  # → layer:workflow
170：2026-08-25 [chore/playback-worker-web-layer] allowlist の Write 許可は Delete を含まない。別名の tool は個別に allow へ書かない限り承認待ちになる  # → layer:platform
171：2026-08-25 [chore/playback-worker-web-layer] git mv は rename を index に載せる。意図した path だけ add したつもりでも、既に staged の rename が同じ commit に混ざる。commit 直前に staged 一覧を確認し、無関係な staged を外してから commit する  # → layer:platform
172：2026-08-25 [chore/playback-worker-web-layer] Feature と Primitive は import 規則が異なる。同一 dir のまま file 名例外で enforce すると新 file 追加で穴が開く。role が違うなら dir 境界へ昇格させて機械検査する  # → layer:terms
173：2026-08-25 [chore/playback-unit-coverage] gate 設定（threshold・lint rule等）を追加した直後の exit code だけで成功と判断しない。設定が実際に評価対象へ効いているかを、期待される違反時に出るはずの ERROR/WARNING ログの有無で確認する。無視された設定は exit=0 を返し続け、成功に見える  # → layer:terms
174：2026-08-25 [chore/playback-unit-coverage] coverage の全体 threshold と層別 threshold を併用する時、層別 threshold の対象 file が全体 threshold の合算対象から自動的に除外されるとは限らない。tool の仕様として全体集計に含まれるなら、層別で緩めた分だけ全体閾値も緩めるか、層別対象を全体閾値の水準まで引き上げるかを選ぶ  # → layer:terms
175：2026-08-25 [chore/playback-unit-coverage] 到達不能に見える分岐へ test を書く前に、型検査で実際に到達可能か検証する。型上も到達不能なら理由付き comment で計測から除外し、型が緩いだけで実は到達可能なら分岐自体を削除するか test で埋める。判断せず test 追加だけで済ませない  # → layer:terms
176：2026-08-25 [chore/playback-unit-coverage] test runner が複数 test 単位（unit/integration 等）を1 process 内の階層構成で束ねる時、集計系の設定（coverage・reporter 等）は階層の末端ではなく最上位でしか有効にならない場合がある。末端に書いて無視されていないか、実際に閾値違反を起こして確認する  # → layer:platform
177：2026-08-26 [docs/architecture-reconsider-react-hono] 質問tool・plan mode切替toolがpermission規約でdenyされた時、機能停止と扱わず、同じ目的（判断の提示・承認確認）をその場のtext出力とuserの次inputで代替する。tool不通は判断や進行を止める理由にしない  # → layer:workflow
178：2026-08-26 [docs/architecture-reconsider-react-hono] workflow skillの説明文にある動詞（show等）を、file書き込みを伴う動詞（create・write・fix）と同一視しない。skill本文の分類が「提示」を指す時は、承認前にartifactを実体化しない  # → layer:workflow
179：2026-08-26 [docs/architecture-reconsider-react-hono] userの短い指示語（remember等）を字面通りの機能呼び出しと解釈する前に、直前の文脈（何を保存すべき情報として提示したか）と照らして意図を確認する。誤った解釈で副作用のある処理を実行すると、後から全体を取り消す作業が発生する  # → layer:workflow
180：2026-08-26 [docs/architecture-reconsider-react-hono] 委譲先が「成功」と報告した内容でも、実行に使ったtoolchain・package manager・生成物が対象repoの既存の正本（lockfile形式、config file）と一致しているかを個別に確認する。報告の成功可否と、環境への副作用の正しさは別の検証項目である  # → layer:workflow
181：2026-08-26 [docs/architecture-reconsider-react-hono] 言語・toolchainの設定directiveが期待通り機能するかは、設定を追加した直後に実際のtoolを実行して確認する。ドキュメントの記憶や類推だけでは、同義とみなした値の組み合わせが実行系によって自動的に無効化される仕様を見落とす  # → layer:terms
182：2026-08-26 [docs/architecture-reconsider-react-hono] 特定の値をハードコードで検証するtestがある状態でその値の重複除去（DRY化）を行うと、実装は正しくなってもtestの断定的assertionが失敗する。値の同期を求めるtestを見つけたら、実装のDRY化とtestの検証方法（直書き値の比較→参照先の存在確認）を同じ単位で直す  # → layer:terms
183：2026-08-26 [docs/architecture-reconsider-react-hono] git操作（stage解除・reset等）がpolicyでblockされた時、blockされた操作を別コマンドで回避しようとせず、現在stageされている内容をそのまま活かせる単位でcommit粒度を再設計する  # → layer:platform
