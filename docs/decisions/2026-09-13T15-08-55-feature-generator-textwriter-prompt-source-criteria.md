---
name: 情報源を P1（Publickey・TechCrunch・クラウド Watch）と P2（HackerNews・Lobsters）へ再編し ITmedia NEWS を廃止する
date: 2026-09-13T15:08:55
branch: feature/generator-textwriter-prompt-source-criteria
---

## 1. Decision

### 1-1. purpose1 / purpose2 の意味（原稿の役割）

1. **purpose1（事実・大きな流れ）**は次を主とする。
   - tech 企業の動向、最新技術、大きな release を **事実ベース・客観的**に紹介する
   - 新しい言葉は解説する
   - 世の中で今起きている大きな流れと、人間全体・日常生活などへの影響を中心にする
   - engineer の意見・議論より **事実の陳列**を優先する
   - ただし engineer 反応は **0 ではない**。主な反応や冷静な分析を **少し**紹介してよい
2. **purpose2（engineer 議論）**は次を主とする。
   - ソフトウェアエンジニアの技術的議論
   - 個別技術への focus、エンジニアリングの未来、プログラミング言語、AI 駆動開発の現在、アーキテクチャ、国際標準策定、研究結果など
   - **今を生きる engineer の熱い議論**を中心にする

### 1-2. どう分けるか（層と源の役割）

1. **分ける単位は「どの Adapter を主戦力にするか」と「TextWriter prompt の書き分け」**である。同一 Fetch 結果を両 purpose に均等利用しない。
2. **`NewFetchSourceItems` は 1 本のまま**とする。源の束を集めるだけであり、P1/P2 の原稿構造を知らない。
3. **`ProduceEpisode` は podcast の P1/P2 分け方を知らない。** Fetch → brief → Write の手順だけを知る。
4. **Port / `SourceItem` に purpose（P1/P2）tag は置かない。** item 単位の振り分けは TextWriter（prompt）が `Summary` / `Detail` / `Discourse` を読んで行う。Adapter が「この源は常に P1」と自己申告するタグも置かない（同一 HN 内でも item により P1 寄り／P2 寄りがあり得るため）。詳細な字段契約は Decision `2026-09-13T15-09-10` を正とする。
5. **TextWriter は 1 本**のままとする。prompt が purpose1 / purpose2 の解説方針と配分、および全体 summary の書き方を知る。prompt 文言の実装は後続（D）とする。
6. **purpose1 の取得源（主戦力）**は Publickey、TechCrunch、Impress クラウド Watch の 3 本とする。契約値（`SourceID`・feed URL・dir）の正本は各 Adapter（A）を参照する。選定方針は「IT を広く取る」ための **high recall（再現率重視）寄り**だが、ITmedia 級の生活・娯楽 noise は避ける（適合率も意識した媒体選定）。
7. **purpose2 の取得源（主戦力）**は HackerNews と Lobsters の 2 本とする（既存 Adapter を維持）。
8. P1 主戦力源が薄い日は、TextWriter が P2 源の item から **P1 寄りの事実**を拾ってよい。逆に P1 源の item を P2 議論の中心にはしない、という運用は prompt（D）側の責務とする。

### 1-3. ITmedia と地図

1. **ITmedia NEWS 経路は廃止する。** `infrastructure/itmedia/` と composition 結線を撤去する。細 feed（`aiplus` / `pcuser` / `news_bursts`）の残置もしない。
2. 本 Decision は先行 Decision（`2026-09-02T14-41-00-feature-hackernews-api-adapter.md`）の「源を HackerNews・Lobsters・ITmedia NEWS の 3 本にする」を **部分 supersede** する。置き換え範囲＝採用源の集合と ITmedia の位置づけ。維持範囲＝認証不要・公式 API/RSS、1 媒体 1 Adapter、RSS 汎用 facade 禁止、GetXAPI 撤去済み、composite は Composition のみ。
3. 地図文書（`DESIGN.md` / `README.md`）の情報取得行は本 Decision に合わせて更新する。源の具体値は Adapter を正とし、地図へ写さない。

## 2. Reason

purpose1 に求める像は「Anthropic / OpenAI / BigTech・大きな release・政策が主で、ガジェットはたまに」である。ITmedia 速報および category union の実測では、キャリア価格比較・生活・娯楽が混ざり precision が低い。Publickey は日本語だが題材は国際技術 release に強く、TechCrunch は米 BigTech / 政策ニュースに強い。日本国内の大手 IT・企業 release の穴は、実測で NTT 系・IIJ・国内企業発表の密度が高かったクラウド Watch で埋める。

purpose2 は議論密度が価値の HackerNews / Lobsters が既に担っており、入れ替えない。

Fetch UseCase を purpose ごとに分けないのは、TextWriter が 1 本で全体 summary を書く前提と、ProduceEpisode が原稿構造を知らない境界を壊さないためである。purpose tag を Port に置かないのは、振り分け知識が Adapter / Application に漏れると層境界が崩れ、かつ同一源内の item 差を表現できないためである。材料の形（Summary / Detail / Discourse）だけ渡して TextWriter が判断する。

ITmedia を細 feed だけ残す案は、母数保険になる一方で purpose1 の精度目標と衝突し、残置理由が「不安のための温存」に近い。完全廃止して 3+2 源に寄せる。

## 3. Rejected

1. **ITmedia を残し category feed を増やして high recall する案** — 件数は増えるが gadget / 生活 noise が増え、求める purpose1 像と合わない。
2. **ITmedia を `aiplus` / `pcuser` / `news_bursts` だけ残す案** — 完全廃止ほどの精度改善にならず、国際テック再掲と雑多が残る。保険として残すコストに見合わない。
3. **日本 IT cover を Publickey だけに任せる案** — Publickey は国内大手・国内政策の系統カバーが弱い（国際技術の日本語解説が主）。
4. **日本 IT cover を IT Leaders またはデジタル庁 RSS を第 1 選択にする案** — IT Leaders はクラウド Watch と同圏で重複しやすい。デジタル庁は政策一次として強いが公募・イベントが混ざり、企業 release 主軸の第 1 本にはしない（追加は未決）。
5. **purpose ごとに `NewFetchSourceItems` を 2 本にする案** — TextWriter 1 本・全体 summary・ProduceEpisode が podcast 詳細を知らない前提と衝突する。
6. **HN/Lobsters だけで purpose1 も賄い報道源を置かない案** — 議論スパイクは来るが、release / 企業発表の定常供給にならない。
7. **`SourceItem` / Adapter に purpose tag を付けて振り分ける案** — 原稿詳細の知識が Port に漏れる。同一源内の item 差も表現しにくい。振り分けは TextWriter + prompt に閉じる。
8. **purpose1 で engineer 反応を完全に 0 にする案** — 事実だけだと「世の中の受け止め」が欠ける。主な反応・冷静な分析を少し入れる。
