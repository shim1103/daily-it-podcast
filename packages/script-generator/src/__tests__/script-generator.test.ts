import { describe, it, expect } from 'vitest';
import { MockScriptGenerator } from '../mock.js';
import type { ThemeInfo } from '@daily-it-podcast/core';

const sampleInfo: ThemeInfo = {
  source: 'manual-text',
  title: 'TypeScript 5.8 リリース',
  rawText: 'TypeScript 5.8 が正式リリースされた。',
  fetchedAt: new Date().toISOString(),
};

describe('MockScriptGenerator', () => {
  it('Given 有効な ThemeInfo When generate() 実行 Then ThemeScript が返る', async () => {
    const generator = new MockScriptGenerator();
    const result = await generator.generate(sampleInfo);

    expect(typeof result.title).toBe('string');
    expect(typeof result.script).toBe('string');
    expect(typeof result.durationEstimateSec).toBe('number');
    expect(result.durationEstimateSec).toBeGreaterThan(0);
    expect(result.title).toBe(sampleInfo.title);
  });
});
