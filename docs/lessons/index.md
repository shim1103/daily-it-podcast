# lessons

- 2026-09-13 [feature/playback-now-playing-audio-listenability] user の文末が確認問い（？？／「〜しますよ？」）だけのとき、それを execute / edit 許可と読まない。許可は動詞が明示された指示（execute・書け・commit 等）に限る。確認を実装と取り違えると規約違反の無断 edit になる  # → layer:workflow
- 2026-09-13 [feature/playback-now-playing-audio-listenability] 「将来やる方針」と「今 Issue 化する実施契約」を混ぜない。方針だけ固まった topic は Decision（と未決細部の lane）に留め、Acceptance が書けないまま C Issue を起こさない。user が「Decision だけ」と明示したら Issue 化案を押し戻す  # → layer:workflow
- 2026-09-13 [feature/playback-now-playing-audio-listenability] HTTP 境界の共有定数と Infrastructure の配置定数を同一 module から import して揃えようとすると、層ルール（Infra が HTTP contracts を import 禁止）と衝突する。配置の正本は配置契約 doc に置き、HTTP 側と Infra 側はそれぞれがその値を持ち comment で正本を指す  # → layer:terms
- 2026-09-13 [feature/generator-textwriter-prompt-source-criteria] Decision Record に raw→field 対応表や契約字段の百科を正本として置かない。方針・Rejected だけを Decision に残し、表は実装前なら Issue・実装後なら Adapter code を正本にする  # → layer:workflow
- 2026-09-13 [feature/generator-textwriter-prompt-source-criteria] Purpose 別の「空＝その Purpose 経路なし」判定を使うなら、その Purpose 用の URL も同じ側の構造に置く。帰属用の袋へ URL をまとめると、本文が空のとき経路が無いと誤読される  # → layer:terms
- 2026-09-13 [feature/generator-textwriter-prompt-source-criteria] 「独立して完了・検証できる境界」は Issue を 1 本にしろという強制ではない。同型の繰り返し実装は 1 本にまとめてよく、媒体ごとに独立完了できるなら分割も妥当  # → layer:workflow
- 2026-09-13 [feature/generator-textwriter-prompt-source-criteria] ある path の正本が別 branch にある、という指示を、作業 branch 全体の rebase 指示と取り違えない。正本取り込みの単位は指定された artifact に限定する  # → layer:workflow
