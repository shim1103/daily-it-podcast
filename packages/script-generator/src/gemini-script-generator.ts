import { GoogleGenAI } from '@google/genai';
import type { ScriptGenerator, ThemeInfo, ThemeScript } from '@daily-it-podcast/core';
import { ScriptGenerationError } from '@daily-it-podcast/core';
import { buildPrompt } from './prompts/default.js';

interface GeminiScriptResponse {
  script: string;
  durationEstimateSec: number;
}

function parseResponse(text: string): GeminiScriptResponse {
  // Gemini がコードブロックで包んで返す場合に対応
  const cleaned = text.replace(/^```(?:json)?\n?/, '').replace(/\n?```$/, '').trim();
  let parsed: unknown;
  try {
    parsed = JSON.parse(cleaned);
  } catch {
    throw new ScriptGenerationError(`Gemini の応答が JSON でない: ${text.slice(0, 100)}`);
  }

  if (
    typeof parsed !== 'object' ||
    parsed === null ||
    typeof (parsed as Record<string, unknown>)['script'] !== 'string' ||
    typeof (parsed as Record<string, unknown>)['durationEstimateSec'] !== 'number'
  ) {
    throw new ScriptGenerationError(`Gemini の応答スキーマ不正: ${cleaned.slice(0, 100)}`);
  }

  return parsed as GeminiScriptResponse;
}

export class GeminiScriptGenerator implements ScriptGenerator {
  private readonly ai: GoogleGenAI;
  private readonly model = 'gemini-2.0-flash';

  constructor() {
    const apiKey = process.env['GEMINI_API_KEY'];
    if (!apiKey) {
      throw new ScriptGenerationError('GEMINI_API_KEY が設定されていません');
    }
    this.ai = new GoogleGenAI({ apiKey });
  }

  async generate(info: ThemeInfo): Promise<ThemeScript> {
    const prompt = buildPrompt(info);

    let response;
    try {
      response = await this.ai.models.generateContent({
        model: this.model,
        contents: [{ parts: [{ text: prompt }] }],
      });
    } catch (err) {
      throw new ScriptGenerationError('Gemini API 呼び出しに失敗しました', err);
    }

    const text = response.candidates?.[0]?.content?.parts?.[0]?.text;
    if (!text) {
      throw new ScriptGenerationError('Gemini API からのテキストが空です');
    }

    const { script, durationEstimateSec } = parseResponse(text);

    return {
      title: info.title,
      script,
      durationEstimateSec,
    };
  }
}
