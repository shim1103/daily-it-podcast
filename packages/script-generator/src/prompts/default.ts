import type { ThemeInfo } from '@daily-it-podcast/core';

/**
 * Gemini に渡す原稿生成 prompt を構築する。
 * 重要度・難しさを LLM 自身に判断させて durationEstimateSec を 60-180 秒で決定する。
 * intro/outro フレーズも含めた完結した原稿を JSON で返させる。
 */
export function buildPrompt(info: ThemeInfo): string {
  return `
あなたは日本語の IT ニュース podcast の原稿ライターです。
以下の IT ニュース情報をもとに、ですます体で自然な podcast 原稿を書いてください。

## ニュース情報
タイトル: ${info.title}
内容: ${info.rawText}

## 出力ルール
- 出力は JSON のみ。説明文・コードブロック記号は不要
- JSON フィールド: { "script": string, "durationEstimateSec": number }
- script: 以下を含む完結した原稿
  - 導入フレーズ（例: 「続いては〜についてお伝えします。」）
  - 内容説明（ニュースの背景・意義・ポイントを分かりやすく）
  - 締めフレーズ（例: 「以上、〜についてでした。」）
- durationEstimateSec: ニュースの重要度と難しさを考慮して 60〜180 の整数を決める
  - 重要度高・難しい → 180 秒寄り
  - 軽めのニュース → 60 秒寄り
- 英語のタイトルや固有名詞はカタカナ読みで読み上げやすくすること
`.trim();
}
