---
name: playback の層境界は code（dir・depcruise・Drive schema）を正とし static gate する
date: 2026-08-25T19:20:00
branch: chore/playback-worker-web-layer
---

## 1. Decision

1. Feature / Primitive の dir・層 import・Drive 原稿検証の正は code にある（`web/src/components/{feature,primitive}/`・`apps/playback/.dependency-cruiser.mjs`・`worker/.../manuscript-schema.ts`）。本 Decision は値を写さない
2. 層違反は `dependency-cruiser` を `scripts/playback/check-static.sh` から実行して検知する。generator の depguard と同型（allow-list・production のみ・Composition/Page 無制限・Application→`apps/playback/contracts` 許可）
3. 本 decision は次を上書きする: `2026-08-20T19-29-21-playback-web-layer-layout`（Feature/Primitive を dir 分割しない）、`2026-08-17T14-45-00` / `2026-08-18T14-35-00` / 旧 `DESIGN.md` §5.9（playback に層 lint を載せない）

## 2. Reason

1. Feature と Primitive は import 規則が違う（Primitive は ViewModel 不可）。同一 dir のまま file 名例外で分けると、新 file 追加のたびに例外追記が必要になり、generator の「allow に無いものは止まる」と逆になる
2. 以前「分けない」とした理由は component file が 0 個の時点の YAGNI だった。いま Feature 複数と Primitive が共存するため、その却下理由は消えている
3. Biome は層 import の allow-list を持たない。ESLint を足すと静的検査が二重になる。`dependency-cruiser` は TS 側で depguard 相当を果たす
4. Application が HTTP 応答を組み立てる現状を skill 厳格だけで切ると、generator（Application→contracts 許可）と playback で読み手が食い違う。横断判断を1つにするなら generator 寄せが足りる
5. Drive 原稿に HTTP 専用 field は無い。HTTP schema から Drive 検証を導くと、HTTP 変更が Drive 読取を壊す。Drive の正は repo 根 `contracts/manuscript.schema.json`（`DESIGN.md`）

## 3. Rejected

1. `components/` 直下のまま Primitive だけ path 例外 — 例外表が肥大し、新 file で穴が開く
2. Feature/Primitive を `web/src/feature/`・`web/src/primitive/` にして `components` を廃止 — 既存の Component 傘と page からの相対 path を広く壊す
3. ESLint + `eslint-plugin-boundaries` — Biome と linter が分裂する
4. Application → `apps/playback/contracts` を禁止して Port 型を手定義 — HTTP 契約と Application 型が二重 SSOT になる
5. Infrastructure が HTTP schema の omit を維持 — Drive と HTTP の混線が残る
6. 境界契約を markdown に置いて code と二重化する案 — 契約 SSOT が分裂する
