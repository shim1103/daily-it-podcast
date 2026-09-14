---
name: httpget・controllable peer NI・接続 cache・runtime 図更新
date: 2026-09-14T16:07:52
session_id: none
branch: feature/generator-item-source-httpget-ni-cache
prev: なし
---

## 1. Summary

`httpget` helper と 5 Adapter 委譲、源 NI（controllable peer）、gate 外接続 cache、runtime 図の 5 源/mp3 所有修正までをこの branch で完了した。途中で「本番 GET=Narrow」誤読を Decision `15-05-00` で置き換えた。

## 2. Changes

1. issue-manager で helper / Adapter / NI / cache を実装後、shim 指摘で NI を httptest+DialTLS に戻し GWT を全 case へ揃えた
2. connection cache suite を明示 tag 実行して `.cache/` に 5 JSON を保存確認
3. runtime 図を 5 源・Go CLI 所有の save・ffmpeg encode 呼び出しに修正（R2/cache node は非 scope）
4. Decision `15-05-00`・DESIGN・lane・lessons を同期

### Commits

- `69c7262`
- `415939b`
- `6854c17`
- `4ba3606`
- `a5740b1`
- `26e9a64`
- `12d1842`
- `afc63e2`
