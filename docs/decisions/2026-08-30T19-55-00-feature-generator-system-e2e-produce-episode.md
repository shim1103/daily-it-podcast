---
name: System の Gemini 認証も暫定で本番 GEMINI_API_KEY を使う
date: 2026-08-30T19:55:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. Generator System workflow の Gemini 認証は、暫定で `secrets.GEMINI_API_KEY`（本番名）を載せる。
2. Cursor の暫定本番採用（Decision 2026-08-30T17-45-00）と並べ、GetX / Google OAuth / Drive は `TEST_*` のまま。
3. `TEST_GEMINI_API_KEY` が TTS 可能な同値と確認できたら `TEST_*` へ戻す。

## 2. Reason

1. run 33307390160 で Cursor draft まで成功したあと、`gemini: http_status: status 403` で落ちた。
2. Cursor と同様、TEST_ key が無効または権限不足である可能性が高い。probe 再現の最短経路は本番名 Secret。

## 3. Rejected

1. Gemini 403 のまま System を止める案 — draft 経路は既に緑で、次の切り分け材料が揃っている。
