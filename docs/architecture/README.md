# Architecture diagram（runtime）

`docs/architecture/runtime.png` は **code-first** で生成する。

## 生成元（SSoT）

1. `apps/diagrams/runtime.py`：runtime 構成図の Diagram 定義と PNG 生成
2. `apps/diagrams/icons.py`：catalog 外 icon の SVG 取得・キャッシュ・rasterize

更新の原則：

- PNG（`docs/architecture/runtime.png`）を手編集しない
- 変更は `apps/diagrams/**` へ入れて再生成する

## Philosphy（参照）

設計規則・DRY/KISS の判断基準は **architecture philosophy** を正とする：

- `file:///Users/shim0729/.cursor/skills/0:meta/philosophy/design-philosophy.md`

## 生成手順

```bash
cd apps/diagrams
python runtime.py
```

