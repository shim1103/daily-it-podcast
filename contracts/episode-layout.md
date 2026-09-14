# Episode 配置契約

Generator が書き、Playback（BFF）が読む。載る成果物は音声と原稿 JSON のみ。
storage 実装（現行 Google Drive / 将来 R2）はこの file に書かない。

## このファイルの責務

**書く:** 所定 object 空間**内**の種類・命名・ペアリング。読取時の対応の仕方。

**書かない:** 原稿 JSON のフィールド（→ `manuscript.schema.json`）。credential・bucket/folder ID・API 手順・prompt・TTS・UI（→ 各 app の Infrastructure / 実行設定 / README）。

## 検証の分担（Generator）

| 知識 | 正本 | 誰が enforce |
|------|------|----------------|
| 原稿 JSON の field・`episodeId` と stem の一致 | `manuscript.schema.json` | **Application**（書込 UseCase の直前） |
| object 名・拡張子・空間内配置 | 本 file | **Infrastructure** の保存 Adapter（put の name / MIME / 親空間） |

Generator の保存 Adapter は schema を import しない。配置（`{episodeId}.json` / `{episodeId}.mp3`）だけを実装する。Playback の読取 Adapter は読取直前に schema を enforce する。

## 配置

所定の object 空間は実行設定が指す（値はここに書かない）。**当該空間の直下**に episode 用 object を置く（episode ごとの sub 空間は作らない）:

| 種別 | 名前 |
|------|------|
| 音声 | `{episodeId}.mp3` |
| 原稿 | `{episodeId}.json` |

- `{episodeId}` は不透明な対応キー。両 object で同一。生成規則は Generator に閉じ、Reader は stem 一致だけ見る
- JSON は `manuscript.schema.json` に適合し、中の `episodeId` は stem と一致

## 読み

- 一覧: `*.json` を列挙し stem を `episodeId` とする。音声の有無は見ない
- 1 件: `{episodeId}.json` と対応音声の再生用参照を返す。音声が無い・JSON が不適合な件は返さない
- UI に object 名を出さない。表示日付は JSON の `date`
