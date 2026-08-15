---
name: Drive 契約は episodeId と表示用 date を分ける
date: 2026-08-15T16:23:07
branch: develop
---

## 1. Decision

原稿 JSON とファイル stem の対応キーは不透明な `episodeId` とする。表示用日付は Generator が Asia/Tokyo で暦日化した `date`（`YYYY-MM-DD`）を JSON に書く。UI は timestamp からの切り出しをしない。契約の正は `contracts/` に置く。

## 2. Reason

ISO timestamp を id・日付・ファイル名に兼ねると、UI が Generator の時刻組み立てを知ることになり直交性が壊れる。表示日付の決定は書込側に閉じる。

## 3. Rejected

- UI や共有 utils で timestamp から日付を切り出す案
- 認証や `DRIVE_FOLDER_ID` を contracts に含める案（実行設定／Adapter に閉じる）
