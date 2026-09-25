# lessons

- 2026-09-22 [refactor/generator-go-design] monorepo 内の lint/format tool の適用範囲は、config の置き場ではなく実行 script が `cd` する app 境界で確認する。未確認のまま「repo 全体」と断定しない  # → layer:platform
- 2026-09-22 [refactor/generator-go-design] Application の配置非対称を「同層だから具象 pointer」と説明すると、内側の Driven Port 差し替え面を無視した誤誘導になる。依存の型は「何を差し替えるか」で選ぶ  # → layer:terms
- 2026-09-22 [refactor/generator-go-design] Port / Composition / UseCase DI の選択は言語非依存の境界設計である。error・context・implicit interface など言語固有の話と混ぜて「言語の思想」として教えない  # → layer:terms
- 2026-09-22 [refactor/generator-go-design] process 終了 API が defer を飛ばす言語では、失敗経路の途中 Exit をやめ、関数の return code を入口一回の Exit に集約して後始末を保証する  # → layer:platform
- 2026-09-22 [refactor/generator-go-design] 公開契約 documentation に処理手順（How）を書くと実装変更で腐る。観測可能な postcondition / invariant だけを残す  # → layer:0:meta
- 2026-09-22 [refactor/generator-go-design] response の thema 分割は自律判断で直交・独立させる。userprompt の見出し構造を写すだけでは分割になっていない  # → layer:0:meta
- 2026-09-22 [refactor/generator-go-design] 言語の強み（例: 軽量並行）が現行仕様の実行形に出ていなくても、その仕様が逐次で足りるなら設計失敗ではない。言語選定の価値と現行 pipeline の形を混同しない  # → layer:terms
- 2026-09-25 [develop] `git branch --merged` は tip commit が base の祖先かを見る。機能が base に在っても tip hash が違うと unmerged と出る  # → layer:platform
- 2026-09-25 [develop] squash merge は複数 tip commit を別 hash の1 commit へ載せ替える。元 tip は base 祖先にならず、`--merged` では残る  # → layer:platform
- 2026-09-25 [develop] 間接 merge（squash・cherry-pick・別 tip での再実装）済みの feature tip は、機能の正本が base にあるなら残置理由にならない  # → layer:platform
- 2026-09-25 [develop] remote branch 削除は名前付き ref を外す操作である。他 branch / tag / merge 祖先から到達できる commit は残る  # → layer:platform
- 2026-09-25 [develop] どの ref からも辿れない tip は orphan（dangling）になりやすい。SHA を知れば一時的に戻せる場合もあるが、保証は弱い  # → layer:platform
- 2026-09-25 [develop] hosting の「履歴が見える」ことは未 merge tip を残す理由にならない。進行中 work の正本参照は branch / tag / PR tip のどれかで明示する  # → layer:platform
- 2026-09-25 [develop] branch の残置基準は「いま必要な参照か」だけ。残すのは保護ブランチと進行中の未完了 work。release 済み・直接/間接 merge 済み・再開しない tip は残置しない  # → layer:terms
