---
name: web↔worker HTTP 契約は TS schema。List は JSON のみ
date: 2026-08-17T17:40:00
branch: feature/playback-web
---

## 1. Decision

1. web↔worker の HTTP 契約の正本は `apps/playback/contracts/` の TS schema。repo 根 `contracts/` は Drive のまま
2. ListEpisodes は JSON の目録だけを返す。wav の有無は GetEpisode で初めて見る
3. 宣言した HTTP status（200 / 400 / 404 / 503）はその行。表に無い番号は `floor(status / 100)` の級へ落とす
4. 不完全ペア専用の error `code` は作らない。Get で渡せない件は `episode_not_found`

## 2. Reason

1. JSON Schema は generator（Go）と playback（TS）が言語を跨ぐ Drive 表現用。web と worker は同じ TS なので型と exhaustiveness は TS schema の方が正確
2. 一覧は目録。音声は再生操作が初めて必要とする
3. 静的な宣言表と未知番号の runtime fallback は両方要る。既知 404 を級で 400 に潰さない
4. 不完全ペアを外へ出すと Drive 内部事情が漏れる。HTTP の失敗は `episode_not_found` / `validation_error` / `unavailable` だけ

## 3. Rejected

1. web↔worker にも JSON Schema file を置く案（言語跨ぎが無いのに Drive 契約の形式をコピーする）
2. HTTP 契約を repo 根 `contracts/` に混ぜる案（generator が playback HTTP を知る経路になる）
3. List で wav と JSON をペア確認する案（再生しない操作が音声を要求する）
4. 不完全ペア専用 `code` を返す案
5. HTTP 契約に ViewModel / 画面分岐を書く案（表示は HTTP 境界の外）
