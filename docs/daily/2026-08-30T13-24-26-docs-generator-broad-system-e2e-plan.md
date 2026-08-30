---
name: Generator Broad/System 境界と GHA 定時運用を固定する
date: 2026-08-30T13:24:26
session_id: none
branch: docs/generator-broad-system-e2e-plan
prev: なし
---

## 1. Summary

Generator の Broad / System の A/B と Broad の C、GHA 本番・System workflow、README/DESIGN/DEPLOY の責務分離までを1 session で固定した。System suite 本体と TEST_* 値登録は後続。

## 2. Changes

1. Decision 5本で gate=Narrow+Broad、System=gate外、分類語=`system`、GHA `TEST_` 写像、定時（毎日/週次 JST）を固定した。
2. System stub package・`test-system.sh`・`produce-episode.sh`・2 workflow を置いた。`-tags=system` の同一 package 再実行を避け `test/system` へ分離した。
3. Broad Integration 達成契約 file を切り、System は lane D に残した。
4. README/DESIGN の運用重複を DEPLOY へ寄せ、condition report 文書の正を DESIGN へ移した。
5. 開閉挨拶定数を episode 向け文案へ更新した（session 外差分を repo 全変更として同梱）。

### Commits

- `68ca9ae`
- `67b17d3`
- `5771e48`
- `cad1267`
- `08066ba`
