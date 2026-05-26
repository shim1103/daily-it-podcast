import { describe, it, expect } from 'vitest';
import { MockInfoFetcher } from '../mock.js';

describe('MockInfoFetcher', () => {
  it('Given 有効な設定 When fetch() 実行 Then ThemeInfo[] が返る', async () => {
    const fetcher = new MockInfoFetcher();
    const result = await fetcher.fetch();

    expect(Array.isArray(result)).toBe(true);
    expect(result.length).toBeGreaterThan(0);

    for (const item of result) {
      expect(typeof item.source).toBe('string');
      expect(typeof item.title).toBe('string');
      expect(typeof item.rawText).toBe('string');
      expect(typeof item.fetchedAt).toBe('string');
      expect(() => new Date(item.fetchedAt)).not.toThrow();
    }
  });
});
