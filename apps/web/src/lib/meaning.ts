import { GoogleGenAI } from '@google/genai';

// server-side でのみ呼ぶ（Route Handler / Server Component 専用）
const FALLBACK = '説明を取得できませんでした。';

function buildPrompt(word: string): string {
  return `IT用語「${word}」を日本語で1〜2文で分かりやすく説明してください。専門知識がない人にも伝わるように書いてください。説明文のみ出力し、余計な前置きは不要です。`;
}

export async function fetchMeaning(word: string): Promise<string> {
  const apiKey = process.env['GEMINI_API_KEY'];
  if (!apiKey) {
    return FALLBACK;
  }

  const ai = new GoogleGenAI({ apiKey });
  let response;
  try {
    response = await ai.models.generateContent({
      model: 'gemini-2.0-flash',
      contents: [{ parts: [{ text: buildPrompt(word) }] }],
    });
  } catch {
    return FALLBACK;
  }

  const text = response.candidates?.[0]?.content?.parts?.[0]?.text;
  return text?.trim() || FALLBACK;
}
