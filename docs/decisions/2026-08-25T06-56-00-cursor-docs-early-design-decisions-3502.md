---
name: NotebookLM 等の再構成型音声サービスを却下し、逐語 TTS と自前分割で timestamp を取る
date: 2026-08-25T06:56:00
branch: cursor/docs-early-design-decisions-3502
---

## 1. Decision

1. NotebookLM（および入力を対話・要約へ再構成する同様の hosted 音声生成）は採用しない。
2. 読み上げは TTS API を原稿に対して逐語で叩き、保存と再生 UI は自前とする。保存先の契約は `contracts/`、読み上げ Port は既存の `SpeechSynthesizer` Decision を正とする。
3. 目次・タイムスタンプはサービス native の TOC に頼らない。トピック単位に分割生成し、自前で時刻を算出する。

## 2. Reason

1. NotebookLM は入力原稿をそのまま読まず AI が内容を再構成する → 「逐語読み上げ」要求に反する。
2. 発言時刻を構造として持たず、タイムスタンプ・目次を native に出せない。
3. 音声長が粗い 2 択程度で、予算から決めた尺の細かい調整に向かない。
4. 対話形式固定で「1 人ニュース風」に合わない。

## 3. Rejected

1. NotebookLM を主生成路にする案
2. 音読さん / ElevenLabs 単体を製品全体の代替にする案（再生・保存・入場・逐語 pipeline を満たさない）
3. hosted 側の TOC を正とし、自前の分割時刻を捨てる案
