import { describe, it, expect, vi, beforeEach } from 'vitest';

// Gemini SDK をモックして実際の API 呼び出しを回避
vi.mock('@google/genai', () => ({
  GoogleGenAI: vi.fn().mockImplementation(() => ({
    models: {
      generateContent: vi.fn(),
    },
  })),
}));

beforeEach(() => {
  vi.resetModules();
  process.env['GEMINI_API_KEY'] = 'test-key';
});

describe('fetchMeaning', () => {
  it('Given 有効な word When fetchMeaning() Then 説明文字列が返る', async () => {
    const { GoogleGenAI } = await import('@google/genai');
    vi.mocked(GoogleGenAI).mockImplementation(() => ({
      models: {
        generateContent: vi.fn().mockResolvedValue({
          candidates: [{ content: { parts: [{ text: 'TypeScript は型安全な言語です。' }] } }],
        }),
      },
    }) as never);

    const { fetchMeaning } = await import('../lib/meaning.js');
    const result = await fetchMeaning('TypeScript');

    expect(typeof result).toBe('string');
    expect(result.length).toBeGreaterThan(0);
  });

  it('Given API が空レスポンス When fetchMeaning() Then fallback 文字列が返る', async () => {
    const { GoogleGenAI } = await import('@google/genai');
    vi.mocked(GoogleGenAI).mockImplementation(() => ({
      models: {
        generateContent: vi.fn().mockResolvedValue({ candidates: [] }),
      },
    }) as never);

    const { fetchMeaning } = await import('../lib/meaning.js');
    const result = await fetchMeaning('unknown');

    expect(typeof result).toBe('string');
  });
});
