---
name: Fetch 窓を実行時刻の rolling 24h から表示 TZ の昨日 half-open 暦日へ変える
date: 2026-09-29T02:10:38
branch: cursor/gemini-structured-yesterday-window-6a75
---

## 1. Decision

1. Source 取得窓は **表示 Location（`DisplayTimeZone` = Asia/Tokyo）の昨日暦日** とする。区間は half-open `[昨日 00:00, 今日 00:00)`。
2. 窓の算出は純関数 `constants.YesterdayHalfOpenWindow(now, loc)` に閉じる。`time.Now` を呼ばない。
3. `FetchSourceItems` は `List(ctx, since)` のあと Application で `OccurredAt ∈ [since, until)` を filter する。Port `ItemSource.List` の signature は変えない。
4. 日次 cron（JST 05:00）は変えない。episode 表示 `date` は従来どおり実行日の JST 暦日。

## 2. Reason

1. rolling `now - 24h` は cron 遅延や再実行で「どの日のニュースか」がずれる。昨日固定なら同じ暦日を何度実行しても同じ窓になる。
2. Port に `until` を足すと全 Adapter の契約・test を同時に動かす。Application filter なら Adapter は既存 `since` 下限のまま、窓の意味は UseCase に集約できる。
3. 純関数に `now` と `loc` を渡す形は既存の `displayDate` と同型で、Sociable Unit が固定時刻だけで境界を突ける。

## 3. Rejected

1. **cron 時刻を真夜中直後へずらして窓を揃える案** — 運用時刻を変えると既存 schedule・容量計画が動く。要求は「実行時刻は据え置き」。
2. **`List(ctx, since, until)` へ Port 拡張する案** — 正しさは上がるが全 Adapter 変更が要る。Application filter で契約を満たせるうちは YAGNI。
3. **`FetchWindow = 24h` を残して since だけ昨日 00:00 にする案** — until 無しだと今日 00:00〜実行時刻の item が混ざる。half-open の両端が必要。
