# lessons

- 2026-09-13 [feature/playback-now-playing-audio-listenability] user の文末が確認問い（？？／「〜しますよ？」）だけのとき、それを execute / edit 許可と読まない。許可は動詞が明示された指示（execute・書け・commit 等）に限る。確認を実装と取り違えると規約違反の無断 edit になる  # → layer:workflow
- 2026-09-13 [feature/playback-now-playing-audio-listenability] 「将来やる方針」と「今 Issue 化する実施契約」を混ぜない。方針だけ固まった topic は Decision（と未決細部の lane）に留め、Acceptance が書けないまま C Issue を起こさない。user が「Decision だけ」と明示したら Issue 化案を押し戻す  # → layer:workflow
- 2026-09-13 [feature/playback-now-playing-audio-listenability] HTTP 境界の共有定数と Infrastructure の配置定数を同一 module から import して揃えようとすると、層ルール（Infra が HTTP contracts を import 禁止）と衝突する。配置の正本は配置契約 doc に置き、HTTP 側と Infra 側はそれぞれがその値を持ち comment で正本を指す  # → layer:terms
