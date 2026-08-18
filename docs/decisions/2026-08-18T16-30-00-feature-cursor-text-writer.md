# Cursor CLI TextWriter（Port/定数/呼び出し規則）

## 1. Decision

1. 原稿構築の **Cursor** 呼び出しは、Application 所有 Port `TextWriter` の `Write(brief)` に委譲する。Cursor への入力は brief のみで、結果は text 断片として返す。(`contracts/` / `manuscript.schema.json` と Cursor の中間表現は一致しない)
2. Cursor CLI は非対話の `agent` を使い、argv は決定的に固定する。
   - `agent -p --mode ask --output-format json --model composer-2.5 --sandbox enabled --trust`
   - `--force` / `auto` / fast model は使わない
3. Cursor CLI 成功時は stdout の JSON envelope を解釈し、`result` フィールド（assistant 最終テキスト）だけを text 断片として返す。失敗は非0 exit と stderr をもって Domain ではなく Infrastructure として失敗扱いする（error型は adapter が定義）。
4. 秘密値は Cursor API key の **値**を code / 設定に埋めない。実行環境で `CURSOR_API_KEY` が供給されている前提で、Adapter は env の存在だけを期待する。
5. Adapter の観測面（success/failure）は stdout JSON + exit code を基準にする。stderr の内容文字列を上位へ写さず、内部診断として扱う。

## 2. Reason

1. **DIP**。Application は vendor（Cursor CLI）の argv / stdout envelope / error形を知らない。
2. **不一致の明示**。`manuscript.schema.json` は UI 完成系であり、Cursor の 1 呼び出し単位（text 断片）とは一致しないため、`TextWriter` の戻りは 断片としてのみ扱う。
3. **一貫性（Principle of Least Astonishment）**。Cursor envelope 解釈と argv 決定を adapter に固定し、呼び出し側は「いつ何回呼ぶか」だけを決める。
4. **Least Privilege / KISS**。Cloud API / SDK / 非決定要素の追加で複雑性を増やさない。

## 3. Rejected

1. `TextWriter` Port に model / voice / prompt / json envelope を渡す案（vendor 依存が Port へ漏れる）
2. `TextWriter` Port を `manuscript.schema.json` と 1:1 にして UI と Cursor 単位を強制一致させる案
3. `--force` / `--yolo` を付けた手順短縮（失敗時の観測面が変わりやすい）
4. fast / auto model への切替を adapter の責務にする案

