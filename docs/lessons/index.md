# lessons

- 2026-09-22 [refactor/generator-go-design] monorepo 内の lint/format tool の適用範囲は、config の置き場ではなく実行 script が `cd` する app 境界で確認する。未確認のまま「repo 全体」と断定しない  # → layer:platform
- 2026-09-22 [refactor/generator-go-design] Application の配置非対称を「同層だから具象 pointer」と説明すると、内側の Driven Port 差し替え面を無視した誤誘導になる。依存の型は「何を差し替えるか」で選ぶ  # → layer:terms
- 2026-09-22 [refactor/generator-go-design] Port / Composition / UseCase DI の選択は言語非依存の境界設計である。error・context・implicit interface など言語固有の話と混ぜて「言語の思想」として教えない  # → layer:terms
- 2026-09-22 [refactor/generator-go-design] process 終了 API が defer を飛ばす言語では、失敗経路の途中 Exit をやめ、関数の return code を入口一回の Exit に集約して後始末を保証する  # → layer:platform
- 2026-09-22 [refactor/generator-go-design] 公開契約 documentation に処理手順（How）を書くと実装変更で腐る。観測可能な postcondition / invariant だけを残す  # → layer:0:meta
- 2026-09-22 [refactor/generator-go-design] response の thema 分割は自律判断で直交・独立させる。userprompt の見出し構造を写すだけでは分割になっていない  # → layer:0:meta
- 2026-09-22 [refactor/generator-go-design] 言語の強み（例: 軽量並行）が現行仕様の実行形に出ていなくても、その仕様が逐次で足りるなら設計失敗ではない。言語選定の価値と現行 pipeline の形を混同しない  # → layer:terms
- 2026-09-23 [refactor/generator-go-performance] `go get` で追加した依存を実際に import する前に `go mod tidy` を実行すると、未使用依存として go.mod/go.sum から除去される。A（境界契約）として import 宣言だけ先に固定したい場合、`go mod tidy` を挟む操作（lint・test 等の全体実行）の後に依存が残っているか確認する  # → layer:platform
- 2026-09-23 [refactor/generator-go-performance] `.gitignore` へ path を追加しても、その path 配下で既に track 済みの file は自動では unstage されない（`git ls-files` で確認できる）。ignore 方針を追加する commit では、既存 tracked file の棚卸しと `git rm --cached` をセットで行う  # → layer:platform

