---
name: listEpisodes の各 item に audioRef を載せる
date: 2026-08-30T03-19-00
branch: feature/playback-audio-player-ui-design-api-refactoring
---

## 1. Decision

1. `ListEpisodesResponse.episodes[]` の各 item に `audioRef`（非空 string）を必須 field として載せる
2. 値は既存の `episodeAudioPath(episodeId)` で組み立て、GetEpisode の `audioRef` と同じ path 規約に揃える
3. list 取得時に client が別途 GetEpisode しなくても、一覧の再生開始に必要な直結 path を持てる

## 2. Reason

1. 一覧の play 操作は episode 選択（detail 展開）と直交する。list に path が無いと、再生のたびに GetEpisode が要り、関心と関心の結合が強くなる
2. GetEpisode 側は既に `audioRef` を持つ。list だけ欠けると契約が非対称になり、client が list/detail で別経路を覚える
3. path の組み立てを `episodeAudioPath` に閉じると、worker と contract の正本が1つになる（DRY）

## 3. Rejected

1. list は `episodeId` だけ返し、client が path を組み立てる — path 規約が contracts から漏れ、client ごとに複製される
2. 再生前に必ず GetEpisode する — selected と played の直交を壊し、未選択再生ができない
3. list の `audioRef` を optional にする — 欠落時の分岐が増え、契約が曖昧になる
