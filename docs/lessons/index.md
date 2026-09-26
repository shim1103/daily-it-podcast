# lessons

- 2026-09-22 [refactor/generator-go-design] monorepo 内の lint/format tool の適用範囲は、config の置き場ではなく実行 script が `cd` する app 境界で確認する。未確認のまま「repo 全体」と断定しない  # → layer:platform
- 2026-09-22 [refactor/generator-go-design] Application の配置非対称を「同層だから具象 pointer」と説明すると、内側の Driven Port 差し替え面を無視した誤誘導になる。依存の型は「何を差し替えるか」で選ぶ  # → layer:terms
- 2026-09-22 [refactor/generator-go-design] Port / Composition / UseCase DI の選択は言語非依存の境界設計である。error・context・implicit interface など言語固有の話と混ぜて「言語の思想」として教えない  # → layer:terms
- 2026-09-22 [refactor/generator-go-design] process 終了 API が defer を飛ばす言語では、失敗経路の途中 Exit をやめ、関数の return code を入口一回の Exit に集約して後始末を保証する  # → layer:platform
- 2026-09-22 [refactor/generator-go-design] 公開契約 documentation に処理手順（How）を書くと実装変更で腐る。観測可能な postcondition / invariant だけを残す  # → layer:0:meta
- 2026-09-22 [refactor/generator-go-design] response の thema 分割は自律判断で直交・独立させる。userprompt の見出し構造を写すだけでは分割になっていない  # → layer:0:meta
- 2026-09-22 [refactor/generator-go-design] 言語の強み（例: 軽量並行）が現行仕様の実行形に出ていなくても、その仕様が逐次で足りるなら設計失敗ではない。言語選定の価値と現行 pipeline の形を混同しない  # → layer:terms
- 2026-09-25 [feature/playback-progress] frontendの契約凍結で、文字・印・CSSを埋めた描画を「slot／props／配置」と呼んで残さない。空のdata属性と型・定数宣言までが契約、中身埋めは実装  # → layer:3:workflow
- 2026-09-25 [feature/playback-progress] 「完成系に寄せる」はsignature・置き場・入口までであり、動くUIや見た目の完成を先取りする根拠にしない  # → layer:3:workflow
- 2026-09-25 [feature/playback-progress] UI信号の定数宣言が契約でも、定数をDOMに描くことまで契約に含めない。宣言と描画を分けて消す  # → layer:3:workflow
