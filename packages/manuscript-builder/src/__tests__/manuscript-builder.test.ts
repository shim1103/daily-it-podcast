import { describe, it, expect, vi } from 'vitest';
import { DefaultManuscriptBuilder } from '../builder.js';
import type { InfoFetcher, ScriptGenerator, ThemeInfo, ThemeScript } from '@daily-it-podcast/core';

const mockFetcher: InfoFetcher = {
  fetch: vi.fn().mockResolvedValue([
    {
      source: 'manual-text',
      title: 'テストテーマ',
      rawText: 'テスト本文',
      fetchedAt: '2026-01-01T00:00:00.000Z',
    } satisfies ThemeInfo,
  ]),
};

const mockGenerator: ScriptGenerator = {
  generate: vi.fn().mockResolvedValue({
    title: 'テストテーマ',
    script: 'テスト原稿テキスト',
    durationEstimateSec: 60,
  } satisfies ThemeScript),
};

describe('DefaultManuscriptBuilder', () => {
  it('Given InfoFetcherモック + ScriptGeneratorモック When build() 実行 Then Manuscript が返り各フィールドが揃う', async () => {
    const builder = new DefaultManuscriptBuilder(mockFetcher, mockGenerator);
    const result = await builder.build();

    expect(typeof result.timestamp).toBe('string');
    expect(() => new Date(result.timestamp)).not.toThrow();
    expect(typeof result.body.opening).toBe('string');
    expect(Array.isArray(result.body.topics)).toBe(true);
    expect(result.body.topics.length).toBe(1);
    expect(typeof result.body.closing).toBe('string');
  });

  it('Given 複数ThemeInfo When build() 実行 Then topics数がThemeInfo数と一致する', async () => {
    const multiFetcher: InfoFetcher = {
      fetch: vi.fn().mockResolvedValue([
        { source: 's', title: 'A', rawText: 'a', fetchedAt: '2026-01-01T00:00:00Z' },
        { source: 's', title: 'B', rawText: 'b', fetchedAt: '2026-01-01T00:00:00Z' },
      ]),
    };
    const builder = new DefaultManuscriptBuilder(multiFetcher, mockGenerator);
    const result = await builder.build();

    expect(result.body.topics.length).toBe(2);
  });
});
