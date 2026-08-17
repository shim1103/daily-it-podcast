# Drive 配置契約

Generator が書き、Playback（BFF）が読む。載る成果物は音声と原稿 JSON のみ。

## このファイルの責務

**書く:** フォルダ**内**のファイル種類・命名・ペアリング。読取時の対応の仕方。

**書かない:** 原稿 JSON のフィールド（→ `manuscript.schema.json`）。OAuth・フォルダ ID・API 手順・prompt・TTS・UI（→ 各 app の Infrastructure / 実行設定 / README）。

## 配置

所定フォルダは実行設定（値はここに書かない）。直下に:

| 種別 | 名前 |
|------|------|
| 音声 | `{episodeId}.mp3` |
| 原稿 | `{episodeId}.json` |

- `{episodeId}` は不透明な対応キー。両ファイルで同一。生成規則は Generator に閉じ、Reader は stem 一致だけ見る
- JSON は `manuscript.schema.json` に適合し、中の `episodeId` は stem と一致

## 読み

- 一覧: `*.json` を列挙し stem を `episodeId` とする。音声の有無は見ない
- 1 件: `{episodeId}.json` と対応音声の再生用参照を返す。音声が無い・JSON が不適合な件は返さない
- UI にファイル名を出さない。表示日付は JSON の `date`
