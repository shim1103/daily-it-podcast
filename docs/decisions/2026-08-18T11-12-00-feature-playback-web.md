---
name: playback web は Vite + TypeScript + Pico.css。React / Next.js / shadcn は使わない
date: 2026-08-18T11:12:00
branch: feature/playback-web
---

## 1. Decision

1. `apps/playback/web` は Vite + TypeScript（vanilla）。React / Next.js は使わない
2. 見た目は Pico.css の classless 版を repo 内に固定し、semantic HTML をそのまま整える
3. 再生 UI は `<audio controls>` を使う。component library（shadcn/ui 等）は導入しない
4. `apps/playback/worker` は TypeScript（Cloudflare Workers の first-class）。Go / Wasm にはしない
5. `apps/generator` は Go stdlib CLI（HTTP framework なし）。GHA cron のまま
6. web↔worker HTTP 契約は `apps/playback/contracts/` の TS schema のまま（`2026-08-17T17-40-00` を維持）

## 2. Reason

1. 要件は Access 内の一覧・再生・原稿表示のみ。SEO・SSR・深い component 木は不要（KISS / YAGNI）
2. state は選択 episode・読込/失敗・`<audio>` の `currentTime` 程度。React の差分更新 merit が小さい
3. 「整った見た目・UI design に時間をかけない・originality 不要」は Pico.css classless が最小 cost。shadcn は React + Tailwind + Base UI/Radix を固定し、player 画面に不要な toolchain を増やす
4. Workers の native 言語は TS。Go on Workers は Wasm 経由で Least Power に反する。generator を Go にする理由（stdlib CLI・GHA batch）と worker を TS にする理由（edge runtime native）は同じ「runtime に合わせた native 言語」判断
5. frontend skill の層（page / feature / view-model / api-client / utils）は「何を知っているか」の分割。React hooks でなく TS + DOM event でも同じ ring 対応が取れる
6. CI / test は React なしでも Vitest + `tsc` で足りる。React 専用 linter は不要

## 3. Rejected

1. Next.js（App Router / Route Handler BFF）— UI framework が Drive 読取 BFF を飲み込み、2 系統分離と SRP を壊す
2. React + Vite — 今の UI 規模に対して component runtime が余剰。慣れは速度の話であり要件ではない
3. shadcn/ui + Tailwind — React 固定。一覧 + 原稿 + 標準 audio player 向けの部品セットではない
4. Tailwind のみ — utility を自分で組む設計時間が残る。Pico より重い
5. worker を Go（TinyGo / Wasm）— 契約が言語横断になり、Workers の first-class から外れる
6. generator に Gin / Echo 等 — HTTP server ではない batch CLI に framework は不要
