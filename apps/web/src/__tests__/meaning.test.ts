import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

const FALLBACK = '説明を取得できませんでした。';

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

afterEach(() => {
  delete process.env['GEMINI_API_KEY'];
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

  it('Given GEMINI_API_KEY が未設定 When fetchMeaning() Then fallback が返る', async () => {
    delete process.env['GEMINI_API_KEY'];

    const { fetchMeaning } = await import('../lib/meaning.js');
    const result = await fetchMeaning('TypeScript');

    expect(result).toBe(FALLBACK);
  });

  it('Given generateContent が例外を throw When fetchMeaning() Then fallback が返る', async () => {
    const { GoogleGenAI } = await import('@google/genai');
    vi.mocked(GoogleGenAI).mockImplementation(() => ({
      models: {
        generateContent: vi.fn().mockRejectedValue(new Error('API error')),
      },
    }) as never);

    const { fetchMeaning } = await import('../lib/meaning.js');
    const result = await fetchMeaning('TypeScript');

    expect(result).toBe(FALLBACK);
  });

  it('Given text が空文字列 When fetchMeaning() Then fallback が返る', async () => {
    const { GoogleGenAI } = await import('@google/genai');
    vi.mocked(GoogleGenAI).mockImplementation(() => ({
      models: {
        generateContent: vi.fn().mockResolvedValue({
          candidates: [{ content: { parts: [{ text: '' }] } }],
        }),
      },
    }) as never);

    const { fetchMeaning } = await import('../lib/meaning.js');
    const result = await fetchMeaning('TypeScript');

    expect(result).toBe(FALLBACK);
  });

  it('Given text が空白のみ When fetchMeaning() Then fallback が返る', async () => {
    const { GoogleGenAI } = await import('@google/genai');
    vi.mocked(GoogleGenAI).mockImplementation(() => ({
      models: {
        generateContent: vi.fn().mockResolvedValue({
          candidates: [{ content: { parts: [{ text: '   ' }] } }],
        }),
      },
    }) as never);

    const { fetchMeaning } = await import('../lib/meaning.js');
    const result = await fetchMeaning('TypeScript');

    expect(result).toBe(FALLBACK);
  });
});
