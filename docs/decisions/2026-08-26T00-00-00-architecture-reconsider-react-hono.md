---
name: playback web/worker の技術選定は「既存機能要件（YAGNI）+ 学習目的の裁量採用」を重視軸にする。React・Hono・Hono RPC を採用、CSS は Pico.css 継続、TanStack 系・Next.js・OpenAPI は見送る
date: 2026-08-26T00:00:00
branch: docs/architecture-reconsider-react-hono
---

## 1. Decision

1. 技術選定の重視軸を次の順で固定する。
   1. 学習コスト・移行コスト・開発者個人の慣れ・生産性は architecture 判断の軸にしない
   2. 既存 repo に対応する機能要件が実在するか（YAGNI）を優先軸にする
   3. 機能要件が薄くても、個人の学習目的（frameworkを採用しつつ内部実装を読み設計意図を理解する）としての採用は正当な裁量として認める。ただし車輪の再発明（自作での代替実装）は学習手段として採用しない
2. この軸のもとで、React・Hono・Hono RPC を採用する
3. CSS は Pico.css（classless）を継続する。Tailwind CSS・shadcn/ui は採用しない
4. TanStack（Query / Router / Start）・Next.js・OpenAPI は採用しない

## 2. Reason

1. 評価軸の確定：Least Astonishment を最優先する `design-philosophy.md §5` に「開発者の慣れ」は軸として存在しない。一方で Rule of Least Power（§4-2）は「独自 DSL 再発明」を明確に戒めており、既存機能要件を自作で代替している箇所（例：worker の自作 router）は学習目的であっても Hono 側に寄せるべき対象になる
2. Hono：`match-playback-route.ts` は method 判定・path 分解・exhaustive check という router 標準機能を自作しており、対応する機能要件が repo 内に実在する。学習目的採用の条件（車輪の再発明をしない）にも合致する
3. Hono RPC：`contracts/` の手動 import による型同期は、Hono 採用後は Hono 側の機構に置き換えられる対応関係にある
4. React：現状の要件（一覧・詳細・audio 再生、state 少数）には対応する機能（差分検知・深い component 木）が薄く、architecture 理由の merit は小さい。ただし Hono 採用により「repo 全体を vanilla・最小 dependency で統一する」という前提自体が崩れるため、React 追加による repo 内一貫性コストは実質消滅する。architecture 上の必然性は無いが、学習目的の裁量採用として認める
5. CSS：重視するのは「凝りすぎない」ではなく「実装コストを掛けない」「優れた default を提供する」。Pico.css classless はこの2条件を満たす。Tailwind は優れた default を持たず規律依存が大きい。shadcn/ui は Radix/Base UI を固定し生成 code が repo に露出する運用で、現規模に過剰
6. TanStack：Router は画面遷移が一覧⇔詳細の2画面のみで型安全 path の恩恵が薄く、Query は現行の手書き fetch で cache・再検証要件が無く、Start は SSR/SEO 機能一式で Next.js 却下理由がそのまま適用される。いずれも対応する機能要件が repo に無い

## 3. Rejected

1. React 不採用の維持（`2026-08-18T11-12-00-feature-playback-web.md`）— Hono 採用で一貫性前提が崩れたため再評価し、学習目的の裁量採用に変更する
2. 「開発者の慣れ・生産性」を正式な評価軸に追加する案 — Least Astonishment との整合が取れず、判断が個人のtoolchain変化のたびに揺れる
3. Tailwind CSS・shadcn/ui — 前者は優れた default を持たない、後者は現規模に過剰かつ生成 code の repo 露出を伴う
4. TanStack Router / Query / Start・Next.js・OpenAPI の採用 — いずれも対応する機能要件が現状に無い

## 4. Non-scope（この decision が固定しない範囲）

1. dir・型・依存追加などの契約固定（A）
2. 既存 code の実装修正（C）
3. deploy・UI 完成・backend 完成（D、既存 `docs/tasks/todo/playback-lane.md` の対象）

この decision は重視軸と技術選定のみを SSOT として固定する。A/C/D は別途、この decision を Canonical Source として着手する。
