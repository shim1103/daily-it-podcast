# lessons

- 2026-09-13 [feature/playback-now-playing-audio-listenability] user の文末が確認問い（？？／「〜しますよ？」）だけのとき、それを execute / edit 許可と読まない。許可は動詞が明示された指示（execute・書け・commit 等）に限る。確認を実装と取り違えると規約違反の無断 edit になる  # → layer:workflow
- 2026-09-13 [feature/playback-now-playing-audio-listenability] 「将来やる方針」と「今 Issue 化する実施契約」を混ぜない。方針だけ固まった topic は Decision（と未決細部の lane）に留め、Acceptance が書けないまま C Issue を起こさない。user が「Decision だけ」と明示したら Issue 化案を押し戻す  # → layer:workflow
- 2026-09-13 [feature/playback-now-playing-audio-listenability] HTTP 境界の共有定数と Infrastructure の配置定数を同一 module から import して揃えようとすると、層ルール（Infra が HTTP contracts を import 禁止）と衝突する。配置の正本は配置契約 doc に置き、HTTP 側と Infra 側はそれぞれがその値を持ち comment で正本を指す  # → layer:terms
- 2026-09-13 [feature/generator-audio-mp3-encode-write] Port 実装（外部 tool / HTTP の話し方）と OS/HTTP 工場は別置き場にする。工場を Adapter 層に置くと「Port を実装する層」と「Client/exec を生成する層」が同名空間で混ざる。Application helper に置けるのは外側 I/O が無い純変換だけ  # → layer:terms
- 2026-09-13 [feature/generator-audio-mp3-encode-write] Composition Root は結線だけにし、production の外側 primitive 工場が厚くなったら独立 package へ出す。結線 package に subprocess 起動実装が居続けると責務が膨れる  # → layer:terms
- 2026-09-13 [feature/generator-audio-mp3-encode-write] UseCase が os/http/exec を直持ちするのは、Composition が同じ道具を Adapter へ inject する政策と矛盾する。inject されていることと UseCase が自分で持つことは同義ではない  # → layer:terms
- 2026-09-13 [feature/generator-wav-to-mp3-port-runtime] package 全体の coverage gate だけ見て新規 unit の到達を宣言しない。巨大 codebase では未計測の新規実装でも aggregate を満たせる。unit 種別ごとの下限（薄い wrapper は到達可能分岐の全通過など）を正とする  # → layer:terms
- 2026-09-13 [feature/generator-wav-to-mp3-port-runtime] Sociable Unit のケース説明は Given/When/Then の構造ラベルを欠かさない。ラベル無しの自由文コメントやケース名だけだと、前提・操作・期待を後から復元できない  # → layer:terms
- 2026-09-13 [feature/generator-wav-to-mp3-port-runtime] coverprofile に現れない定数だけの file は、別実行経路からその定数が使われていれば個別 ignore の対象にしない。計測不能な分岐（環境依存で再現不能な panic 等）だけを局所コメントで切る  # → layer:platform
- 2026-09-13 [feature/generator-wav-to-mp3-port-runtime] 「特定 file の正本は origin/base」と「現在 branch を base へ rebase する」を取り違えない。前者は当該 path の内容同期だけ、後者は履歴全体の載せ替えである  # → layer:workflow
