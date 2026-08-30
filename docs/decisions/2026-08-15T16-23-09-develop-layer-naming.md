---
name: Backend は層名、web は frontend 慣習、Port は Application 所有
date: 2026-08-15T16:23:09
branch: develop
---

## 1. Decision

generator と playback/worker は backend skill の層 dir（entities / application / infrastructure / composition）に寄せる。Port IF は Application が所有する。playback/web は frontend skill（pages / components / api / utils）。機能軸は app 単位、層軸は app 内。

## 2. Reason

DIP によりビジネス手順が具象 I/O に依存しない。web まで Entities 名に揃えると React 慣習と衝突する。

## 3. Rejected

- 全系で app/domain/port/adapter という独自語に揃える案（skill 用語と二重になる）
- 工程機能ごとに packages を切る案
