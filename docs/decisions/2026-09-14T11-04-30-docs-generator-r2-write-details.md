---
name: R2 実施の形は prod/test 別 bucket・flat 同空間・Worker binding 読取・GHA の S3 credential・Drive 継承の公開型書込
date: 2026-09-14T11:04:30
branch: docs/generator-r2-write-details
---

## 1. Decision

1. 先行 Decision `2026-09-13T14-22-55`（正本を R2 へ完全移行）の **実施の形**を本 file が埋める。方針・着手順は先行を正とし、再掲しない。
2. **object 空間**: 本番と test は **別 R2 bucket**。JSON と mp3 は **同一 bucket**。key は flat `{episodeId}.json` / `{episodeId}.mp3`（配置の正本は A の `contracts/episode-layout.md`）。
3. **playback 読取**: Worker の **R2 binding** で bytes を取り、既存 `/episodes*` 経由で返す。公開直 URL / `r2.dev` 直配信は採らない。
4. **credential**: playback は binding のみ（S3 key を Worker に持たない）。generator の実接続は GHA の S3 互換 Access Key ペア（本番は process env 同名、System は `TEST_`）。Account ID・bucket 名・endpoint 式・binding 名は非 Secret。local / Integration gate に実接続 secret を常設しない。
5. **書込意味論**: Drive 時代の公開順（json→音声）・同 stem upsert・不完全ペア許容を維持する（先行 `2026-08-30T23-32-00` / `23-31-00`）。R2 の単 key 強整合を **json+mp3 の multi-object transaction** とみなさない。storage の二次系 fallback / Drive+R2 二重正本は採らない。
6. **失敗方針**: network / 5xx / 429 は有限 retry。その他 4xx は fail-fast。Gemini TextWriter 型の quota secondary は作らない。
7. **storage class**: Standard（無料枠内運用を前提）。Infrequent Access は採らない。
8. **現行 runtime**: Drive のまま。DEPLOY latest・本番結線の R2 差し替えと OAuth 削除は **C の cutover（Issue Acceptance / lane release）** で行う。本 Decision に手順百科を置かない。

## 2. Reason

1. Drive の folder 分離（`DRIVE_FOLDER_ID` / `TEST_*`）と flat ペアリングを bucket に写すと、配置契約と test/prod 隔離がそのまま使える。JSON/mp3 を別 bucket にすると list/get・不完全隠蔽・credential が二重になる。
2. Access が hostname 全体を守る現行（`DEPLOY.md` / `2026-08-25T17-10-00`）と直結 URL は両立しない。Worker binding proxy なら OAuth を消しつつ入場境界を保てる。
3. generator は GHA の Go CLI、playback は Workers なので認証手段が非対称になるのは必然。実接続 secret を GHA に閉じるのは現行 Drive 運用と同型。
4. R2 に multi-object transaction は無い。Drive で既に公開型へ落ちた妥協（`23-32-00`）を「DB だから」と覆す根拠にならない。単 key 強整合・httpMetadata 等は Adapter 実装の得であり、原子性の約束に昇格させない。
5. 日次 Put は数回で、Gemini 型の枠枯渇 secondary が要る負荷ではない。transient だけ粘れば足りる。
6. Standard の無料枠で個人利用は足りる。Infrequent Access は free tier 外と retrieval 課金があり不適。
7. 切替・OAuth 削除・DEPLOY 差し替えは検証可能な完了条件なので Issue / release が持ち、Decision は「いつ持つか」だけを指す。

## 3. Rejected

1. JSON 用・mp3 用の別 bucket — ペアリング契約と注入が割れる。
2. 1 bucket + prefix だけで prod/test を分ける案を必須にする — 誤書込の blast radius が広い。別 bucket を正とする。
3. 公開 `r2.dev` / signed URL 直配信 — Access 外の攻撃面と `r2.dev` rate limit。
4. playback だけ R2・generator は Drive — 先行 `14-22-55` Rejected と同型の二重正本。
5. json+mp3 を transaction / staging 必須にする案 — 能力が無く、layout と Adapter 責任が増える（Drive Rejected と同型）。
6. R2 失敗時の storage secondary（別 bucket / Drive 退避）— 正本が再び割れる。
7. Infrequent Access — 無料枠が無く retrieval 課金が入る。
8. 本 Decision と同時に cutover 手順を Decision 本文へ書く案 — 手順・Verification は C。
