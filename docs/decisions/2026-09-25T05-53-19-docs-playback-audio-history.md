---
name: 進捗の同期と表示導出は既存VMに寄せ、RowはHTTPとderiveを持たない
date: 2026-09-25T05:53:19
branch: docs/playback-audio-history
---

## 1. Decision

1. 進捗の push／pull・retry 呼び出しは **新 hook を切らず**、既存の list／playback ViewModel 面に寄せる。公開戻り・client method の正本は A。
2. 進捗の **表示導出**は ViewModel（`playback-state` の公開 derive）が所有する。Feature（Row／Item）は derive せず、渡された表示 props だけを描く。HTTP も持たない。
3. 配置の意味（日付右＝完走印、下段＝位置＋`lastPlayedAt`）は方針として固定する。色・寸法・bar 塗り・`style`・slot の中身埋めは A にしない。
4. **UI／components／hooks の file 構成全体は、本 Decision では確定しない。** UI信号定数を A に置くかは、構成の話し合いのあと都度判断する。

## 2. Reason

1. 同期を独立 hook に切ると、list の mount 寿命・playback phase・同一 ApiClient との配線が二系統になり、既存 VM に閉じるより境界が増える。
2. Feature に HTTP や derive を載せると frontend の層境界（表示単位 vs 状態・API）が崩れ、同じ表示判断を Row と VM で繰り返す。
3. 見た目 polish まで A にすると design 未確定のまま実装を先取りし、C の描画 Issue と二重になる。
4. file 構成を話し合わずに「定数は A」と固定すると、置き場の判断を先取りする。

## 3. Rejected

1. **進捗同期専用の新 hook（例: `useProgressSync`）** — 同期入口が分裂し、既存 list／playback 寿命との接続点が二系統になる。
2. **Row が progress API／derive を持つ** — Feature が API Client を知り、表示契約の正本が分散する。
3. **進捗印・bar の見た目まで A で完成させる** — design 未確定の polish を契約に混ぜ、C の描画 Issue が空になる／二重になる。
4. **UI／components／hooks の file 構成を、本 Decision で一括確定する** — まだ話し合っていない範囲を契約に含めない。
