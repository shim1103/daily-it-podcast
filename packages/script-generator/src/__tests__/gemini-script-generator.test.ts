import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ThemeInfo } from '@daily-it-podcast/core';

// @google/genai をモックして実際の API 呼び出しを回避
vi.mock('@google/genai', () => ({
  GoogleGenAI: vi.fn().mockImplementation(() => ({
    models: {
      generateContent: vi.fn(),
    },
  })),
}));

const sampleInfo: ThemeInfo = {
  source: 'hackernews',
  title: 'TypeScript 6.0 Released',
  rawText: 'TypeScript 6.0 has been released with major improvements.',
  fetchedAt: new Date().toISOString(),
};

const validResponse = {
  candidates: [
    {
      content: {
        parts: [
          {
            text: JSON.stringify({
              script: 'TypeScriptの新バージョンについてお伝えします。大きな改善が加えられました。以上でした。',
              durationEstimateSec: 90,
            }),
          },
        ],
      },
    },
  ],
};

beforeEach(() => {
  vi.resetModules();
  process.env['GEMINI_API_KEY'] = 'test-key';
});

describe('GeminiScriptGenerator', () => {
  it('Given 有効な ThemeInfo When generate() Then ThemeScript が返る', async () => {
    const { GoogleGenAI } = await import('@google/genai');
    const mockGenerateContent = vi.fn().mockResolvedValue(validResponse);
    vi.mocked(GoogleGenAI).mockImplementation(() => ({
      models: { generateContent: mockGenerateContent },
    }) as never);

    const { GeminiScriptGenerator } = await import('../gemini-script-generator.js');
    const generator = new GeminiScriptGenerator();
    const result = await generator.generate(sampleInfo);

    expect(result.title).toBe(sampleInfo.title);
    expect(typeof result.script).toBe('string');
    expect(result.script.length).toBeGreaterThan(0);
    expect(typeof result.durationEstimateSec).toBe('number');
    expect(result.durationEstimateSec).toBeGreaterThan(0);
  });

  it('Given GEMINI_API_KEY 未設定 When new GeminiScriptGenerator() Then ScriptGenerationError が throw される', async () => {
    delete process.env['GEMINI_API_KEY'];
    const { GeminiScriptGenerator } = await import('../gemini-script-generator.js');
    expect(() => new GeminiScriptGenerator()).toThrow('GEMINI_API_KEY');
  });

  it('Given API が不正な JSON を返す When generate() Then error が throw される', async () => {
    const { GoogleGenAI } = await import('@google/genai');
    const badResponse = {
      candidates: [{ content: { parts: [{ text: 'not-json' }] } }],
    };
    const mockGenerateContent = vi.fn().mockResolvedValue(badResponse);
    vi.mocked(GoogleGenAI).mockImplementation(() => ({
      models: { generateContent: mockGenerateContent },
    }) as never);

    const { GeminiScriptGenerator } = await import('../gemini-script-generator.js');
    const generator = new GeminiScriptGenerator();
    await expect(generator.generate(sampleInfo)).rejects.toThrow();
  });
});
