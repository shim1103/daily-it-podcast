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

## Playback

| topic | 未完了 |
|---|---|
| Access下の二段cache実測 | 音声のedge cache、browser cache、未認証requestがcache hitせずAccessで止まることを本番で確認する |
