# lessons

2：2026-08-15 [develop] 表示用日付と対応キーと生成時刻を1つの timestamp に兼ねない。表示値は書込側が確定し、読取 UI は切り出し規則を持たない  # → layer:terms
4：2026-08-15 [develop] デプロイ単位で機能を分け、各単位の内部は「何を知っているか」の層で分ける。両軸を同一ディレクトリ命名に混在させない  # → layer:terms
5：2026-08-15 [develop] process 境界に載る保存物の形だけを契約にし、認証・フォルダ ID・生成 prompt など手段は契約に入れない  # → layer:terms
7：2026-08-15 [docs/agentsecrets-secret-management] 未完了記録用の tasks/todo は明示指示があるまで作らない。決定記録や実装と自動でセットにしない  # → layer:workflow
8：2026-08-15 [docs/agentsecrets-secret-management] AgentSecrets の project 紐付けは git root 自動検出ではなく、カレントディレクトリの project.json 基準である  # → layer:platform
9：2026-08-15 [docs/agentsecrets-secret-management] AgentSecrets の push 失敗原因は「secret が空」ではなく project 未紐付けであることが多い。空でも push は通る  # → layer:platform
9：2026-08-15 [feature/x-api-adoption] manager は Port・Domain 型・契約・後回し範囲まで先に固定し、具象実装と test は所有 path を切った Issue へ委譲する  # → layer:workflow
17：2026-08-16 [docs/agentsecrets-secret-export] zero-knowledge HTTP proxy は注入ヘッダだけでなく session token 検証がある。公式ヘッダ一覧と Architecture の検証順を両方見る  # → layer:platform
14：2026-08-15 [chore/test-and-ci] ホーム配下の AGENTS.md は製品の User Rules ではない。User Rules は cloud 保存で symlink できない  # → layer:platform
16：2026-08-15 [chore/test-and-ci] git worktree では .git が directory とは限らない。hook 配置は git-path で解決する  # → layer:platform
18：2026-08-15 [chore/test-and-ci] create-rule が作るのは project の rules であり、User Rules（global）ではない。道具と目的を取り違えない  # → layer:workflow
21：2026-08-17 [feature/x-post-source-adapter] vendor 公式の auth 方式と秘密注入の語形は別知識。凍結前に両方を突き合わせ、一致しない注入を正にしない  # → layer:platform
25：2026-08-17 [feature/x-post-source-adapter] 下位の配線変換を検証する test と、上位がどの注入を選ぶかを検証する test は所有者を分ける。全寄せすると選択事実が無所有者になる  # → layer:terms
28：2026-08-17 [feature/x-fetch-watched-posts-usecase] 完了条件に skill 準拠があるとき、manager は path 委譲だけで足りたとせず reviewer に照合させる  # → layer:workflow
27：2026-08-17 [chore/test-and-ci] 依存ゼロの自作検査は見た目の単純さであり、標準と生態系の解析器を捨てて再発明することではない。Least Power は独自 DSL を増やさない側に働く  # → layer:0:meta
28：2026-08-17 [chore/test-and-ci] allow-unless-denied は既知の禁止だけを止める。未知の新層を自動で止めたいなら deny 列挙ではなく allow 限定にする  # → layer:terms
31：2026-08-17 [chore/test-and-ci] Adapter が Port を実装するための import と、組み立て点の結線は別行為である。前者は interface の型が要り、後者だけが全層を組む  # → layer:terms
33：2026-08-17 [feature/tts-speech-synthesizer] 公開境界は呼び出し入口が未決でも、1操作の入出力契約として先に固定できる  # → layer:terms
34：2026-08-17 [feature/tts-speech-synthesizer] vendor 固有の model・voice・演出は Adapter 定数に閉じ、空でも Port 引数へ上げない  # → layer:terms
35：2026-08-17 [feature/tts-speech-synthesizer] 再試行の上限を未設定のままにすると打ち切れない。有限の正の回数を先に置く  # → layer:terms
36：2026-08-17 [feature/tts-speech-synthesizer] 公式が指示文の読み上げ失敗を書く合成系では、本文と演出を同じ入力に混ぜず Adapter が包む  # → layer:platform
38：2026-08-17 [feature/tts-speech-synthesizer] 凍結済み旧実装の置き場が directory とは限らず tag のことがある。path 不在を記録不在と読まない  # → layer:workflow
40：2026-08-18 [feature/tts-speech-synthesizer] 再生・保存が目的なら非可逆圧縮を足さず、標本を自己記述の標準容器へ包む。圧縮は帯域要件が現れてからにする  # → layer:terms
41：2026-08-18 [feature/tts-speech-synthesizer] 同一製品名の音声合成でも API 窓口が違うと戻り形式が違う。直出し encoding を前提にせず、その窓口の公式例を正とする  # → layer:platform
42：2026-08-18 [feature/tts-speech-synthesizer] 標本列と容器は別知識。容器なしの生標本を保存形式にすると、読取側が標本パラメータを契約として持つ  # → layer:terms
45：2026-08-18 [feature/tts-speech-synthesizer] 外部 module を外したあと checksum file だけ残さない。require が無いなら checksum も不要  # → layer:platform
46：2026-08-18 [feature/tts-speech-synthesizer] 有名な圧縮器への乗り換えは不安の解消であり、目的に対して過剰な力なら Least Power に負ける  # → layer:0:meta
33：2026-08-17 [feature/x-getxapi-adapter] 外部系で束ねた directory は Adapter の棚であり、差分吸収 facade ではない。統一 interface は Application の Port、切替は Composition Root  # → layer:terms
34：2026-08-17 [feature/x-getxapi-adapter] Driven Adapter の見た目が似ていても、vendor 契約（auth・page 名・error 形）が違うなら mechanism を共通化しない。同じ知識が繰り返されてから分ける  # → layer:terms
35：2026-08-17 [feature/x-getxapi-adapter] 作業レーン上の通し番号と GitHub Issue 番号は別識別子。片方の番号で他方を指したことにしない  # → layer:workflow
41：2026-08-18 [chore/ci-script-gha-dry] 手順を1箇所に置くこと（知識の DRY）と、同じ手順を複数 runner で実行することは別である。後者は再現であり重複ではない  # → layer:terms
48：2026-08-18 [feature/tts-speech-synthesizer] remote の mergeable は push 直後に CONFLICTING を残すことがある。local の conflict 解消後に再取得して判定する  # → layer:workflow
49：2026-08-18 [feature/tts-speech-synthesizer] hook が別 app の unit を回すとき、merge で入った package の依存未導入は import 失敗になる。hook 失敗を merge 内容の欠陥と同一視しない  # → layer:platform

78：2026-08-18 [refactor-playback-worker-http] Workers の HTTP response body に byte を返す場合、`Uint8Array.buffer` の `SharedArrayBuffer` union により `ArrayBuffer` 契約を崩し得る。境界で正規化して型と観測可能な body を揃える。 # → layer:platform

79：2026-08-18 [pr-c-playback-biome-tsc] sandbox の read-deny 対象 file は、全体scanする `git status` では変更なし扱いに落ちるが、path 指定の `git diff -- <path>` は open 失敗を deleted と誤判定することがある。両者が食い違う時は `git status` の全体判定を正とし、疑わしい削除は復元前に判断材料不足を declare する  # → layer:platform
80：2026-08-18 [pr-c-playback-biome-tsc] uncommitted 変更の取り消しは `git checkout`/`git restore` が hook で deny されることがある。tracked file を元に戻す代替として、編集tool（Edit等）で HEAD 相当の内容へ直接書き戻す手段を残す  # → layer:workflow
79：2026-08-18 [chore/generator-static-lint-format] サンドボックスの read 拒否で出る stderr の file 名は、対象範囲を確定させる `git diff`/`git status` の構造化出力（--porcelain・--name-status）と別物である。査読 agent の指摘は委譲元が自分でその構造化出力を取り直してから採否を判断する # → layer:workflow
80：2026-08-19 [feature/playback-worker-drive-adapter] 既存の公開 symbol の signature を変えない制約を優先すると、新しい分岐条件（env 等）が固定値経由でしか呼ばれない配線を作りうる。受け入れ条件が「実行時に本当にその分岐へ到達するか」を要求する時は、既存 export を保つことより実行経路の到達可能性を先に検証する # → layer:workflow
81：2026-08-19 [feature/playback-worker-drive-adapter] 「throw しない」という組み立て層の禁止は、値の妥当性判定という責務自体を放棄してよい意味ではない。判定結果を無言で代替実装へ逃がすと、設定漏れが観測不能になる。判定はして、判定の結果どう失敗を表現するか（throw か戻り値の分類か）だけを禁止範囲の外へ出す # → layer:terms
82：2026-08-19 [refactor/generator-source-port] 抽象境界の optional field は構造化である。一部実装だけが埋める形を上位 schema に置くと、上位が欠落の意味を解釈し始める。必須以外は opaque な余りへ寄せ、上位は parse しない  # → layer:terms
83：2026-08-19 [refactor/generator-source-port] 上位の契約と下位の内部 field を同じ会話で混ぜると、必須 schema が具象の形に汚染される。層を分けてから field の帰属を決める  # → layer:terms
84：2026-08-19 [refactor/generator-source-port] 情報要素の時刻は情報源データに付随する発生時刻であり、取得時刻でも呼び出し元の実行時刻でもない。実行時刻から導くのは窓の下限だけ  # → layer:terms
85：2026-08-19 [refactor/generator-source-port] 一意性の保証は返す側の義務である。下流が同一主体を推論するための識別子は、必須 schema ではなく余りへ載せてよい  # → layer:terms
86：2026-08-19 [feature/generator-drive-adapter-redo2] `log-session` を読んだだけで daily だけ書いて終えると、shim 指摘や agent の誤読が `lessons` へ残らない。logging の振り分けでは、実装説明から確認論点が誤解・判断ミスへ移った時点で、その誤り自体を `docs/lessons/` 候補として先に切り出す  # → layer:workflow
87：2026-08-19 [feature/generator-drive-adapter-redo2] KISS/DRY 指摘で doc を薄くする時も、情報を消すのではなく記録先を分ける。README / DESIGN / decision / Issue / daily / lessons の責務を混ぜると、修正後の agent が transcript なしで再現できなくなる  # → layer:0:meta
88：2026-08-19 [feature/generator-drive-storage-adapter] 呼び出し元が対象 branch・起点を明示指定している時、既存 branch が対象 Issue に関連する commit を持っているという状況証拠だけで、その branch を対象と決め打ちしない。明示指定と状況証拠が食い違う場合は明示指定を優先し、決め打ち前に対象の実在（起点 branch・commit）を裏取りする  # → layer:workflow
89：2026-08-19 [feature/generator-drive-storage-adapter] 確認 tool が deny され、deny 理由に「自律判断で完遂し、質問は stdout へ出力せよ」と明記された後は、同種の判断が再度必要になっても同じ tool を再試行しない。1 回の deny 理由を以降の全判断の運用方針として引き継ぐ  # → layer:workflow
90：2026-08-19 [feature/generator-drive-storage-adapter] 公開契約文書に書かれた invariant は、書かれているという理由だけで現行仕様と断定しない。契約文言も実装同様に変更履歴を持つ。文言が指す語彙が実装から消えている・別 Port が同種語彙を新しい語彙へ改訂済みである等の兆候があれば、契約を書いた commit まで遡り、旧設計の残骸か恒久 invariant かを先に切り分ける  # → layer:terms
91：2026-08-19 [feature/generator-drive-storage-adapter] 秘密値を注入する proxy 越し通信を「実際に HTTP を飛ばす」とだけ説明すると、外部 API へ接続していると誤解される。宛先 URL がヘッダ等の値として運ばれるだけで、実接続先は迂回用の代役 endpoint という構造は、消費先が本物か代役かを明示して初めて正確になる  # → layer:terms
88：2026-08-19 [feature/playback-worker-http-refactor] `non-edit` roleを理由に、明示されたexecutor委譲・検証・完了処理まで止めてはいけない。roleの権限境界とuserが要求した作業範囲を分け、委譲可能な実装はexecutorへ渡す  # → layer:workflow
89：2026-08-19 [feature/playback-worker-http-refactor] 完了済みtodoの削除と未完了follow-upの保持を同じ判断に混ぜない。完了契約を満たしたfileだけを削除し、未完了作業は別todoとして残す  # → layer:workflow
90：2026-08-19 [tmp-branch] agentを自立実行させるpromptでは、最初のbranch作成・実行権限・確認待ちを明示し、managerの非編集責務とexecutorの編集責務を混ぜない  # → layer:workflow
91：2026-08-19 [tmp-branch] 複数Adapterが同じ意味の識別子を返す時は共有境界の定数を使い、変換testは部分文字列ではなく必須値・時刻・全出力を検証する  # → layer:terms
92：2026-08-19 [tmp-branch] constants fileへの分割は定数の存在だけで決めず、共有範囲と責務の独立性で決める。test差し替えが必要なmutable設定は定数群へ混ぜない  # → layer:terms
93：2026-08-20 [feature/playback-runtime-config-boundary] 内部Errorを外部Errorへ変換する境界と、外部ErrorをHTTP responseへ直列化する境界を分ける。HTTP status・codeを内部層へ逆流させない  # → layer:terms
94：2026-08-20 [feature/playback-runtime-config-boundary] Broad Integrationは対象boundaryのobservableな協調を検証する。外部実通信を含む正常系まで無理に通すとscopeとdouble境界を壊すため、別Integrationへ分離する  # → layer:terms
93：2026-08-20 [playback-web-api-client] 網羅性検査の`never`代入と、到達時のfallback値は別責務。前者を消すと分岐欠落がcompileを通るため、fallbackの形を変える時も代入自体は残す  # → layer:terms
94：2026-08-20 [playback-web-api-client] callerが失敗codeごとに操作を変える境界では、未知の分類を既存codeへ倒すと起きていない失敗を伝える。想定外の前提違反はResultへ畳まずthrowで大域脱出させる  # → layer:terms
95：2026-08-20 [playback-web-api-client] 隣接層が同名の識別子集合を持つ時、上位の型をそのまま再exportせず、別layerの型として宣言し写像表で1対1に変換する。写像表を全数対応の形にすると上位の追加をcompileが強制する  # → layer:terms
96：2026-08-20 [playback-web-api-client] 応答のstatusとbodyが同じ失敗を1対1で表す時、両方を読むと情報源が二重になる。片方を単一情報源に決め、もう片方は読まない  # → layer:terms
97：2026-08-20 [playback-web-api-client] 外部境界から運ぶ値は、metadataごと保持する器を選ぶと下位層がその知識を書かずに済む。生の値だけを渡すと、下位層がmetadataを再指定する責務を負う  # → layer:terms
98：2026-08-20 [playback-web-api-client] 接続先の位置情報は隠すべき具体値であり、各操作の引数に置くとcallerが同じ値を持ち回る。組み立て時に1度だけ受け取ると正規化規則も構造的に1箇所へ集まる  # → layer:terms
99：2026-08-20 [playback-web-api-client] 決定済みの契約・定数・interfaceを自然言語のIssueへ書き写して別sessionへ渡すと、決定が二重管理になる。契約はcodeで固定し、Issueには残る実装だけを書く  # → layer:workflow
100：2026-08-20 [playback-web-api-client] 分類名をfile名へ持たせる規約は、runnerの収集条件も同じ分類名で絞って初めて機械的に守られる。命名規約だけでは命名忘れを検出できない  # → layer:workflow
101：2026-08-20 [playback-web-api-client] 既存fileが規約違反の状態にある時、新規fileを既存へ揃えるのは違反の追認である。規約と既存のどちらが正かを先に判定する  # → layer:workflow
102：2026-08-20 [playback-web-api-client] 委譲先の報告は実物と独立検証で裏を取る。検証commandの再実行と、報告に無い生成物の確認を両方行う  # → layer:workflow
103：2026-08-20 [playback-web-api-client] 非編集の指示で作業を止める範囲は、成果物の作成有無ではなく編集の有無で決まる。計画fileの作成も編集であり、指示されていなければ行わない  # → layer:workflow
104：2026-08-20 [playback-web-api-client] 質問toolがdenyされる環境では、判断材料が揃う限り自律で確定し、不可逆な変更に関わる時だけ質問事項をstdoutへ出して停止する  # → layer:workflow
105：2026-08-20 [playback-web-api-client] 実装済みcodeを持たない領域へは「今修正する」分類が成立しない。分類軸は変更対象の有無ではなく、決定の固定先がcodeか別sessionかで決める  # → layer:workflow
106：2026-08-20 [playback-web-api-client] sandboxがbuild cacheへの書き込みを禁じると、compile済み言語のlintを呼ぶcommit hookが失敗する。hook失敗を変更内容の欠陥と同一視せず、権限側の原因を切り分ける  # → layer:platform
107：2026-08-20 [playback-web-api-client] 作業tree上の未整理差分が上流branchと完全一致する時、それは作業ではなく上流へ戻す逆差分である。commitせず捨ててよいかは、打ち消される側のcommitがremoteにあるかで判定する  # → layer:workflow
108：2026-08-20 [generator-episode-validation] coverage対象外のComposition Rootでも、実装選択を固定する結線testは省略しない。coverage除外は計測範囲の設定であり、責務境界の検証免除ではない  # → layer:terms
109：2026-08-20 [generator-episode-validation] Driven Portを実装するAdapterと、そのPortを呼ぶUseCaseを同じ型へ兼務させない。Composition RootはUseCaseへAdapterを注入し、呼び出し側がvalidationを迂回できない戻り値を公開する  # → layer:terms
110：2026-08-20 [generator-google-oauth-adapter] `non-edit` の manager は実装を直接変更せず、executor へ委譲して reviewer の指摘を再実行で反映する。権限境界を守ることは完了処理を止めることではない  # → layer:workflow
111：2026-08-20 [generator-google-oauth-adapter] workflow が起点branchを指定している時、既存branchが同じcommitを指す状況証拠で代用せず、指定されたbranch作成手順をそのまま実行する  # → layer:workflow
112：2026-08-20 [playback-runtime-config-http-contract] workflowの仮branch名はtaskのpropertyから機械的に導出し、作成直後に現在branch名を検査する。placeholderを正式branchとして残さない  # → layer:workflow
113：2026-08-20 [playback-runtime-config-http-contract] ApiErrorはUI表示文ではなく操作判断用の中間語彙として保つ。server内部障害と外部service一時不能は、UI表示が同じでもretry意味が違うためcodeを統合しない  # → layer:terms
