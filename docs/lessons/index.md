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
