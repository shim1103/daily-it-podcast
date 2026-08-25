---
name: Playback は同一 origin と Access メール OTP と Worker secret だけで守る
date: 2026-08-25T17:10:00
branch: feature/playback-worker-deploy
---

## 1. Decision

1. 公開配信は **同一 origin** とする。静的 UI と `/episodes*` API を 1 つの Worker hostname に載せる。
2. hostname は `*.workers.dev` とする。custom domain は使わない。
3. Cloudflare Access は **メール OTP**、許可 identity は **自分の email 1 つ**、対象は本番 hostname **全体**、session は **30 日**とする。
4. app 内で Access JWT を再検証しない。Service Token / WARP / 追加 IdP は使わない。
5. 利用・共有してよい URL は Access 対象の **本番 hostname のみ**とする。preview / version URL は使わない・公開しない。
6. Playback Worker の `PlaybackEnv` 4 key はすべて Workers **secret** とする。production で in-memory repository mode は使わない。
7. 公開 URL を出す前に Access Application と Allow ポリシーを用意する。
8. 上記の最新値・Verification 観点の正本は `DEPLOY.md` とする。wrangler の具体 config 値の正本は `apps/playback/wrangler.jsonc` とする。

## 2. Reason

1. Web の production `baseUrl` は既に `""`（相対 path）である。別 origin（Pages UI + 別 Worker API）にすると Access cookie / CORS / credential の設計が必要になり、個人利用の脅威モデルに対して過剰である。
2. Workers Static Assets で同一 Worker が静的 file と API を出せば、Access は hostname 1 つで足りる。hash routing のため、未知 path を全部 `index.html` へ落とす厚い SPA fallback は不要で、`/episodes*` 先回りだけで衝突を避けられる。
3. custom domain は DNS・証明書・Access Application の追加作業だけが増え、初回の手動 deploy 前に必須の価値がない（YAGNI）。
4. 個人・非公開・自分だけが聴く前提では、外部 IdP や WARP なしのメール OTP で入場を閉じれば足りる。許可 email を 1 つに固定するとポリシーが単純で、誤って広い Allow を置きにくい。
5. session 30 日は再 OTP の手間と、端末共有・cookie 漏洩時の窓の折衷である。短すぎると毎回 OTP、長すぎると紛失端末のリスクが伸びる。個人利用では 30 日を採用する。
6. Access が edge で未認証を止めるため、Worker 内 JWT 再検証は同じ脅威に対する二度目の実装になる。JWKS・issuer・audience の保守だけが増える（YAGNI）。Service Token は browser 以外の呼び出し用であり、今の呼び出し手は自分の browser だけである。
7. Access Application を本番 hostname だけに付けると、preview / version URL が Access 外の裏口になり得る。初回は本番 hostname 以外を使わない方針で窓を閉じる。
8. OAuth 3 値と `DRIVE_FOLDER_ID` を分けて vars にすると、「非 secret」扱いの誤設定・dashboard 露出面が増える。4 つとも secret に揃える方が注入経路が一つで、production の欠落も同じ検証に乗る。in-memory は local / unit 専用であり、production 相当の欠落をごまかさない。
9. Worker を先に公開してから Access を付けると、Drive 代理 API が一瞬でも無防備になる。Application と Allow を先に用意する。

## 3. Rejected

1. Pages（UI）+ 別 Worker（API）— 別 origin。Access 二重または CORS が要る。`baseUrl: ""` と矛盾する。
2. custom domain を初回から入れる案 — DNS / 証明書 / Access 対象の追加だけが増える。
3. app 内 Access JWT 検証 — edge Access と重複し、JWKS 保守が増える。
4. Service Token / WARP / 追加 IdP — 呼び出し手が自分の browser だけなので不要。
5. `DRIVE_FOLDER_ID` だけ vars、OAuth だけ secret — 注入区分が割れ、誤って非 secret 扱いしやすい。
6. production で env 欠落時に in-memory へ落とす案 — 設定不備が黙って Fake になり、観測できない。
7. preview / version URL を Access 外のまま使う案 — 本番以外の裏口になる。
8. Worker `name` に `playback` を足す案 — 識別に不要。正本は `wrangler.jsonc` の `name`。
