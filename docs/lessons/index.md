# lessons

1：2026-08-15 [develop] 利用者が「案を出せ」と明示した段階ではファイルを変更せず、採用指示後に初めて書き込む  # → layer:workflow
2：2026-08-15 [develop] 表示用日付と対応キーと生成時刻を1つの timestamp に兼ねない。表示値は書込側が確定し、読取 UI は切り出し規則を持たない  # → layer:terms
3：2026-08-15 [develop] 小規模リポジトリの構造ドキュメントはパス一覧ではなく依存規則だけを書き、詳細定義は既存 skill へ委譲して重複を避ける  # → layer:0:meta
4：2026-08-15 [develop] デプロイ単位で機能を分け、各単位の内部は「何を知っているか」の層で分ける。両軸を同一ディレクトリ命名に混在させない  # → layer:terms
5：2026-08-15 [develop] process 境界に載る保存物の形だけを契約にし、認証・フォルダ ID・生成 prompt など手段は契約に入れない  # → layer:terms
6：2026-08-15 [feature/x-api-adoption] decisions には指示された判断だけを1ファイルで残す。推測で仕様を確定したことにせず、指示外の判断ファイルを増やさない  # → layer:0:meta
7：2026-08-15 [feature/x-api-adoption] 利用者が完了や execute を明示したら、確認待ちや Plan mode 退避で止めない  # → layer:workflow
8：2026-08-15 [feature/x-api-adoption] 公開境界の契約コメントの自然言語は、利用者が指定した言語に合わせる  # → layer:0:meta
9：2026-08-15 [feature/x-api-adoption] manager は Port・Domain 型・契約・後回し範囲まで先に固定し、具象実装と test は所有 path を切った Issue へ委譲する  # → layer:workflow
