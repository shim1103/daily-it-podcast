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
