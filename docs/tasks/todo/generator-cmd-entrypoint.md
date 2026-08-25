## 1. Summary

この Issue では、`cmd/generator` を薄い Driving Adapter として完了させ、既存 Port / UseCase 境界を A・B に揃える。完了後、binary 入口は Composition 経由でだけ UseCase を呼び、秘密・生成手順・Infra を持たない。`ProduceEpisode.Run` の生成ロジック本体は扱わない。

## 2. Context

1. lane 上の cmd は「`.gitkeep` のみ」と残っているが、作業 tree には `main.go` stub がある。責務の Decision が無かったため `docs/decisions/2026-08-25T23-12-31-feature-generator-cmd-usecase-boundary.md` を正とする。
2. TextWriter は string 戻りへ戻済み（`2026-08-25T23-02-35`）。Intro-only Draft を Adapter に置く案は却下済み。
3. `ProduceEpisode` は契約 stub（`Run` は panic）。本 Issue はその詳細仕様・実装をしない（lane / D）。
4. 仮定: Composition `NewProduceEpisode` の結線は維持する。Run が stub の間、起動すると panic しうる。それは本 Issue の失敗ではなく D 待ちとする。

## 3. Canonical Sources

1. `docs/decisions/2026-08-25T23-12-31-feature-generator-cmd-usecase-boundary.md` — cmd Driving Adapter 責務。
2. `docs/decisions/2026-08-25T22-37-27-feature-generator-cmd-usecase-boundary.md` — Builder / Gate。
3. `docs/decisions/2026-08-25T22-37-30-feature-generator-cmd-usecase-boundary.md` — Composition の戻り単位。
4. `docs/decisions/2026-08-25T23-02-35-feature-generator-cmd-usecase-boundary.md` — TextWriter は string、Draft 化は ProduceEpisode。
5. `apps/generator/cmd/generator/main.go` — 入口の正本候補。
6. `apps/generator/internal/composition/produce_episode.go` — 結線。
7. architecture — `backend/route-handler.md`（薄い入口）、`backend/composition-root.md`、`backend/application.md`（Builder/Gate）。
8. test 方針 — `testing-strategy`。

## 4. Scope

### In Scope

1. `cmd/generator` を B-cmd に適合する薄い入口として完了する（signal ctx、Composition 呼び出し、stderr + 非0 exit）。
2. cmd が Infrastructure / Port 具象を import しないことの固定。
3. 既存実装の A・B 整合（TextWriter=string、WriteEpisode が組み立てを持たない、契約 comment の食い違い掃除）。
4. `generator-lane.md` の cmd / 束ね UseCase 表記を実態と Issue 参照へ更新する。

### Out of Scope

1. `ProduceEpisode.Run` の詳細仕様・実装（Draft parse、定型結合、TTS 順、尺、完成 JSON）。
2. `manuscriptDraftFromWriterOutput` / `wavDurationSec` / `concatWAV` の本体実装。
3. 定型挨拶文言・brief 文言・空 Fetch policy・`episodeId` 規則の確定。
4. GHA workflow。
5. GitHub Issue の remote 作成（本 file が Issue SSOT。`shim gh create-issue` は別指示）。

## 5. Contract

1. 入口 package は `apps/generator/cmd/generator`。`main` のみ。
2. 成功時: `Run` が nil を返したら process exit 0。
3. 失敗時: `Run` が non-nil error を返したら stderr に出し exit 非0。
4. cmd は `internal/composition` 以外の `internal/infrastructure` / `internal/application/port` を import しない。
5. 既存 `port.TextWriter.Write` は `(string, error)` のまま（変えない）。

## 6. Constraints

1. env / 秘密値を cmd で読まない・注入しない。
2. `ProduceEpisode.Run` 本体を本 Issue で実装しない。
3. Gate（`WriteEpisode`）に組み立て・TTS・尺を移さない。
4. 汎用規則は architecture / testing-strategy を参照し、Issue へ再定義しない。

## 7. Acceptance Criteria

1. [ ] `docs/decisions/2026-08-25T23-12-31-feature-generator-cmd-usecase-boundary.md` が cmd 責務の正として存在する。
2. [ ] `cmd/generator` に `main.go` があり、`.gitkeep` だけではない。
3. [ ] `main` は signal 付き context と `composition.NewProduceEpisode().Run` のみで UseCase に入る。
4. [ ] `cmd/generator` の import に `internal/infrastructure` と `application/port` が無い。
5. [ ] `TextWriter` Port / Cursor Adapter が string 戻りであり、Adapter が `ManuscriptDraft` を組み立てない。
6. [ ] `generator-lane.md` の cmd 行が本 Issue を参照し、`.gitkeep` のみ表記が消えている。
7. [ ] Generator static / Unit gate が pass する。

## 8. Verification

```bash
cd apps/generator && go build -o /dev/null ./cmd/generator/
cd apps/generator && go build ./...
./scripts/generator/check-static.sh
./scripts/generator/test-unit.sh
```

1. `go list -f '{{.Imports}}' ./cmd/generator`（または同等）で infrastructure / port が混入していないこと。
2. `ProduceEpisode.Run` の panic stub は本 Issue の失敗条件にしない。

## 9. Dependencies

1. related: Builder/Gate・Port string の既存 Decision（Canonical Sources）。
2. blocks: `ProduceEpisode.Run` 実装（D / 後続 Issue）。本 Issue 完了が Run 実装の前提入口になる。

## 10. Risks

1. Run stub のまま「cmd 完了」と誤解し lane 全体を閉じる → Out of Scope と Notes で D を明示する。
2. exit mapping を厚くして Domain/Infra 分類を cmd に持ち込む → B-cmd どおり薄く保つ。詳細 mapping が必要なら別 Issue。

## 11. Notes

1. 採らない: cmd で Fetch と Write を直列呼び出しする案（B-cmd Rejected）。
2. follow-up: `ProduceEpisode.Run`、GHA、挨拶文言・空 items policy。
