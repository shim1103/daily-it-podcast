---
name: SNS 取得は TwitterAPI.io（試作）/ GetXAPI（本運用）を採用
date: 2026-08-15T16:39:20
branch: feature/x-api-adoption
---

## 1. Decision

Generator の情報取得源の一つとして、非公式 X API を採用する。

- **試作段階**: TwitterAPI.io
- **本運用**: GetXAPI
- 用途: 個人利用・取得のみ（write 無し）
- 頻度: 1 set/day（cron 駆動）
- 1 set: 20〜100 user、各 user の直近 24h（`FetchWindow`）
- 当面の取得対象: オリジナル投稿のみ。Reply / Repost / 引用は後回し（`2026-08-15T17-43-09`）
- 想定 volume: 6,000〜90,000 tweet/月
- 対象 account は定数 file に人物 id のみ集約（`WatchUserIDs`）
- 戻り形の正: `entities/models.Post` と `application/port.PostSource` の契約
- API key・token は secret 管理（試作 env 名: `TWITTERAPI_IO_API_KEY`）
- 重複排除 key: `Post.ID`（tweet id）
- 取得用 X account は本 account と分離（BAN 対策）
- 流れ: cron(1/day) → fetch → idempotent upsert(DB) → media local 保存 → log（retry 付き・exponential backoff）。upsert 以降は別 task
- log・失敗通知は最低限実装
- 多重契約による fallback は設計しない

想定月額（Mid: 36K tweet/月）: GetXAPI $1.80 / TwitterAPI.io $5.40（trial 消化後）

## 2. Reason

- TwitterAPI.io: $1 trial credit が厚く、数千 call 試せる。契約前に endpoint 仕様・RT 取得完全性を実測できる
- GetXAPI: $0.05/1K tweet で最安クラス、公式 X API v2 と response schema 互換、sign in 即 API key 発行で setup 最軽
- 全第三者 API は reverse-engineered proxy であり、X frontend 変更時は同時停止 risk。多層 fallback は成立しない

## 3. Rejected

- **公式 X API**: 2026-02 に pay-per-use 化、$0.005/read。shim scale で第三者比約 100 倍高。2M read/月 hard cap 超過は Enterprise $42K+/月。free tier 廃止。write 不要な用途で公式を選ぶ利点なし
- **Sorsa**: 単価 $0.02/1K で最安だが、batch endpoint が「特定 user の直近 24h tweet 取得」に対応するか未確認。仕様適合が確定したら再検討
- **Apify**: $0.40/1K tweet で第三者中最高価、High scenario で $36/月。browser-based scraper で X frontend 変更に脆弱、2025 年に人気 actor が 3 回 multi-day outage。actor run 課金で cost 予測困難
- **TweetAPI / Zernio（subscription 系）**: 月 $17〜19 固定。低 volume 用途で pay-per-use より不利。使わない月も課金発生
- **第三者 API + scraper fallback 多層構成**: 両者は内部実装が同種（reverse-engineered）で相関 failure する。X 側変更で同時停止し fallback が成立しない。真の冗長化には別種 route（RSS bridge 等）が必要だが個人利用では over-engineering
