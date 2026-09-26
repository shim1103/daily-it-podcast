---
name: 進捗の達成契約は疎通・D1 infra・Application・web同期・web表示・releaseに分け、FakeはAに置かない
date: 2026-09-26T03:48:06
branch: feature/playback-progress
---

## 1. Decision

1. 進捗featureの達成契約は次の境界に分ける（pathの正は`docs/tasks/todo/`）。
   1. 一過性D1疎通（`workflow_dispatch`＋PASS後削除）
   2. D1 infra（local peer本実装・到達NI・Fake・adapter・Root）を**1本**
   3. Application（Write/pull merge と list embed）を**1本**
   4. web同期とweb表示は**別本**（表示は同期の後）
   5. 常設smoke／I・E2E追加とintegration PRは**release検証1本**
2. **Fake／test doubleはAに置かない。** Aは入口の型・signature・stub（zero・実peer未起動）まで。Fakeは当該Cのtest support。
3. list embedはweb同期Issueへ寄せない（worker Applicationの責務）。

## 2. Reason

1. 検証対象が違う（環境疎通／binding到達／merge意味／UI描画／常設gate）ものを同居させると完了条件が閉じない。R2の疎通とAdapter分離と同軸。
2. local peerの到達NIとadapter SUは同じinfra完了に要る。peerとadapterを別IssueにするとNIの所有が割れ、着手順だけが増える。
3. Fakeはtest実行時にだけ要る。境界契約（A）に先置きするとstubと混同し、完成系でない注入物が契約面に残る。
4. embedは`listEpisodes`のApplication合成であり、browser VMのpush/pullとはruntimeが違う。webへ寄せると完了条件とVerificationが溶ける。

## 3. Rejected

1. **local peer Issueとadapter Issueを分ける案** — NIが両方に必要になり、境界が二重になる。
2. **list embedをweb同期Issueへ統合する案** — workerとwebが同居し、Port doubleで閉じるApplication検証が壊れる。
3. **FakeをAの契約stubとして先置きする案** — test専用物を境界契約に混ぜ、stubの意味が壊れる。
4. **一過性dispatchと常設smokeを同一Issueにする案** — 消すものと残すものが同居する（先行R2 Decisionと同軸で却下）。
