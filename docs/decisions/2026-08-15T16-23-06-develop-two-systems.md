---
name: Playback と Generator の2系統に限定する
date: 2026-08-15T16:23:06
branch: develop
---

## 1. Decision

ランタイムは Playback（web + worker）と Generator（Go + GHA）の2系統のみとする。工程ごとの多 package 分割やマイクロサービス化はしない。共有は個人 Google Drive 上のファイルだけとする。

## 2. Reason

個人利用・直交開発・KISS/YAGNI のため。デプロイ単位と所有権が2つに揃い、Clean Arch の多 package 先行は過剰になる。

## 3. Rejected

- 旧 monorepo の工程別 packages を温存しながら漸進移行する案（互換コストが大きく all rewrite と矛盾）
- repo を新規作成する案（Issue / secrets / 文書履歴の二重化）
