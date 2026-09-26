# lane

未完了・依存の進捗index。達成契約本文は各`docs/tasks/todo/{issue-name}.md`。書き方は`0:meta/logging/tasks`と`1:terms/contribution-lifecycle/branch-strategy`。

## Generator（未決）

1. [ ] compositeのsource横断sort — Application / Compositionのどちらで持つか決める
2. [ ] Links件数に応じた先fetch — TextWriter前にApplicationが取るか決める
3. [ ] 議論commentの深掘り — nestedを辿るか決める
4. [ ] TextWriterのURL取得実測 — 取得可否と件数上限を測る
5. [ ] no-repo原稿運用実測 — 品質・日次消費・GHA所要を測る
6. [ ] Gemini free-tier実測 — 日次produce＋draft retryでquotaが足りるか測る

## Playback（未決）

1. [ ] Access下の二段cache実測 — edge／browser／未認証がAccessで止まることを本番確認
2. [ ] PlaybackState.activeのepisodeId/audioRef非正規化 — Decision 2026-09-04 §1-1維持か正規化か
3. [ ] listEpisodesの原稿parse失敗の観測 — 除外warningの置き場を決める

## release

```text
master（release）
  └ develop（integration）
       └ docs/playback-audio-history（feature-integration）
            ├ part/playback-d1-dispatch-smoke ← docs/tasks/todo/playback-d1-dispatch-smoke.md
            ├ part/playback-d1-progress-infra ← docs/tasks/todo/playback-d1-progress-infra.md
            ├ part/playback-progress-application ← docs/tasks/todo/playback-progress-application.md
            ├ part/playback-web-progress-sync ← docs/tasks/todo/playback-web-progress-sync.md
            │    └ part/playback-web-progress-ui ← docs/tasks/todo/playback-web-progress-ui.md
            └ part/playback-progress-release-verify ← docs/tasks/todo/playback-progress-release-verify.md
                 （上のpart到達後。完成後 → PR → develop）
  developの完成束 → PR → master
```

### 未完了

1. [ ] 進捗clock skew秒 — 事実: 数値未決 / 次: 決定または仮値運用のまま落とすか
2. [ ] 進捗push／pull間隔 — 事実: 数値未決 / 次: sync Issueは仮値可、正本は後で
3. [ ] `durationSec < 3` 短尺の完走扱い — 事実: 未決 / 次: 定数化の要否
4. [ ] catalog寿命≠progress寿命／hook所有権 — 推測: hooks内で分かれうる / 次: 材料が足りたらDecision。公開面は凍らせない
