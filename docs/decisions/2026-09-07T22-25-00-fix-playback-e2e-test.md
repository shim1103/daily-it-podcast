---
name: playback worker の OAuth consent screen は Production 固定とし、refresh token の 7 日失効を運用から外す
date: 2026-09-07T22:25:00
branch: fix/playback-e2e-test
---

## 1. Decision

`daily-it-podcast` の Google Cloud project で、playback worker が使う OAuth client の **consent screen 公開ステータスを "In production" に固定する**。Testing へ戻さない。`GOOGLE_OAUTH_REFRESH_TOKEN` は Production 化した client で 1 度発行したものを使い続け、失効時のみ再発行する（手順の正は `DEPLOY.md` §3）。

## 2. Reason

1. Google の仕様上、consent screen が **Testing の間に発行された refresh token は 7 日で失効する**。Production では他の失効条件（180 日未使用・手動 revoke・同一 client/user/scope での過剰発行）に触れない限り無期限。
2. playback worker は `fetchAccessToken()` で毎リクエスト refresh token → access token を取り直す。refresh token が失効すると token endpoint が `400 invalid_grant` を返し、`/episodes` が `503` になる。週次 `playback-e2e` はこの経路を必ず通るため、Testing のままだと **7 日ごとに e2e が落ちる**。実際 2026-08-31 の schedule 成功から 2026-09-06 の失敗まで約 7 日で、Testing 放置が原因だった。
3. daily-it-podcast は許可 identity が自分 1 件の個人利用 podcast（`DEPLOY.md` §2）。`drive.readonly` は sensitive scope だが、Google の verification 審査は「自分だけが使う unverified app」なら不要で、consent 時の警告画面を自分で通過すれば足りる。審査コストを負わずに Production 化できる。

## 3. Rejected

1. **Testing のまま運用し、7 日ごとに refresh token を手動再発行する** — 週次 e2e より短い周期で人手の再認可が必要になり、e2e が「token が生きているか」の当落だけで赤緑を繰り返す。検証したい playback の回帰が token 失効ノイズに埋もれる。
2. **e2e の失敗を許容し、落ちたら都度 token を差し替える** — 週次 gate が常時赤だと「赤は想定内」の慣れが生まれ、本物の回帰を見逃す。gate の意味が失われる。
3. **Drive + OAuth をやめ R2 へ移して refresh token 自体を無くす** — 失効する部品を消せる根本策で有力だが、playback worker と generator の両方に跨る中規模の移行で本 Decision の射程を超える。別軸として `docs/tasks/todo/playback-lane.md` の未決 index へ送る。本 Decision は「Drive + OAuth を使い続ける前提で、consent screen をどう置くか」だけを固定する。
