---
name: runtime図生成の code-first 化と pr-completion
date: 2026-09-04T01-23-01
session_id: none
branch: docs-architecture-diagram
prev: なし
---

## 1. Summary
`docs/architecture/runtime.png` を **手編集しない** 方針を、`docs/architecture/README.md` に明文化した。更新は `apps/diagrams/runtime.py` および `apps/diagrams/icons.py` へ寄せ、生成物は `python apps/diagrams/runtime.py` で再生成する。さらに `pr-completion` の工程（commit → log-session → `gh pr create`）で DRY/KISS が崩れないよう lessons へ追記した。

## 2. Changes

1. `docs/architecture/README.md`：PNGは手編集しない、生成元をSSoTとして参照
2. `docs/lessons/index.md`：pr-completion 時の図更新注意を追記

### Commits

- b3a0f78
- f0f9b22
- f768a83

