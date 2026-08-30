---
name: Access OTP は手動、週次 E2E は storageState、Drive は Worker。PLAYWRIGHT_* の正は DEPLOY
date: 2026-08-30T16:20:03
branch: docs/playback-integration-e2e-plan
---

## 1. Decision

1. Access メール OTP の許可 / 拒否証明は **deploy Phase 2 の手動 Verification**（`DEPLOY.md` §7）。対象は Access 付き **本番 hostname のみ**。preview / version URL は使わない（先行 Decision `2026-08-25T17-10-00` を維持）。
2. 週次 remote E2E は OTP を毎回自動化しない。Playwright **`storageState`（許可 session）** で入場済みとし、同一 origin の一覧・再生・原稿だけを回帰する。Service Token で Access を迂回しない。
3. Google OAuth / Drive は **Worker** の secret 経路。E2E が Google ログインや Drive UI を操作しない。remote E2E の app 経路では `/episodes*` と Drive を double しない。
4. `PLAYWRIGHT_BASE_URL` / storageState 用 Secret の意味・取得・GHA への写像手順の latest は **`DEPLOY.md`** を正とする。本 Decision に手順や値を写さない。

## 2. Reason

1. OTP 取得は Playwright 単体では完結しない。毎週メール依存にすると回帰の赤が Access / メール起因で常態化する。
2. preview URL は Access 外の裏口になり得る。本番 hostname のみの先行方針を破らない。
3. Drive は browser 境界ではない。Worker が refresh_token で読む。E2E が OAuth 画面を触ると所有が壊れる。
4. 登録名・取得手順は運用 SSOT（`DEPLOY.md`）に置き、Decision は方針だけにする（logging/decisions の選別）。

## 3. Rejected

1. 毎 deploy で OTP を自動化する案 — Playwright 単体で完結せず、メール起因の赤が毎回乗る。
2. preview URL で Verification する案 — Access 外の裏口。
3. Service Token で CI から Access を迂回する案 — 先行 Rejected を破る。
4. E2E で Google OAuth / Drive UI を操作する案 — Worker 所有の境界を browser に持ち込む。
5. Decision 本文に Secret 値や取得コマンド全文を正本化する案 — 運用 latest と二重になる。
