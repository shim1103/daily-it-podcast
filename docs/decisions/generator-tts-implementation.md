---
name: "音声合成（TTS）をどう実装するか？"
description: TextWriterと同型の境界（ApplicationにDomainと進行、Adapterは薄いHTTP）でCloudflare AI Gateway経由にする。既定の無料経路はWorkers AIのTTS（Neurons無料枠）とする。
---

## 1. Decision

1. 音声合成（`port.SpeechSynthesizer`）の実装境界を、原稿 TextWriter（`generator-text-writer-implementation`）と**同型**にする。
   1. **Application**: セグメント texts の意味、結合前の音声尺チェック、部分成功の保持と残り texts の進行（`SynthesizeAll`）。
   2. **Adapter（薄い）**: Gateway へ audio 呼び出し 1 種と、全滅時の番兵／infra error への写像だけ。
   3. **Gateway（設定）**: model 列、fallback、retry、analytics / logs。vendor 固有の Interactions URL・callGap・SynthesizeBudget・重装 retry 表は持たない。
2. 呼び出しは **Cloudflare AI Gateway** 経由にする。hosting / R2 / TextWriter と同じ Cloudflare アカウントに揃える。
3. 既定の無料寄せ先は **Workers AI の TTS（Neurons 無料枠）** とする。model へのこだわりは置かない。具体 model id は Adapter 定数または Gateway 設定を正とし、本 Decision に写さない。
4. **Unified Billing / カタログ上の Google TTS 従量（audio token 課金）は、無料運用の既定にしない**。観測・課金一本化の予備経路として Gateway に載せてよいが、日次 produce の主経路は Workers AI Neurons 側にする。
5. Application 側の **セグメント列ループ・部分成功の保持**は残す。これは Domain/UseCase の進行であり、Gateway の model fallback（1 呼び出しの model 列）とは別軸である。
6. 本 Decision は、Google 直叩きの TTS Adapter を主とする先行判断のうち、**transport / retry / model 追従を自前に置く部分**を置き換える。

## 2. Reason

### なぜ TextWriter と同じ境界か

TTS の Domain は「セグメント texts → 音声バイト列（WAV 等）と尺」であり、vendor の Interactions URL・callGap・SynthesizeBudget ではない。Gateway 経由で audio を返す model へ薄い Adapter を向けば、「Application = 検証と進行、Gateway = モデル運用」が一貫する。TextWriter だけ寄せて TTS を Google 直のまま残すと、枠監視と Adapter 肉が二系統のまま残る。

### なぜ Workers AI Neurons を既定の無料経路にするか

model へのこだわりが無く、日次 10〜15 分程度の尺を Neurons 無料枠で賄う計算上の余裕がある。Google BYOK は品質維持の予備になり得るが、RPM/RPD は Google 側に残り、Gateway は枠を消さない。Unified Billing の Google TTS は audio token 従量があり「無料運用の既定」ではない。日本語ポッドキャスト品質は未実測であり、品質不合格なら別判断で差し替える（差し替え先の Reason は本 Decision を更新する）。

### なぜセグメント進行を Application に残すか

Gateway の fallback は「1 呼び出しの model 列」であり、episode 内のセグメント進行・部分成功の蓄積ではない。product 固有の進行を Gateway 設定へ押し込むと、vendor 設定と Domain が混ざる。

## 3. Rejected

1. **TTS を Google 直叩きのまま残し、TextWriter だけ Gateway に寄せる案** — 枠監視と Adapter 肉が二系統残り、寄せる目的が半減する。
2. **Unified Billing の Google TTS を無料運用の主経路にする案** — 公式に audio token 従量がある。
3. **Google AI Studio BYOK を無料運用の既定にする案** — model 非固定・Neurons 既定の方針と食い違う。BYOK は品質予備として残し得るが既定にはしない。
4. **Cloudflare Agents SDK（Durable Object / Think）ごと載せる案** — 日次の texts→audio には過剰。Gateway の HTTP + fallback で足りる。
5. **「尺 10〜15 分なら必ず無料枠内」と保証する案** — Neurons 側は計算上余裕があるが、品質・日本語・声の要件は未実測。保証はしない。
