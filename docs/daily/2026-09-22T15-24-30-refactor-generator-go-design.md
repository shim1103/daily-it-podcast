---
name: generator Go 設計一本線と Decision 固定
date: 2026-09-22T15:24:30
session_id: 522bb03b-50c2-4f9b-8f42-84ba9eb430cd
branch: refactor/generator-go-design
prev: なし
---

## 1. Summary

Go vs JS/TS の学習 Q&A を経て、generator の Application 配置・CLI 入口・root npm 重複を一本線へ揃え、再発する設計判断を Decision に固定した。system e2e 契約は観測可能な postcondition に圧縮した。

## 2. Changes

1. root の Biome 専用 `package.json` / lock を削除し、npm 正本を playback のみにした
2. `cmd` を `runCLI` + `os.Exit(runCLI())` にし、defer 後始末と exit を両立させた
3. `fetch` / `writeepisode` を独立 package にし、Gate を `EpisodeWriter` として `ProduceEpisode` へ渡し、Fetch は UseCase DI のまま残した
4. system produce_episode 契約から How 列挙を外した
5. 上記を Decision 3 本と DESIGN の総称参照で固定した
6. 学習中の agent 誤認（Biome=monorepo lint、同層だから pointer、Go 思想と port 配置の混同）を訂正し、lessons へ一般化した

### Commits

- `ad4b70a`
- `f409fa9`
- `3910ba1`
- `5c13c32`
- `ad67dfa`
