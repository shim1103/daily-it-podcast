---
name: cmd/generator は薄い Driving Adapter であり秘密・手順・Infra を持たない
date: 2026-08-25T23:12:31
branch: feature/generator-cmd-usecase-boundary
---

## 1. Decision

1. `apps/generator/cmd/generator` は CLI の **Driving Adapter**（Frameworks & Drivers）とする。HTTP の Route Handler と同型の薄い入口である。
2. cmd がしてよいこと: process の signal 付き `context` を作る、`composition.NewProduceEpisode()` を呼ぶ、`Run(ctx, time.Now())` を呼ぶ、失敗を stderr へ出し非0 exit にする。
3. cmd がしてはいけないこと: 秘密・env の注入、brief / Draft / TTS / 尺 / 完成原稿の手順、Port や Infrastructure の直接 import、完成 `manuscript.schema.json` の Validate。
4. 秘密の供給は process の caller（GHA / local shell / AgentSecrets wrapper）と Composition の結線が所有する。原則の正は architecture `backend/route-handler.md`（薄い入口）・`backend/composition-root.md`（結線）および既存 `2026-08-25T13-53-55`（env 注入は Composition / runtime、GHA は caller）。

## 2. Reason

1. Driving Adapter に手順を置くと、Application の Builder（`ProduceEpisode`）と入口が二重に方針を持ち、CLI と将来の別入口で手順が分岐する。入口は wire と終了コードだけに閉じ、方針は Application に残す（philosophy SRP / Orthogonality）。
2. cmd が `os.Getenv` で秘密を読むと、Composition の `SecretRef` binding と processenv 契約を迂回し、秘密境界が入口ごとに増える。caller が env を載せ Composition が結線する形が既存決定と一致する。
3. Infrastructure を cmd から import すると Composition Root の全層 import 特権が入口へ漏れ、depguard / Dependency Rule の例外点が複数になる。

## 3. Rejected

1. cmd が Fetch → TextWriter → TTS → Write を直列に呼ぶ案 — 手順が Driving Adapter に漏れる。
2. cmd が必須 env 名の値を読んで Adapter を new する案 — 結線と秘密供給が Composition から外れる。
3. cmd が `EpisodeWriter` / `SpeechSynthesizer` 具象を直接組み立てる案 — Composition Root の唯一点が崩れる。
4. ProduceEpisode 未実装のうち cmd で panic を recover して「成功」にする案 — 失敗を隠す。Run 本体は別 scope（D）。
