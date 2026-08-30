---
name: System の Gemini 認証も暫定で本番 GEMINI_API_KEY を使う
date: 2026-08-30T19:55:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. Generator System workflow の Gemini 認証は、暫定で `secrets.GEMINI_API_KEY`（本番名）を `GEMINI_API_KEY` へ載せる。
2. Cursor は Decision `2026-08-30T17-45-00` どおり本番名のまま。GetX / Google OAuth / Drive は `TEST_*` のまま。
3. `TEST_GEMINI_API_KEY` が 403 無しで通ると確認できたら `TEST_*` へ戻す。

## 2. Reason

1. run 33307390160 で Cursor〜ManuscriptDraft まで成功したあと、`gemini: http_status: status 403` で落ちた。
2. Cursor と同型で、Secret 名（値）の差が原因候補である。

## 3. Rejected

1. Gemini 403 のまま System を止める案 — draft 経路は既に緑で、残る出口検証を遅らせる。
