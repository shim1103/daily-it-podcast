import { describe, it, expect, vi, beforeEach } from 'vitest';
import { HackerNewsInfoFetcher } from '../hackernews.js';

// HN API をネットワーク呼び出しなしに検証するため fetch をモック
const mockFetch = vi.fn();
vi.stubGlobal('fetch', mockFetch);

const TOP_STORIES_URL = 'https://hacker-news.firebaseio.com/v0/topstories.json';
const ITEM_URL = (id: number) => `https://hacker-news.firebaseio.com/v0/item/${id}.json`;

function makeStory(id: number) {
  return {
    id,
    title: `Story ${id}`,
    url: `https://example.com/${id}`,
    text: `本文 ${id}`,
    score: 100 + id,
    time: Math.floor(Date.now() / 1000),
    type: 'story',
  };
}

beforeEach(() => {
  mockFetch.mockReset();
});

describe('HackerNewsInfoFetcher', () => {
  it('Given top stories When fetch() Then ThemeInfo[] が maxItems 件返る', async () => {
    const ids = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
    mockFetch
      .mockResolvedValueOnce({ ok: true, json: async () => ids })
      .mockImplementation(async (url: string) => {
        const id = Number(url.split('/item/')[1]?.replace('.json', ''));
        return { ok: true, json: async () => makeStory(id) };
      });

    const fetcher = new HackerNewsInfoFetcher({ maxItems: 5 });
    const result = await fetcher.fetch();

    expect(result).toHaveLength(5);
    for (const item of result) {
      expect(item.source).toBe('hackernews');
      expect(typeof item.title).toBe('string');
      expect(typeof item.rawText).toBe('string');
      expect(typeof item.fetchedAt).toBe('string');
      expect(() => new Date(item.fetchedAt)).not.toThrow();
    }
  });

  it('Given top stories API fail When fetch() Then error が throw される', async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, status: 500 });

    const fetcher = new HackerNewsInfoFetcher({ maxItems: 5 });
    await expect(fetcher.fetch()).rejects.toThrow();
  });

  it('Given story without url (Ask HN) When fetch() Then rawText は text フィールドを使う', async () => {
    const ids = [42];
    const askStory = { id: 42, title: 'Ask HN: test', text: 'ask text', score: 50, time: 1, type: 'story' };
    mockFetch
      .mockResolvedValueOnce({ ok: true, json: async () => ids })
      .mockResolvedValueOnce({ ok: true, json: async () => askStory });

    const fetcher = new HackerNewsInfoFetcher({ maxItems: 1 });
    const result = await fetcher.fetch();

    expect(result[0]?.rawText).toContain('ask text');
  });
});
