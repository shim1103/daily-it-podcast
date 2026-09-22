---
name: 進捗Write失敗はuser向け通知せずconsole.errorに留め有限retryしsession終了で捨てる
date: 2026-09-19T19:12:30
branch: docs/playback-audio-history
---

## 1. Decision

1. 進捗の create / update（push）/ complete が失敗しても、再生そのものや一覧の通常操作を止めない。user向けの同期失敗UI（toast・バナー・文言）は出さない。
2. 失敗は開発者向けに `console.error` 程度で残し、有限回の retry を試みてよい。browser session が終わったら in-memory の retry は破棄してよい（session またぎの永続送信キューは持たない）。
3. 進捗UIの共有状態は、Write成功後に更新する前提と両立させる（失敗中に共有進捗印を確定表示しない）。音の再生開始は進捗Writeの完了を待たない。

## 2. Reason

1. 本アプリの主価値は音声を聞くことであり、進捗同期は端末横断の補助である。同期失敗をuserに説明しても操作の選択肢が増えず、知らなくてよい概念をUIに出すことになる。
2. Access配下の個人利用では、失敗を握りつぶして次の play / stop / complete や次sessionの操作・pull に再同期を任せても実害が小さい。先勝ち／後勝ちのmergeがあるため、再送や再完了は冪等に寄せられる。
3. sessionまたぎの確実配信キューは、永続・衝突・プライバシー面の設計が増え、今の「成功後UI・黙ってretry」より重い。
4. 完全無言だと障害調査ができないため、browserの `console.error` に失敗を残す。user向けUXは増やさない。

## 3. Rejected

1. **同期失敗をtoast等でuserに通知する案** — 主操作を妨げず得も少ない。同期という概念をUIに露出する。
2. **失敗したら再生も止める／playを禁止する案** — ローカルで聞ける体験を同期の成否に結合させ、主価値を落とす。
3. **IndexedDB等に未送信を貯めsessionまたぎで必ず送る案** — 今の規模と失敗許容度に対して過剰である。
4. **失敗をログにも残さない案** — 握りつぶしと障害切り分けが両立しない。
