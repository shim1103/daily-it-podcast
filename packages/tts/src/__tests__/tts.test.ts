import { describe, it, expect } from 'vitest';
import { MockTtsService } from '../mock.js';
import type { Manuscript } from '@daily-it-podcast/core';

const sampleManuscript: Manuscript = {
  timestamp: '2026-01-01T00:00:00.000Z',
  body: {
    opening: 'オープニング',
    topics: [{ title: 'テーマ', script: '原稿', durationEstimateSec: 60 }],
    closing: 'クロージング',
  },
};

describe('MockTtsService', () => {
  it('Given 有効な Manuscript When synthesize() 実行 Then Buffer が返る', async () => {
    const tts = new MockTtsService();
    const result = await tts.synthesize(sampleManuscript);

    expect(Buffer.isBuffer(result)).toBe(true);
  });
});
