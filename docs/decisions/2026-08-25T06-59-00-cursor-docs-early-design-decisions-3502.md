---
name: Playback UI 入場は Cloudflare Access。許可 identity は自分のみ
date: 2026-08-25T06:59:00
branch: cursor/docs-early-design-decisions-3502
---

## 1. Decision

1. Playback UI の入場制御は Cloudflare Access（メール OTP）とする。
2. 許可 identity は自分の email のみとする。アプリ内マルチテナント OAuth は作らない。
3. Drive の長期 credential はブラウザに置かない。worker / generator の Infrastructure が持つ（`DESIGN.md` §4 と同旨。手順の詳細はここに書かない）。

## 2. Reason

1. 要求は「自分以外には聴かせない」だけであり、組織ディレクトリや社内 IdP は前提にしない。
2. Access を入場に使い、Drive OAuth を読取代理に閉じると、入場と保存の秘密が混ざらない。

## 3. Rejected

1. アプリ内ユーザー DB・セッションで入場を自作する案
2. Drive の OAuth をブラウザへ直接載せて入場を兼ねる案
3. 不特定多数向けの公開 Playback にする案
