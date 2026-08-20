---
name: playback web API clientの応答責務分割とtest境界修正
date: 2026-08-20T19:27:00
session_id: none
branch: issue/playback-web-api-client
prev: 2026-08-20T14-18-45-playback-web-api-client.md
---

## 1. Summary

`playback-api-client` のHTTP応答処理を責務ごとに分割し、success判定・契約error mapping・web固有error mappingの境界を固定した。あわせて、response処理とclient配線のtest scopeを分離し、外部mapperをdoubleにした伝播検証へ整理した。

## 2. Changes

1. `Response.ok` をsuccess判定の正規rootへ変更し、失敗responseのbodyを読まないresponse moduleを追加した
2. shared contractは既知の契約error codeだけを変換し、未知statusのweb error mappingをweb側へ移した
3. client testからJSON・schema・Blob・status・networkのresponse意味論を除き、URL・委譲・Result伝播だけを検証した
4. response testへ正常系・異常系・境界系を追加し、mapper Stubの戻り値伝播、JSON read failure、schema failure、Blob failure、failed body未読を検証した
5. executor実装をmanagerが独立検証し、review指摘のJSON parse failure test不足を追加修正した
6. `npm run test:unit` は `122 passed`、`typecheck`、`lint`、`format:check` はpassした

### Commits

1. `c512a97` — API response責務を分離する
2. `dd716bb` — API境界のtest責務を分ける
