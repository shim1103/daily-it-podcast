---
name: Gemini Interactions API の audio 応答は steps[].content[].data から取る
date: 2026-09-02T18:01:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. Gemini Interactions API（`POST /v1beta/interactions`、`response_format: audio`）の成功応答から audio base64 を取り出す場所は **`steps[].content[].data`** とする。`output_audio.data` は読まない。
2. `steps` を走査し、空を飛ばして最初に見つかった `content[].data` を採用する。`steps[0].content[0]` 決め打ちにしない。
3. どの step / content にも非空 data が無いときだけ「audio 欠落」と判定し、切り分けのため body のトップレベルキー一覧を error へ添える（`topLevelKeysHint`）。

## 2. Reason

1. System の TTS が `decode_pcm: output audio is missing` で決定論的に落ち続けていた（run 33581258235 / 33582558173 / 33607924229）。当初「Gemini preview の Limitation で応答が時々 audio 無しで返る」「無料枠 日次 quota」と記録していたが、いずれも誤診だった。
2. 診断 snippet を 4000 byte へ広げ、トップレベルキー一覧を error へ添えたところ（run 33609034783）、実応答が次の形と判明した。
   ```json
   { "id": "...", "object": "...", "model": "...", "status": "completed",
     "service_tier": "standard", "usage": { ... },
     "steps": [ { "content": [ { "data": "<base64 PCM 16bit 24kHz mono>" } ] } ] }
   ```
   audio は `steps[].content[].data` にあり、`output_audio` フィールドは存在しない。旧構造を読んでいたため毎セグメントで decode が失敗し、retryable 扱いで retry と日次 quota を浪費していた。
3. `output_audio.data` は旧 `generateContent` TTS か旧 preview の形。現行 Interactions API は step / content の入れ子で返す。将来 SDK 例が `interaction.output_audio` を示していても、生 HTTP の実 body はこの形。
4. `steps[0].content[0]` 決め打ちにしないのは、複数 step / 複数 content で返る余地に備えるため。最初の非空 data で足りる。

## 3. Rejected

1. `output_audio.data` を読み続け、audio 欠落を「一過性」として retry で吸収する案 — 決定論的失敗なので retry では直らない。日次 quota を浪費するだけ。
2. audio の場所を SDK ドキュメントの `interaction.output_audio` に合わせる案 — 実 HTTP body と一致しない。実測した body 構造を正とする。
3. `steps[0].content[0].data` 決め打ち案 — 複数 step / content の応答に脆い。走査コストは無視できる。
