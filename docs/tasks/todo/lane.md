# 未完了 index

未完了の調査・実測だけをGenerator / Playback横断で管理する。実装済み事項・実施順・方針は載せない。達成契約をIssue化した場合は`docs/tasks/todo/`の個別fileを正とする。

## Generator

| topic | 未完了 |
|---|---|
| compositeのsource横断sort | 複数情報源の`OccurredAt`順をApplication / Compositionのどちらで持つか決める |
| Links件数に応じた先fetch | `SourceBody.Links`件数を見て、TextWriter前にApplicationが取得するか決める |
| 議論commentの深掘り | Hacker News / Lobstersのnested commentを辿るか決める |
| TextWriterのURL取得実測 | Cloud Agents / Geminiが`Detail.Links` / `Discourse.Links`を実際に取得できるかと件数上限を測る |
| no-repo原稿運用実測 | Cloud Agents no-repoの原稿品質・日次消費・GHA所要を測る |
| Gemini free-tier実測 | 日次produceとdraft retryを含む実運用でquotaが足りるか測る |
| 外部呼び出しの完了ログ | `RetryReporter`は異常系（retry試行）だけを観測する。個々の外部呼び出し（HTTP call等）の正常完了ログ（`ProgressReporter`相当）は未追加。粒度をAdapter単位1回にするかHTTP呼び出し単位にするか未決 |
| `LogWriter.Event`の排他制御 | 現在`sync.Mutex`無し。今は逐次実行のみで実害無いが、source取得等をgoroutineで並行化する際にログ行のinterleavingが起きうる。並行化着手前に追加するか、着手時にまとめて追加するか未決 |

## Playback

| topic | 未完了 |
|---|---|
| Access下の二段cache実測 | 音声のedge cache、browser cache、未認証requestがcache hitせずAccessで止まることを本番で確認する |
| PlaybackState.activeのepisodeId/audioRef非正規化 | catalog非依存で独立保持する現設計（Decision 2026-09-04 §1-1）を維持するか、catalogから都度引く正規化形へ寄せるか。独立性（一覧再取得後も再生継続）と同期コストのトレードオフを検討する |
| listEpisodesの原稿parse失敗の観測 | `selectValidListItem`がschema不適合entryを黙って除外する。application層はdependency-cruiser制約でconsole.*を呼べないため、除外時のwarning出力をどこに置くか（use case戻り値経由でroute層まで持ち帰る等）を決める |
