---
name: 再生進捗のlist向けGetはlistEpisodes応答へembedし、catalogとprogressを1往復で返す
date: 2026-09-19T16:26:00
branch: feature/playback-progress
---

## 1. Decision

1. episode一覧のGet（現行の `listEpisodes` 相当）は、Server側で進捗storeを読み、各episodeに進捗を**embed**して1応答で返す。
2. list表示に必要な進捗の意味（`null` または各字段の意味）は先行Decision（`2026-09-19T16-25-18-feature-playback-progress.md`）に従う。本 Decision は**返し方（合成単位）**だけを固定し、字段の百科は写さない。
3. 進捗専用のlist向けGetを初手では分けない。DB上でcontent（R2）と進捗（D1）が分かれていることと、HTTP Getを分けることは同一ではない。

## 2. Reason

1. listのUI更新単位は「1行に配信情報・長さ・進捗印・位置・最終再生日が揃った状態」である。catalog Getとprogress Getを直列にすると、先に長さだけ出て後からbarや日付が載るチラつきが起きやすく、表示timingを揃えたい要件と衝突する。
2. Serverが両storeを待って1応答にすれば、clientのwaterfallと中間描画の分岐が要らない。件数は日次追加規模であり、一覧1発のjoinコストは今の運用想定で許容できる。
3. storeを直交させる理由（書込主体・cache寿命の分離）はHTTPを分割する理由にはならない。BFFが読取時に合成するのは、永続の直交を壊さずにUI単位へ合わせる定石である。
4. 将来catalogがpage化しても、進捗が `episodeId` keyedである限り、そのpageのID集合だけを読んでembedすればよい。embed自体がpagination耐性を奪うわけではない。耐性を壊すのは、進捗をcatalog masterの単一blobへ埋め込む側である（それは先行Decisionが既に採らない）。

## 3. Rejected

1. **progress専用Getをlistと並列で呼び、clientで合流する案** — 往復と描画タイミングが増え、同一timing表示の方が追加実装なしで満たせる今は過剰である。
2. **catalogだけGetし、進捗はbrowser正本で足す案** — 端末横断の同一進捗と矛盾する（先行Decisionの正本がServer側状態行）。
3. **DB直交のためにHTTPも必ず分離する案** — 永続境界と表現境界を混同している。合成はGetの責務として許す。
