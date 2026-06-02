import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { Manuscript } from '@daily-it-podcast/core';

const mockGenerateContent = vi.fn();

vi.mock('@google/genai', () => ({
  GoogleGenAI: vi.fn().mockImplementation(() => ({
    models: {
      generateContent: mockGenerateContent,
    },
  })),
  Modality: { AUDIO: 'AUDIO' },
}));

const sampleManuscript: Manuscript = {
  timestamp: '2026-01-01T00:00:00.000Z',
  body: {
    opening: 'オープニング',
    topics: [{ title: 'テーマ', script: '原稿テキスト', durationEstimateSec: 60 }],
    closing: 'クロージング',
  },
};

describe('GeminiTtsService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    process.env['GEMINI_API_KEY'] = 'test-api-key';
  });

  it('Given 有効な Manuscript When synthesize() 実行 Then Buffer が返る', async () => {
    const fakeAudioBase64 = Buffer.from('fake-audio-data').toString('base64');
    mockGenerateContent.mockResolvedValueOnce({
      candidates: [
        {
          content: {
            parts: [
              {
                inlineData: {
                  data: fakeAudioBase64,
                  mimeType: 'audio/wav',
                },
              },
            ],
          },
        },
      ],
    });

    const { GeminiTtsService } = await import('../gemini-tts.js');
    const tts = new GeminiTtsService();
    const result = await tts.synthesize(sampleManuscript);

    expect(Buffer.isBuffer(result)).toBe(true);
    expect(result.length).toBeGreaterThan(0);
  });

  it('Given API KEY なし When インスタンス生成 Then TtsError が throw される', async () => {
    delete process.env['GEMINI_API_KEY'];
    const { GeminiTtsService } = await import('../gemini-tts.js');
    const { TtsError } = await import('@daily-it-podcast/core');

    expect(() => new GeminiTtsService()).toThrow(TtsError);
  });

  it('Given API レスポンスに audio データなし When synthesize() 実行 Then TtsError が throw される', async () => {
    mockGenerateContent.mockResolvedValueOnce({
      candidates: [{ content: { parts: [{ text: 'テキストのみ' }] } }],
    });

    const { GeminiTtsService } = await import('../gemini-tts.js');
    const { TtsError } = await import('@daily-it-podcast/core');
    const tts = new GeminiTtsService();

    await expect(tts.synthesize(sampleManuscript)).rejects.toThrow(TtsError);
  });

  it('Given 複数topic When synthesize() 実行 Then generateContent が1回呼ばれる', async () => {
    const fakeAudioBase64 = Buffer.from('audio').toString('base64');
    mockGenerateContent.mockResolvedValueOnce({
      candidates: [
        {
          content: {
            parts: [{ inlineData: { data: fakeAudioBase64, mimeType: 'audio/wav' } }],
          },
        },
      ],
    });

    const manuscriptMultiTopics: Manuscript = {
      ...sampleManuscript,
      body: {
        ...sampleManuscript.body,
        topics: [
          { title: 'テーマ1', script: 'スクリプト1', durationEstimateSec: 30 },
          { title: 'テーマ2', script: 'スクリプト2', durationEstimateSec: 30 },
        ],
      },
    };

    const { GeminiTtsService } = await import('../gemini-tts.js');
    const tts = new GeminiTtsService();
    await tts.synthesize(manuscriptMultiTopics);

    expect(mockGenerateContent).toHaveBeenCalledTimes(1);
  });
});
