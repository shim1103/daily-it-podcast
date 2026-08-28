## 1. Summary

このIssueでは、`apps/generator` の error 実装が3層（Domain / Infra / Config）で非対称なのを、Infra の pattern（`type Error struct { <識別子> string; Err error }` + `xxxErr(<識別子>, err)` helper + `Error()` + `Unwrap()`）へ揃える。完了後は3層すべてが同じ struct 形・helper 命名・`Error()` prefix パターンを持ち、`errors.Is` / `errors.As` / `Unwrap()` が全 error で一様に効く。

## 2. Context

`apps/generator` の error は現状3層で表現方式がばらばらである。

1. **Domain**（`internal/entities/errors/*.go`）: 種別ごとに固有 field を持つ型を量産している。`CorruptSpeechAudio{Err}` / `EmptyEpisodeID{}` / `EmptyAudio{}` / `EpisodeIDMismatch{Expected,Actual}` / `InvalidManuscript{Err}` / `InvalidManuscriptDraft{Err}` の6型。生成は `&domainerrors.X{}` の直書きで、`internal/application/write_episode.go` に10箇所ある。`Unwrap()` は一部の型（`CorruptSpeechAudio` / `InvalidManuscript` / `InvalidManuscriptDraft`）だけが持ち、`EmptyEpisodeID` / `EmptyAudio` / `EpisodeIDMismatch` は持たない。
2. **Infra**（`internal/infrastructure/*/error.go`、7 package が同一形）: `type Error struct { Op string; Err error }` + `func (e *Error) Error() string` + `func (e *Error) Unwrap() error` + `func infraErr(op string, err error) error`。`Error()` は `"<package名>: " + Op + ": " + Err.Error()` の prefix パターン。呼び出しは60箇所を超える。
3. **Config**（`internal/config/error.go`、runtime config loader session で追加）: sentinel error 3個（`ErrMissing` / `ErrEmpty` / `ErrInvalidFormat`）。`load.go` で `fmt.Errorf("%s: %w", key, err)` して `errors.Join` で束ねる。

error を表示するのは1箇所のみ、`cmd/generator/main.go` の `fmt.Fprintf(stderr, "generator: %v\n", err)` である。表示点は1箇所だが、design-philosophy §5-1（Principle of Least Astonishment / 一貫性）を §5-3（KISS）より優先し、「error の作り方・触り方」の一貫性を取る。

## 3. Canonical Sources

1. `apps/generator/internal/infrastructure/x/getxapi/error.go` — 倣う基準形。`type Error struct { Op string; Err error }` + `infraErr()` + `Error()` prefix + `Unwrap()`。7 package が同一形で、変更コストが最大なので動かさない。
2. `apps/generator/internal/entities/errors/*.go` — 廃止対象の Domain 6型（`corrupt_speech_audio.go` / `empty_audio.go` / `empty_episode_id.go` / `episode_id_mismatch.go` / `invalid_manuscript_draft.go` / `invalid_manuscript.go`）。
3. `apps/generator/internal/application/write_episode.go` — Domain error の生成10箇所。
4. `apps/generator/internal/config/error.go` + `apps/generator/internal/config/load.go` — Config 側の sentinel と `errors.Join`。
5. `apps/generator/cmd/generator/main.go` — 唯一の表示点。`%v` のまま変えない。
6. skill `error-handling`（`defensive-design.md`） — Error クラス設計・wrapping・静的網羅性・表示の住み分け。
7. skill `architecture`（`error-taxonomy.md` / `logging-policy.md`） — Domain / Infra / External の層分類と logging 方針。
8. skill `philosophy`（`design-philosophy.md` §5） — §5-1 一貫性 > §5-3 KISS の優先順位。
9. skill `testing-strategy`（`contracts.md`） — 各 test level の GWT 契約。

## 4. Scope

### In Scope

1. Domain 6型を廃し、`internal/entities/errors` に `type Error struct { Op string; Err error }` 1型 + `domainErr(op string, err error) error` helper へ集約する。`Op` は `"empty_episode_id"` / `"empty_audio"` / `"invalid_manuscript"` / `"invalid_manuscript_draft"` / `"episode_id_mismatch"` / `"corrupt_speech_audio"` 等の文字列で語彙を表す。`Error()` は Infra と対称な `"<prefix>: <op>: <詳細>"` パターン、`Unwrap()` は `Err` を返す。
2. `write_episode.go` の生成10箇所を `domainErr()` 経由へ追従させる。`EpisodeIDMismatch{Expected, Actual}` の2値は `Err` へ畳む（例: `domainErr("episode_id_mismatch", fmt.Errorf("expected %q actual %q", expected, actual))`）。`EmptyEpisodeID{}` / `EmptyAudio{}` の引数なし型は `domainErr("empty_episode_id", nil)` 相当へ（`Err == nil` 時の `Error()` は詳細なしで返す）。
3. `write_episode_sociable_unit_test.go` と `domain_error_sociable_unit_test.go` の `errors.As` 型判定を、新しい単一 `*errors.Error` + `.Op` 文字列判定へ書き換える。GWT 構造は保つ。
4. `internal/config/error.go` を sentinel から `type Error struct { Key string; Err error }` + `configErr(key string, err error) error` へ。sentinel（`ErrMissing` 等）を `Err` の中身として保持するか、`Kind` 相当をどう表現するかは実装判断だが、Infra 対称を優先する。
5. `load.go` の `errors.Join` を、`[]*config.Error` を束ねるラッパへ。複数違反を1 error で返す振る舞いと `errors.Is` / `errors.As` の両対応を保つ。

### Out of Scope

1. Infra 7 package の `error.go` の変更（`infraErr` / `{Op,Err}` が基準なので触らない）。
2. error の表示ロジックの変更（`main.go` の `%v` のまま）。
3. 新しい error 分類の追加。
4. 層をまたぐ error 変換規則（Infra→Domain 等の translate ポリシー）の変更。
5. `error-taxonomy.md` の「Domain / Infra / External は発生層が違う」という層分類そのものの変更。統一するのは *表現形式* であって層の意味ではない。

## 5. Contract

1. 3層の error 型が対称になる。struct 形（`type Error struct { <識別子> string; Err error }`）、helper 名（`domainErr` / `infraErr` / `configErr`）、`Error()` の prefix パターン（`"<prefix>: <識別子>: <詳細>"`）が揃う。field 名は層の識別子の性質に合わせてよい（Infra=`Op`、Config=`Key`、Domain=`Op`）。
2. `errors.Is` / `errors.As` が3層すべての error で効く。
3. `Unwrap()` で cause chain がたどれる。`Err == nil` の error も `Error()` が panic せず詳細なしメッセージを返す。
4. raw secret / raw runtime 値を `Error()` の文字列に含めない。
5. Config は複数違反を1 error に束ねたまま、束の各要素へ `errors.As` で到達できる。

## 6. Constraints

1. Infra の `infraErr` / `{Op,Err}` を変更しない。
2. GitHub Issue 化しない。本 file が契約の正。
3. 既存の各層 test の GWT（Given / When / Then）構造を保つ。
4. `main.go` の表示（`fmt.Fprintf(stderr, "generator: %v\n", err)`）を変えない。
5. `apps/generator` 外の code を変更しない。

## 7. Acceptance Criteria

1. [ ] Domain が単一 `Error` 型 + `domainErr()` になり、6型（`CorruptSpeechAudio` / `EmptyAudio` / `EmptyEpisodeID` / `EpisodeIDMismatch` / `InvalidManuscript` / `InvalidManuscriptDraft`）が消えている。
2. [ ] `write_episode.go` の error 生成が全て `domainErr()` 経由になっている。
3. [ ] Domain の各 test が `errors.As(err, &target)` （`target` は `*errors.Error`）+ `.Op` 文字列で種別を確認している。
4. [ ] `EpisodeIDMismatch` 相当の情報（expected / actual）が `Err` 経由で保たれ、`Error()` 文字列から読める。
5. [ ] Config が `type Error struct { Key string; Err error }` + `configErr()` + 複数違反ラッパになり、Infra と対称になっている。
6. [ ] `errors.Is` と `errors.As` が Config error で両対応する（束ねた各違反へ `errors.As` で到達できる）。
7. [ ] 3層の `Error()` 文字列が `"<prefix>: <識別子>: <詳細>"` パターンで揃っている。
8. [ ] `Error()` に raw secret / raw runtime 値が露出していない。
9. [ ] `go test ./... -count=1` と `go test ./... -race -count=1` が pass する。
10. [ ] `./scripts/generator/check-static.sh` が 0 issues。
11. [ ] 総 coverage gate が維持されている。

## 8. Verification

```bash
cd apps/generator
go test ./... -count=1
go test ./... -race -count=1
cd ../..
./scripts/generator/check-static.sh
./scripts/generator/test-unit.sh
git diff --check
```

## 9. Dependencies

1. `apps/generator/internal/config` の config loader 実装・KISS 化（runtime config loader session の PR で完了。進捗は `docs/tasks/todo/generator-lane.md` の C-04 行）に後続する。
2. `docs/tasks/todo/generator-composition-http-adapters.md`（M1）とは独立だが、両方が config package を触るため、着手が重なる場合は順序調整が要る。

## 10. Risks

1. Domain 6型の廃止は `write_episode.go` と test 2 file（`write_episode_sociable_unit_test.go` / `domain_error_sociable_unit_test.go`）へ波及が広い。1 commit でやると configuration boundary（Config 側）と Domain migration の失敗が混ざる。→ Domain と Config を別 commit に分ける。
2. `Op` 文字列の typo は compile で捕まらない。→ `Op` を定数化するか、有効な語彙を table で検証する。
3. `errors.As` を型から文字列 `.Op` 判定へ変えることで、誤った `Op` を assert する test が silently pass する余地が出る。→ 各 test で `Op` の期待値を明示し、`Error()` 文字列も併せて assert する。

## 11. Notes

1. この runtime config loader session の transcript が設計判断（Infra 基準・§5-1 > §5-3・6型畳み込み）の出所である。決定は本 Issue 側へ SSOT 化したので、transcript なしで着手できる。
2. 却下案: (a) 3層とも sentinel 化（案B）→ Infra 7 package 全書き換え + `Op` の60種粒度を sentinel にすると変数爆発。(b) 現状維持で interface 一貫のみ → 明示的に却下済み。
