---
name: topic の seek bar は再生位置を決めるだけで、停止中に再生を開始しない
date: 2026-09-04T23:17:00
branch: feature/playback-e2e-redeploy-master
---

## 1. Decision

`useEpisodePlayback.seek(episodeId, audioRef, positionSec)` は、呼んだ時点で実際に再生進行中（`phase:"loading"` または `"playing"`）なら、その episode（別 episode なら音源を張り替えて）をその位置から再生継続する。何も再生していなければ（`idle`、または `active` でも `paused`/`ended`/`error` で止まっている）、位置だけ動かし、`phase` は `paused` のまま再生を開始しない。判定は episode の同一性ではなく「今まさに音が出ている（出ようとしている）か」だけで行う。

先行 Decision（本 file 群にあった「押した時点の state に関わらずその位置から再生する」という記述）は本 Decision が上書きする。

## 2. Reason

1. topic の sec bar は「そこから聴く**位置**を決める」操作であり、「**再生を始めろ**」という独立の指示ではない。停止は shim の明示操作（「停止」button）の結果であり、位置だけを見るために sec bar を押した操作が再生開始という別の意図を持ち込むと、user の操作意図と実際の挙動がずれる。
2. 判定を「今音が出ているか」の1軸にし episode の同一性を問わないのは、別 episode の sec bar を押した時の挙動を一貫させるため。今何かを聴いている最中に別 topic を押せば聴き続け、何も聴いていない時に押せば止まったまま位置だけ決まる、という単純な規則の方が、episode 単位で例外を作るより例外が少ない。

## 3. Rejected

1. 「同じ episode が停止中」だけを対象にし、idle（一度も再生していない）からの seek は従来どおり再生開始する — idle も「今音が出ていない」点で paused と変わらず、この境界だけ再生を始めると「初回は再生される・2回目以降は止まったまま」という user から見て説明しづらい非対称が残る。
2. 停止中の seek で再生を開始しないが、`phase` は `loading` のまま次の「再生」button 押下を待つ（paused ではなく別の中間状態を新設する）— 既存の `PlaybackPhase` 語彙（`loading`/`playing`/`paused`/`ended`/`error`）に「位置は決まったが未再生」を表す値が無い。新しい値を1つ足すコストに見合う区別ではなく、`paused`（既に「今は鳴っていない」を表す既存語彙）で足りる。

