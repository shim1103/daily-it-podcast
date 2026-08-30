# Drive 配置契約

Generator が書き、Playback（BFF）が読む。載る成果物は音声と原稿 JSON のみ。

## このファイルの責務

**書く:** フォルダ**内**のファイル種類・命名・ペアリング。読取時の対応の仕方。

**書かない:** 原稿 JSON のフィールド（→ `manuscript.schema.json`）。OAuth・フォルダ ID・API 手順・prompt・TTS・UI（→ 各 app の Infrastructure / 実行設定 / README）。

## 検証の分担（Generator）

| 知識 | 正本 | 誰が enforce |
|------|------|----------------|
| 原稿 JSON の field・`episodeId` と stem の一致 | `manuscript.schema.json` | **Application**（書込 UseCase の直前） |
| ファイル名・拡張子・folder 内配置 | 本 file | **Infrastructure** の保存 Adapter（HTTP put の name / MIME / parent） |

Generator の保存 Adapter は schema を import しない。配置（`{episodeId}.json` / `{episodeId}.wav`）だけを実装する。Playback worker の読取 Adapter は従来どおり読取直前に schema を enforce する。

## 配置

所定 folder は実行設定 `DRIVE_FOLDER_ID`（値はここに書かない）。**当該 folder 直下**に episode 用 file を置く（episode ごとの sub folder は作らない）:

| 種別 | 名前 |
|------|------|
| 音声 | `{episodeId}.wav` |
| 原稿 | `{episodeId}.json` |

- `{episodeId}` は不透明な対応キー。両ファイルで同一。生成規則は Generator に閉じ、Reader は stem 一致だけ見る
- JSON は `manuscript.schema.json` に適合し、中の `episodeId` は stem と一致

## 読み

- 一覧: `*.json` を列挙し stem を `episodeId` とする。音声の有無は見ない
- 1 件: `{episodeId}.json` と対応音声の再生用参照を返す。音声が無い・JSON が不適合な件は返さない
- UI にファイル名を出さない。表示日付は JSON の `date`
