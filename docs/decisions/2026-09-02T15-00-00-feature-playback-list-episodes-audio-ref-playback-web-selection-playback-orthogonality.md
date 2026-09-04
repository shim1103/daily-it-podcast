---
name: playback web では episode の selection（原稿展開）と playback（音声再生）を直交させる
date: 2026-09-02T15:00:00
branch: feature-playback-list-episodes-audio-ref
---

## 1. Decision

1. **selection**（どの episode の原稿を展開するか）と **playback**（どの episode の音声を流すか）は、同一 user intent に束ねない
2. `Select` / `Deselect` は `playedEpisodeId` を変えない
3. `Play` / `Stop` は `selectedEpisodeId` を変えない
4. 別 episode へ `Play` したら、直前の audio は stop し `currentTime` を reset してから新 episode を load する
5. `Deselect` しても再生中 audio は止めない。Entry（原稿）は閉じるが playback は継続する
6. page が 1 つである前提（`2026-08-25T05-10-48-feature-playback-ui-structure`）のため、catalog 失敗・一覧に無い episodeId の選択・audio load 失敗は、いずれも **blocking error** として同一 Error surface で示す。list 専用・detail 専用の Error UI は作らない
7. 直交を UI 責務へ写す。**Row** は 1 episode の meta と select / play affordance を持つ。**Entry** は選択中 episode の manuscript のみ。**AudioControls** は再生対象 episode の audio のみ。`title` / `date` は Row が既に示すため Entry では重ねない（同 Decision `2026-08-25T05-10-48` の DRY を維持）
8. 型・derive 規則・hook / component の signature は A artifact を正とする（`web/src/view-models/playback-state.ts` 他）。hook の層責務・derive と state の一般規則は `architecture/frontend`（`view-model.md` / `feature-component.md`）を正とし、本 file へ再掲しない

置き換え範囲: `2026-09-02T14-30-00-feature-playback-list-episodes-audio-ref-playback-web-orchestration.md` は同一 branch で作成した誤った束ね Decision である。本 file を正とする。維持範囲: 物語・視覚・hash 同期・list `audioRef`・1 page など、別判断軸の先行 Decision は参照のみで維持する。

## 2. Reason

1. 一覧で play しながら別 episode の原稿を読む、または原稿だけ閉じて聴き続ける、という **playback 固有**の使い方を同時に満たすため。`2026-08-30T03-19-00-feature-playback-list-episodes-audio-ref` が list に `audioRef` を載せたのも detail 展開なしに play できる契約であり、UI 側で selection と playback を連動させると API 契約の意図と矛盾する
2. `Deselect` 後も再生継続は、「原稿を畳む」と「聴く」を別 intent として扱うため。連動停止にすると、原稿だけ閉じたい user が毎回再生を失う
3. 別 episode への `Play` で前 audio を reset するのは、同時に 1 本だけが正しい再生対象であることを保つため。reset なしだと切り替え後も前の位置・状態が残り、どの episode を聴いているか曖昧になる
4. **Row / Entry / AudioControls** は直交の結果としての **domain 配置**である。Row に select と play の両 affordance が載るのは「同じ episode 行で二つの intent を選べる」からであり、API field 単位の分解（`episode-topic` 等）を維持する理由にはならない。一般の「user 操作単位で component を切る」規則は frontend skill が持ち、ここでは playback における配置だけを決める
5. blocking error の一本化は、route が無い 1 page app だからである。失敗種別ごとに surface を増やすと、同じ「使えない」状態に複数の拡張点ができ、次の変更でどこを直すか迷う
6. state の derive（例: `isPlaying` は `playbackPhase` から導出）や hook を catalog / selection / hash-sync / playback に分ける **一般規則**は、project Decision ではなく architecture skill と A artifact が担う。本 Decision に混在させると、playback 以外の frontend でも同じ規則を再掲する DRY 違反になる（`design-philosophy.md` §1 SRP / DRY）

## 3. Rejected

1. `Play` が `Select` を暗黙に呼ぶ — 直交を壊し、list 行の play だけで原稿展開が強制される
2. `Deselect` で再生を止める — 原稿閉じと停止を同一 intent に束ね、直交を壊す
3. selection と playback の一般規則（derive 一覧・hook 分割原則）を本 project Decision に書く — skill / A と二重 SSOT になり、1 doc 1責務（SRP）を破る
4. capability 列挙・state inventory 百科・hook 名一覧を本 Decision に書く — 仕様書化であり Decision の問い「直交をどう扱うか」から外れる。正本は A と先行 Decision へ委譲する
5. `2026-09-02T14-30-00-feature-playback-list-episodes-audio-ref-playback-web-orchestration.md` のように、独立した判断軸（一般 frontend 規則 + domain 直交 + component 百科）を 1 file に束ねる — `decisions.md` §4-5 違反。論点番号や設計フェーズを機械的に file 化したものでも同様
