---
name: 進捗の frontend A は API client で止め、compose／UI 公開面は構成 B 後
date: 2026-09-25T05:53:19
branch: docs/playback-audio-history
---

## 1. Decision

1. 進捗の **frontend A は API client 面で止める**（create／update／complete／pull の signature・stub・足場）。UI／hooks／page compose の公開面は、本 Decision では凍らせない。
2. Feature（Row／Item）は **HTTP と progress derive を持たない**。渡された表示 props だけを描く。配置の**意味**（日付右＝完走印、下段＝位置＋`lastPlayedAt`）は方針として固定する。色・寸法・bar・`style`・slot 埋めは A にしない。
3. **hooks の file 構成・新 hook の採否・page／compose の公開戻り**は未確定。catalog 寿命と progress 寿命が一致しない疑いがある間は、今の単一 compose へ進捗を押し込む形を A／B の確定形にしない。
4. UI 信号定数を A に置くかは、構成の話し合いのあと都度判断する。

## 2. Reason

1. API client だけで後続の同期実装は進められる。UI 公開面まで A に広げると、未確定の hook 分割を仮 compose に合わせて固定し、**悪い合わせ**になる。
2. Feature に HTTP／derive を載せると frontend 層境界が崩れ、同じ表示判断が分散する。
3. 見た目 polish まで A にすると design 未確定のまま実装を先取りし、C と二重になる。
4. 「既存 VM に寄せる／新 hook を切る」は構成の選択であり、寿命・責務が揃うまで決めない。先に寄せると分割の余地を潰す。

## 3. Rejected

1. **page／compose hook の公開戻りや Item／Row の progress props を、構成 B 前に A で凍らせる** — catalog≠progress 寿命の疑いがあるのに仮合わせを契約化する。
2. **進捗同期専用 hook を、構成未確定のまま採否確定する**（切る／切らないの両方） — 所有権の材料が揃う前に択一する。
3. **Row が progress API／derive を持つ** — Feature が API Client を知り、表示契約の正本が分散する。
4. **進捗印・bar の見た目まで A で完成させる** — polish を契約に混ぜ、C が空／二重になる。
5. **UI／components／hooks の file 構成を本 Decision で一括確定する** — 話し合っていない範囲を含めない。
