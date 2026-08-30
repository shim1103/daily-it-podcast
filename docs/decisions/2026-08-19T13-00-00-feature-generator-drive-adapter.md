---
name: repo 根 contracts/ は Drive Adapter（Infrastructure）が file として読む
date: 2026-08-19T13:00:00
branch: feature/generator-drive-adapter
---

## Status

**Superseded.** generator については `2026-08-19T15-00-00-feature-generator-drive-adapter-layer-split.md` が正。schema の読み手は Application。playback worker の Drive 読取 Adapter は本 decision の「Infrastructure が読む」を **generator とは別 runtime として維持**（`DESIGN.md` 参照）。

## 1. Decision

repo 根 `contracts/`（Drive 配置・原稿 JSON Schema）を読むのは、generator と playback worker の **Infrastructure（Drive Adapter）** だけとする。Entities / Application / Composition Root / cmd / playback web は読まない。正本の置き場と読み手の対応は `DESIGN.md` を正とする。

## 2. Reason

`contracts/` は language 横断の wire format 正本であり、Go / TS の import 可能な package ではない。誰も読めないと書いたままだと、Schema が死んだ文書になり、各 runtime が field を手写しする。Drive の I/O 直前で enforce する責務は Driven Adapter にある。skill の「境界共有型を Infrastructure が import してはならない」は `apps/playback/contracts/`（HTTP DTO）向けであり、repo 根 `contracts/` と混同しない。

## 3. Rejected

1. `contracts/` は文書専用で runtime は誰も読まない案（正本が使われず、手写しが第二の SSOT になる）
2. Application / UseCase が Schema を読んで検証する案（Application が Drive wire を知る）
3. repo 根 `contracts/` を Go / TS package にして内側層から import する案（言語横断の型 module 禁止、Entities の外部依存ゼロを破る）
4. repo 根 `contracts/` を HTTP 境界共有型と同じ import 禁止に置く案（Drive Schema と playback HTTP を同一規則へ潰す）
