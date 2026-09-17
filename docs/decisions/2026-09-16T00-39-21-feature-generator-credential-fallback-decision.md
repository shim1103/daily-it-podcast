---
name: Gemini/Cursor credential の役割区分と、本番・test・rate計測 workflow への割当
date: 2026-09-16T00:39:21
branch: feature/generator-credential-fallback-decision
---

## 1. Decision

1. Gemini API credential を **役割で 3 分割**する：`GEMINI_API_KEY`（本番 primary、free）、`TEST_GEMINI_API_KEY`（test/rate計測 primary、free、別 account）、`SPARE_GEMINI_API_KEY`（全 workflow 共通の paid final-fallback）。3 つとも config の必須 key として揃える（Decision `2026-09-09T10-00-00` §「なぜ config の必須 key として足すか」の方針を維持）。
2. `SPARE_GEMINI_API_KEY` は produce-episode / system-test / generator-draft-rate / generator-tts-rate の **全 workflow が共通で使う**。rate計測 2 workflow（`generator-tts-rate.yml` / `generator-draft-rate.yml`）が「本番 Secrets を使わない」としていた invariant（Decision `2026-09-03T14-45-00`）は、`SPARE_GEMINI_API_KEY` に限り本 Decision が supersede する。
3. `CURSOR_API_KEY` は produce-episode・system-test・generator-draft-rate で使う（TextWriter を持つ workflow 共通）。generator-tts-rate は使わない（TTS は Cursor が提供しない機能のため）。
4. `TEST_CURSOR_API_KEY` を廃止し、対応する本番 credential（`CURSOR_API_KEY`）へ統合する。値が本番・test で同一のため、`TEST_` 接頭辞を持つ理由がない。
5. `TEST_SPARE_GEMINI_API_KEY` を廃止する。system-test 実行時の Spare fallback 先も `SPARE_GEMINI_API_KEY`（本番 paid）を直接使う（§1-2）。
6. `R2_BUCKET` / `R2_ACCESS_KEY_ID` / `R2_SECRET_ACCESS_KEY` は引き続き本番・test で値を分け、`TEST_` 接頭辞を維持する（生成物の書込先を分離する意図。Decision `2026-09-03T14-45-00` の folder 分離方針を維持）。`R2_ACCOUNT_ID` のみ本番・test で値が同一のため `TEST_R2_ACCOUNT_ID` を廃止し統合する。
7. `TEST_` 接頭辞の付与基準は「値が本番と異なる credential/variable にのみ付ける」に統一する。値が同一なのに接頭辞だけ変えている credential は残さない。

### 1-1. workflow 別 fallback 順序

| workflow | 機能 | primary | fallback 順序 |
|---|---|---|---|
| produce-episode | TextWriter | `GEMINI_API_KEY`（本番 free） | → `CURSOR_API_KEY` → `SPARE_GEMINI_API_KEY`（本番 paid, final） |
| produce-episode | TTS | `GEMINI_API_KEY`（本番 free） | → `SPARE_GEMINI_API_KEY`（本番 paid, final） |
| system-test | TextWriter | `TEST_GEMINI_API_KEY`（test free） | → `CURSOR_API_KEY`（本番と共用） → `SPARE_GEMINI_API_KEY`（本番 paid, final） |
| system-test | TTS | `TEST_GEMINI_API_KEY`（test free） | → `SPARE_GEMINI_API_KEY`（本番 paid, final） |
| generator-draft-rate | TextWriter 計測 | `TEST_GEMINI_API_KEY`（test free） | → `CURSOR_API_KEY`（本番と共用） → `SPARE_GEMINI_API_KEY`（本番 paid, final） |
| generator-tts-rate | TTS 計測 | `TEST_GEMINI_API_KEY`（test free） | → `SPARE_GEMINI_API_KEY`（本番 paid, final） |

TTS を持つ workflow（produce-episode TTS / system-test TTS / generator-tts-rate）は Cursor を fallback に含めない（Cursor は TTS を提供しない）。

### 1-2. credential 一覧

| Secret/Variable | 用途 | 本番/test |
|---|---|---|
| `CURSOR_API_KEY` | TextWriter fallback（cursor 段） | 共通 |
| `GEMINI_API_KEY` | TextWriter/TTS primary（本番） | 本番専用 |
| `TEST_GEMINI_API_KEY` | TextWriter/TTS primary（test/rate 計測） | test 専用 |
| `SPARE_GEMINI_API_KEY` | TextWriter/TTS final-fallback（paid） | 全 workflow 共通 |
| `R2_ACCOUNT_ID` | R2 アカウント | 共通 |
| `R2_BUCKET` / `TEST_R2_BUCKET` | R2 バケット | 分離 |
| `R2_ACCESS_KEY_ID` / `TEST_R2_ACCESS_KEY_ID` | R2 認証 | 分離 |
| `R2_SECRET_ACCESS_KEY` / `TEST_R2_SECRET_ACCESS_KEY` | R2 認証 | 分離 |

正本は config（`internal/config/names.go` / `config.go`）と各 workflow yaml の `env:` である。この表は現時点の方針を可読な形でまとめたものであり、credential を追加・削除・付替えするたびに本 Decision を書き換えない（構成が変われば新規 Decision を CREATE して本 Decision を supersede する）。

Google Drive / OAuth の credential（`GOOGLE_OAUTH_*` / `DRIVE_FOLDER_ID`）は、develop の R2 完全移行（`e86c831 refactor(storage): generator/playbackからGoogle OAuth/Drive実装を削除する`）により generator から既に撤去済みのため、本 Decision の対象外とする。

## 2. Reason

1. 先行 Decision（`2026-09-09T10-00-00`）は「TTS と原稿 fallback で key を分ける」という 2 分割を確定させたが、system-test 側が `TEST_SPARE_GEMINI_API_KEY` という test 専用の Spare を別に持つ構成だった。test 用 account（別 Google account）には Spare 相当の paid project を新設していないため、test 側だけ Spare fallback が事実上機能しない歪みがあった。`SPARE_GEMINI_API_KEY` を全 workflow 共通にすることで、この歪みを無くし、test 用に paid project をもう1つ増やす必要もなくなる。
2. rate計測 workflow が「本番 Secrets を使わない」としていた理由（Decision `2026-09-03T14-45-00` 系）は、無料枠の primary key を本番と共有して枯渇を招かないことが目的だった。`SPARE_GEMINI_API_KEY` は元々枯渇時の最終手段であり、rate計測が failure-path の検証（fallback が実際に機能するか）を行うなら、その経路の終端も本番と同じ `SPARE_GEMINI_API_KEY` を通す方が、rate計測結果が本番の実挙動を代表する。
3. `TEST_` 接頭辞を「test 専用の別 project/account」を指す印として使うなら、値が本番と同一の credential にまで機械的に接頭辞を複製するのは冗長で、どの credential が実際に分離されているかを名前から読み取れなくなる。値が同一なものを統合することで、`TEST_` が付く credential＝値が本当に違う、という対応が name だけで分かるようにする。
4. R2 の bucket・access key は書込先そのものを分離する理由で分離を維持する。一方 `R2_ACCOUNT_ID` は書込先を分離する主体ではなく、Cloudflare account を指す識別子に過ぎず本番・test で共有できる値のため統合する。
5. fallback 順序（primary → 何段目に何を試すか）と credential 一覧は、文章で workflow ごとに書くより表にまとめた方が、workflow 横断で読み取れる。表は decisions.md が禁じる「境界契約 artifact の値の再掲」には当たらない——禁じられているのは config の具体的な値（secret の中身等）の複製であり、credential 名と役割・順序という方針の記述は Decision の本来の役目である。

## 3. Rejected

1. **test account にも Spare 相当の paid project を新設し `TEST_SPARE_GEMINI_API_KEY` を維持する案** — project 数が増え、無料枠を増やす目的の複製に近い構造になる（前回セッションで規約リスクとして議論済み）。共通の `SPARE_GEMINI_API_KEY` を使えば project を増やさずに済む。
2. **rate計測 workflow は引き続き本番 Secrets を一切使わず、Spare fallback 経路をそもそも検証しない案** — fallback 経路（Gemini free → Cursor → Gemini paid）の実挙動を確認する目的の rate計測であれば、経路の一部（Spare 段）だけ検証対象外にすると計測の代表性が失われる。
3. **`TEST_` 接頭辞を全 credential から一律で外し、本番・test を workflow の実行 context だけで区別する案** — 値が実際に異なる credential（`TEST_GEMINI_API_KEY` 等）まで接頭辞を外すと、どの credential が test 用の別実体を指すか名前から読み取れなくなる。値が異なるものだけ接頭辞を残す方が、config 一覧を見ただけで分離の有無が分かる。
4. **`R2_ACCOUNT_ID` も本番・test で分離したまま残す案** — account 識別子自体は書込先を分離する主体ではなく、値が同一であることが確認できているため、分離を維持する理由がない。
