---
title: Playback Workerのruntime config検証とHTTP Error変換を分離する
date: 2026-08-20T12:57:00+09:00
branch: feature/playback-runtime-config-boundary
---

## 1. Decision

1. Playback Workerのruntime config検証は、repositoryとControllerを結線するComposition Root本体から分離した専用moduleが所有する。
2. config検証は、4 keyを個別の検証logicで確認する。Errorの種類は共通とし、欠落原因ごとにserver log用messageを分ける。
3. local / unit testのInMemory選択は、専用の明示optionがある場合だけ許可する。production相当のenv不足を`misconfigured`という判定結果としてHTTP入口へ返さない。
4. config不備の内部Errorは、HTTP Route HandlerがExternal Errorへ変換する。`createHttpErrorResponse`はExternal ErrorをHTTP responseへ変換する。詳細messageはcause chain経由でserver logへ渡し、外部response bodyにはError codeだけを返す。
5. `README.md`はsecret inventory、`PlaybackEnv`とconfig検証moduleはPlayback Worker runtimeの実行時SSOTとして、責務を分離する。

## 2. Reason

1. Composition Rootは全PortとAdapterを結線する責務を持つが、env keyごとの検証規則まで同じmoduleへ持つと、config変更とrepository配線変更の理由が混在する。
2. 欠落keyは同じruntime config不備というError分類に属する。一方、server logでは原因を識別できる必要があるため、Error classを増やさずmessageを原因ごとに分ける。
3. `misconfigured`をfetchへ返すと、Composition Root内部の判定表現がHTTP境界へ漏れる。Route HandlerでInternal ErrorをExternal Errorへ変換し、その後のHTTP response変換を`createHttpErrorResponse`へ委譲する方が責務が明確になる。
4. `createHttpErrorResponse`はExternal Errorとcause chainをlogし、外部へ公開するError codeを制限する。内部診断と公開契約を同じmessageにしない。
5. secret名の全runtime共通schemaを作ると、Generator・Playback Worker・Playback Webの異なる注入経路を結合する。

## 3. Rejected

1. 欠落keyごとに別のError classを作る案。Error分類を増やし、HTTP mappingとtestの重複を増やす。
2. `misconfigured`をComposition Rootの戻り値として維持する案。内部の判定表現をHTTP境界へ漏らし、External Error変換の責務を曖昧にする。
3. `README.md`のsecret inventoryをruntimeの実行時検証へ流用する案。READMEは運用一覧であり、Workerの注入契約を所有しない。
4. config不備の詳細messageをHTTP response bodyへ返す案。runtimeの設定構造を外部へ露出する。
5. `configuration_error`など新しいHTTP error codeを今回追加する案。既存のPlayback HTTP contract変更を伴うため、別Issueで扱う。
