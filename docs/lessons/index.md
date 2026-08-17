# lessons

1：2026-08-15 [develop] 利用者が「案を出せ」と明示した段階ではファイルを変更せず、採用指示後に初めて書き込む  # → layer:workflow
2：2026-08-15 [develop] 表示用日付と対応キーと生成時刻を1つの timestamp に兼ねない。表示値は書込側が確定し、読取 UI は切り出し規則を持たない  # → layer:terms
3：2026-08-15 [develop] 小規模リポジトリの構造ドキュメントはパス一覧ではなく依存規則だけを書き、詳細定義は既存 skill へ委譲して重複を避ける  # → layer:0:meta
4：2026-08-15 [develop] デプロイ単位で機能を分け、各単位の内部は「何を知っているか」の層で分ける。両軸を同一ディレクトリ命名に混在させない  # → layer:terms
5：2026-08-15 [develop] process 境界に載る保存物の形だけを契約にし、認証・フォルダ ID・生成 prompt など手段は契約に入れない  # → layer:terms
6：2026-08-15 [docs/agentsecrets-secret-management] repo と一緒に共有すべき agent の読み取り禁止は HOME の global 設定ではなく、その repo 配下の設定へ置く  # → layer:workflow
7：2026-08-15 [docs/agentsecrets-secret-management] 未完了記録用の tasks/todo は明示指示があるまで作らない。決定記録や実装と自動でセットにしない  # → layer:workflow
8：2026-08-15 [docs/agentsecrets-secret-management] AgentSecrets の project 紐付けは git root 自動検出ではなく、カレントディレクトリの project.json 基準である  # → layer:platform
9：2026-08-15 [docs/agentsecrets-secret-management] AgentSecrets の push 失敗原因は「secret が空」ではなく project 未紐付けであることが多い。空でも push は通る  # → layer:platform
10：2026-08-15 [docs/agentsecrets-secret-management] 同一対象への deny を permissions と sandbox など複数層に書くとき、glob のカバー範囲表記を揃えないと意図がブレる  # → layer:platform
11：2026-08-15 [docs/agentsecrets-secret-management] repo 共有する紐付け設定に個人表示名・ローカル操作時刻を残すとノイズ差分と個人情報の混入になる。共有に必要な ID 以外は空にする  # → layer:workflow
12：2026-08-15 [docs/agentsecrets-secret-management] create-pr 実行指示で gh 禁止があるとき、調査用の gh 呼び出しも禁止対象に含める。wrapper 経由以外の gh を使わない  # → layer:workflow
6：2026-08-15 [feature/x-api-adoption] decisions には指示された判断だけを1ファイルで残す。推測で仕様を確定したことにせず、指示外の判断ファイルを増やさない  # → layer:0:meta
7：2026-08-15 [feature/x-api-adoption] 利用者が完了や execute を明示したら、確認待ちや Plan mode 退避で止めない  # → layer:workflow
8：2026-08-15 [feature/x-api-adoption] 公開境界の契約コメントの自然言語は、利用者が指定した言語に合わせる  # → layer:0:meta
9：2026-08-15 [feature/x-api-adoption] manager は Port・Domain 型・契約・後回し範囲まで先に固定し、具象実装と test は所有 path を切った Issue へ委譲する  # → layer:workflow
13：2026-08-15 [develop] 追記専用の通し番号 list が branch 分岐で番号衝突した時、既存 entry の番号は振り直さず両側をそのまま残す。番号の一意性より記録の保存を優先する  # → layer:workflow
14：2026-08-16 [docs/agentsecrets-secret-export] 提案だけの指示では file を書かない。採用指示後に初めて記録・実装する  # → layer:workflow
15：2026-08-16 [docs/agentsecrets-secret-export] skill 例文の説明句は test へコピペしない。GWT ラベルの後にはその case の前提・呼出・期待を書く  # → layer:0:meta
16：2026-08-16 [docs/agentsecrets-secret-export] test の主対象は期待振る舞い。失敗系は公開された postcondition として契約した範囲だけを個別 case にする  # → layer:terms
17：2026-08-16 [docs/agentsecrets-secret-export] zero-knowledge HTTP proxy は注入ヘッダだけでなく session token 検証がある。公式ヘッダ一覧と Architecture の検証順を両方見る  # → layer:platform
18：2026-08-16 [docs/agentsecrets-secret-export] 公開契約コメントの自然言語は利用者が指定した言語に合わせる  # → layer:0:meta
14：2026-08-15 [chore/test-and-ci] ホーム配下の AGENTS.md は製品の User Rules ではない。User Rules は cloud 保存で symlink できない  # → layer:platform
15：2026-08-15 [chore/test-and-ci] 使い方（実行手順）と規則（配置・gate）を別文書に分けたなら、同じ知識を両側へ書かない  # → layer:0:meta
16：2026-08-15 [chore/test-and-ci] git worktree では .git が directory とは限らない。hook 配置は git-path で解決する  # → layer:platform
17：2026-08-15 [chore/test-and-ci] 空の package 集合に対し list が成功でも test が失敗することがある。入口 script は空集合を成功として扱う  # → layer:platform
18：2026-08-15 [chore/test-and-ci] create-rule が作るのは project の rules であり、User Rules（global）ではない。道具と目的を取り違えない  # → layer:workflow
19：2026-08-16 [docs/agentsecrets-secret-export] 秘密キー名の定数は値と誤読されない名前にする（例: Suffix Name）  # → layer:terms
20：2026-08-16 [docs/agentsecrets-secret-export] 秘密キー名の定数は Name 接尾辞などで値と誤認しない命名にする  # → layer:terms
21：2026-08-17 [feature/x-post-source-adapter] vendor 公式の auth 方式と秘密注入の語形は別知識。凍結前に両方を突き合わせ、一致しない注入を正にしない  # → layer:platform
22：2026-08-17 [feature/x-post-source-adapter] Port の契約コメントは interface 宣言が SSoT。実装 method へ @ensure を複製しない  # → layer:terms
23：2026-08-17 [feature/x-post-source-adapter] 同じ observable な失敗へ落ちる内部枝は、枝ごとの test case を増やさない。公開した postcondition の所有者を1つにする  # → layer:terms
24：2026-08-17 [feature/x-post-source-adapter] 秘密名の正は1箇所に置き、decision・DESIGN・todo へ別名を再掲しない  # → layer:0:meta
25：2026-08-17 [feature/x-post-source-adapter] 下位の配線変換を検証する test と、上位がどの注入を選ぶかを検証する test は所有者を分ける。全寄せすると選択事実が無所有者になる  # → layer:terms
26：2026-08-17 [feature/x-post-source-adapter] create-pr wrapper は repo 根の workflow/project.toml を必須とする。欠けると代替経路へ落ちる前に欠落を表面化する  # → layer:workflow
27：2026-08-17 [feature/x-fetch-watched-posts-usecase] simplify の低情報 comment 削除を GWT 構造 label に適用しない。GWT tag は意図説明ではなく構造 label である  # → layer:terms
28：2026-08-17 [feature/x-fetch-watched-posts-usecase] 完了条件に skill 準拠があるとき、manager は path 委譲だけで足りたとせず reviewer に照合させる  # → layer:workflow
29：2026-08-17 [feature/x-fetch-watched-posts-usecase] 公開契約の postcondition（部分結果なし等）は先頭失敗だけでは不足しうる。後続失敗 path を別 case にする  # → layer:terms
27：2026-08-17 [chore/test-and-ci] 依存ゼロの自作検査は見た目の単純さであり、標準と生態系の解析器を捨てて再発明することではない。Least Power は独自 DSL を増やさない側に働く  # → layer:0:meta
28：2026-08-17 [chore/test-and-ci] allow-unless-denied は既知の禁止だけを止める。未知の新層を自動で止めたいなら deny 列挙ではなく allow 限定にする  # → layer:terms
29：2026-08-17 [chore/test-and-ci] glob の再帰 pattern は直下 file を含まないことがある。直下と子孫の両方を対象にするなら両方の pattern を書く  # → layer:platform
30：2026-08-17 [chore/test-and-ci] 計測と閾値 fail は別能力である。言語標準が cover を出せても、未満で落とす手段が無いなら gate は別途になる  # → layer:platform
31：2026-08-17 [chore/test-and-ci] Adapter が Port を実装するための import と、組み立て点の結線は別行為である。前者は interface の型が要り、後者だけが全層を組む  # → layer:terms
32：2026-08-17 [chore/test-and-ci] 確認待ちを structured question tool に載せる前に、その tool が hook で常時 deny されていないかを見る。deny されるなら本文に選択肢を書く  # → layer:workflow
