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
