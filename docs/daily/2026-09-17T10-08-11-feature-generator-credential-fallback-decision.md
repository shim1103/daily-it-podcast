---
name: Gemini/Cursor credential役割区分のdecisionを記録
date: 2026-09-17T10:08:11
session_id: none
branch: feature/generator-credential-fallback-decision
prev: なし
---

## 1. Summary

TextWriter/TTSのmodel切り替えfallback実装（後続PR）に先立ち、Gemini/Cursor credentialをどのworkflowがどの順序で使うかを固定するdecisionを1件追加した。元のfeature/generator-textwriter-adapter-fallback branchで作った草案から、develop側で既に完了していたGoogle Drive/OAuth撤去（R2完全移行）と矛盾する記述を除去し、develop向けのstacked PRの1番目として切り出した。

## 2. Changes

1. `GEMINI_API_KEY`/`TEST_GEMINI_API_KEY`/`SPARE_GEMINI_API_KEY`の役割3分割と、`SPARE_GEMINI_API_KEY`を全workflow共通のfinal-fallbackにする判断を記録
2. `TEST_`接頭辞の付与基準を「値が本番と異なるcredential/variableにのみ付ける」に統一し、値が同一な`TEST_CURSOR_API_KEY`等の廃止方針を記録
3. 元decision草案にあったGoogle Drive/OAuth credential（`GOOGLE_OAUTH_*`/`DRIVE_FOLDER_ID`）関連の記述を、develop側で既に撤去済み（`e86c831`）だったため除去し「対象外」と明記
4. PR #176としてdevelop向けstacked PRを作成、push・CI確認まで完了

### Commits

- `4d072b4`
