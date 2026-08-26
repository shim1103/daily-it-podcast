# lessons

177：2026-08-26 [feature/generator-agentsecrets-http-proxy-absorb] 同じ置き場名でも出口契約が違うなら runtime package を出口ごとに分ける。見た目の重複を DRY 違反とみなし1袋へ畳むと Orthogonality が壊れる  # → layer:terms
178：2026-08-26 [feature/generator-agentsecrets-http-proxy-absorb] path の移動は契約吸収ではない。所有境界の固定と、出口契約を満たす単一実装への再設計を同一の完了条件にしない  # → layer:terms
179：2026-08-26 [feature/generator-agentsecrets-http-proxy-absorb] default 接続先の test は dial 失敗を証拠にしない。差し込んだ transport が受け取った URL を assert し、port 占有や到達可否に依存しない  # → layer:terms
180：2026-08-26 [feature/generator-agentsecrets-http-proxy-absorb] 削除 tool の承認 UI の表示対象と実際の削除 path を混同しない。承認前に対象 path を確認し、無関係な lockfile と同一視しない  # → layer:workflow
181：2026-08-26 [docs/migrate-lessons-2] goalに完了条件（例：特定fileが空であること）が明記されている時、同時に書かれたrole宣言（non-edit等）はreview・planへの権限制限であり、goal自体の完遂を差し止める指示ではない。role宣言を理由にgoalの完了条件を満たさないまま止めない  # → layer:workflow
182：2026-08-26 [docs/migrate-lessons-2] 複数runtimeへ同一知識をsymlink展開する構成で、SSOT側に実体が存在しないことだけを見て新設と判断しない。展開先の各runtimeを横断的に確認し、どこか1つにのみ実体が存在する非対称配置（linkは追加されたが実体の一部runtimeへの反映が漏れていた等）を見落とすと、既存知識を上書きして消失させる  # → layer:workflow
183：2026-08-26 [feature/generator-cmd-entrypoint] 公開境界の契約 documentation は entrypoint と置換可能な公開 symbol に限る。unexported helper や test からだけ呼ばれる symbol へ契約 tag を置くと公開境界との二重定義になり DRY を破る  # → layer:terms
184：2026-08-26 [feature/generator-cmd-entrypoint] 完了済み仕様を未完了 index に残さない。checklist を完了にしたら待ち一覧からも同時に外し、完了記録と未完了 index の責務を混ぜない  # → layer:workflow
185：2026-08-26 [feature/generator-cmd-entrypoint] 設計・実装の議論軌跡（なぜそうした／しなかったか）で code comment に残さなかった判断は Decision の Reason / Rejected へ残す。局所の Why not だけ code に置き、再発する問いへの答えは decisions が持つ  # → layer:meta
186：2026-08-26 [feature/generator-cmd-entrypoint] CLI の成否観測は runtime 内部状態ではなく OS process の exit code と stderr を正とする。同一 binary の写像を local と CI で分けない  # → layer:platform
187：2026-08-26 [feature/generator-cmd-entrypoint] 査読必須指摘は委譲先報告をそのまま採用せず、規約原文と現物を委譲元が照合してから修正対象へ格上げする。borderline 見送りも明文違反なら格上げする  # → layer:workflow
