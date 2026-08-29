---
name: HTTP vendor 4本の SU/NI latest 完了と PR 準備
date: 2026-08-30T03:13:00
session_id: none
branch: feature/generator-su-ni
prev: なし
---

## 1. Summary

getxapi / oauth / gemini / gdrive の Sociable Unit と Narrow Integration を `*http.Client` + capability Config 形へ揃え、境界 I/O と Adapter 内分岐の責務を分離した。oauth / getxapi は Narrow 新規、gemini は RoundTripper Spy へ降格していた Narrow を httptest TLS + DialTLS へ復元、gdrive は SU と Narrow の二重観測を解消した。達成契約 4 file を削除し lane を完了へ更新した。

## 2. Changes

- manager は non-edit。executor / reviewer へ委譲して実装・再 review・AC audit まで完了。
- Narrow 成功 path の最小観測は「動詞 + 認証 header + 成功成果物」の 3 点に固定（Ask で need/sufficient を確認）。
- gemini Narrow の RoundTripper Spy 化は levels 上の Narrow 定義違反として review で BLOCK、TLS redirect へ戻した。
- pre-commit は playback の Node 22 + biome が必要。worktree で `npm install` し、ついでに package-lock の libc フィールド差分を chore commit。
- GitHub Issue は無し（local task file が契約の正）。

### Commits

- `273a976`
- `d75f770`
- `b44caaf`
- `1592fc2`
- `ba2e723`
- `6565682`
